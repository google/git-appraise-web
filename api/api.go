/*
Copyright 2016 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/git-appraise/repository"
	"github.com/google/git-appraise/review"
)

type RepoListItem struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}
type ReposList []RepoListItem

type RepoCacheItem struct {
	Repo          repository.Repo
	RepoState     string
	OpenReviews   []review.Summary
	ClosedReviews []review.Summary
}

func (r *RepoCacheItem) Update() error {
	stateHash, err := r.Repo.GetRepoStateHash()
	if err != nil {
		return err
	}
	if stateHash == r.RepoState {
		return nil
	}
	allReviews := review.ListAll(r.Repo)
	var openReviews []review.Summary
	var closedReviews []review.Summary
	for _, review := range allReviews {
		if review.Submitted {
			closedReviews = append(closedReviews, review)
		} else {
			openReviews = append(openReviews, review)
		}
	}
	r.OpenReviews = openReviews
	r.ClosedReviews = closedReviews
	r.RepoState = stateHash
	return nil
}

type RepoCache map[string]*RepoCacheItem

// Get a fixed-length, obfuscated ID for the given repo.
func getRepoId(repo repository.Repo) string {
	return fmt.Sprintf("%.6x", sha1.Sum([]byte(repo.GetPath())))
}

func (cache RepoCache) AddRepo(repo repository.Repo) {
	cache[getRepoId(repo)] = &RepoCacheItem{
		Repo: repo,
	}
}

func (cache RepoCache) GetRepoCacheItem(r *http.Request) (*RepoCacheItem, error) {
	repoParam := r.URL.Query().Get("repo")
	if repoParam == "" {
		return nil, errors.New("No repository specified")
	}
	repoCacheItem, ok := cache[repoParam]
	if !ok {
		return nil, errors.New("Invalid repository specified")
	}
	if err := repoCacheItem.Update(); err != nil {
		return nil, err
	}
	return repoCacheItem, nil
}

func (cache RepoCache) GetReview(r *http.Request) (*review.Review, error) {
	repoCacheItem, err := cache.GetRepoCacheItem(r)
	if err != nil {
		return nil, err
	}
	reviewParam := r.URL.Query().Get("review")
	if reviewParam == "" {
		return nil, errors.New("No review specified")
	}
	reviewDetails, err := review.Get(repoCacheItem.Repo, reviewParam)
	if err != nil {
		return nil, errors.New("Invalid review specified")
	}
	return reviewDetails, nil
}

func serveJson(v interface{}, w http.ResponseWriter) {
	json, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

func (cache RepoCache) ServeListReposJson(w http.ResponseWriter, r *http.Request) {
	var reposList ReposList
	for id, cacheItem := range cache {
		reposList = append(reposList, RepoListItem{
			ID:   id,
			Path: cacheItem.Repo.GetPath(),
		})
	}
	serveJson(reposList, w)
}

type RepoSummary struct {
	Path              string `json:"path"`
	OpenReviewCount   int    `json:"openReviewCount"`
	ClosedReviewCount int    `json:"closedReviewCount"`
}

func (cache RepoCache) ServeRepoSummaryJson(w http.ResponseWriter, r *http.Request) {
	repoCacheItem, err := cache.GetRepoCacheItem(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	summary := RepoSummary{
		Path:              repoCacheItem.Repo.GetPath(),
		OpenReviewCount:   len(repoCacheItem.OpenReviews),
		ClosedReviewCount: len(repoCacheItem.ClosedReviews),
	}
	serveJson(summary, w)
}

func (cache RepoCache) ServeClosedReviewsJson(w http.ResponseWriter, r *http.Request) {
	repoCacheItem, err := cache.GetRepoCacheItem(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	serveJson(repoCacheItem.ClosedReviews, w)
}

func (cache RepoCache) ServeOpenReviewsJson(w http.ResponseWriter, r *http.Request) {
	repoCacheItem, err := cache.GetRepoCacheItem(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	serveJson(repoCacheItem.OpenReviews, w)
}

func (cache RepoCache) ServeReviewDetailsJson(w http.ResponseWriter, r *http.Request) {
	reviewDetails, err := cache.GetReview(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	serveJson(reviewDetails, w)
}

func (cache RepoCache) ServeReviewDiff(w http.ResponseWriter, r *http.Request) {
	reviewDetails, err := cache.GetReview(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	diff, err := reviewDetails.GetDiff()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(diff))
}

func (cache RepoCache) ServeEntryPointRedirect(w http.ResponseWriter, r *http.Request) {
	if len(cache) == 1 {
		for id := range cache {
			http.Redirect(w, r, "/static/reviews.html?repo="+id, http.StatusTemporaryRedirect)
			return
		}
	}
	http.Redirect(w, r, "/static/repos.html", http.StatusTemporaryRedirect)
	return
}
