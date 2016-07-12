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
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/google/git-appraise/repository"
	"github.com/google/git-appraise/review"
)

const (
	// SHA1 produces 160 bit hashes, so a hex-encoded hash should be no more than 40 characters.
	maxHashLength = 40
)

// RepoCache encapsulates everything that the API server currently knows about every repository.
type RepoCache map[string]*RepoDetails

// AddRepo adds the given repository to the cache.
func (cache RepoCache) AddRepo(repo repository.Repo) {
	repoDetails := NewRepoDetails(repo)
	cache[repoDetails.ID] = repoDetails
}

func checkStringLooksLikeHash(s string) error {
	if len(s) > maxHashLength {
		return errors.New("Invalid hash parameter")
	}
	for _, c := range s {
		if ((c < 'a') || (c > 'f')) && ((c < '0') || (c > '9')) {
			return errors.New("Invalid hash character")
		}
	}
	return nil
}

func (cache RepoCache) getRepoDetails(r *http.Request) (*RepoDetails, error) {
	repoParam := r.URL.Query().Get("repo")
	if repoParam == "" {
		return nil, errors.New("No repository specified")
	}
	if err := checkStringLooksLikeHash(repoParam); err != nil {
		return nil, err
	}
	repoDetails, ok := cache[repoParam]
	if !ok {
		return nil, errors.New("Invalid repository specified")
	}
	return repoDetails, nil
}

func (cache RepoCache) getReview(r *http.Request) (*review.Review, error) {
	repoDetails, err := cache.getRepoDetails(r)
	if err != nil {
		return nil, err
	}
	reviewParam := r.URL.Query().Get("review")
	if reviewParam == "" {
		return nil, errors.New("No review specified")
	}
	if err := checkStringLooksLikeHash(reviewParam); err != nil {
		return nil, err
	}
	reviewDetails, err := repoDetails.GetReview(reviewParam)
	if err != nil {
		return nil, errors.New("Invalid review specified")
	}
	return reviewDetails, nil
}

func serveJSON(v interface{}, w http.ResponseWriter) {
	json, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

// ServeListReposJSON writes the list of repositories to the given writer.
func (cache RepoCache) ServeListReposJSON(w http.ResponseWriter, r *http.Request) {
	var reposList ReposList
	for _, repoDetails := range cache {
		reposList = append(reposList, repoDetails.GetListItem())
	}
	sort.Stable(reposList)
	serveJSON(reposList, w)
}

// ServeRepoSummaryJSON writes the summary of a given repository to the given writer.
//
// The repository to summarize is given by the 'repo' URL parameter.
func (cache RepoCache) ServeRepoSummaryJSON(w http.ResponseWriter, r *http.Request) {
	repoDetails, err := cache.getRepoDetails(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	summary, err := repoDetails.GetSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveJSON(summary, w)
}

// ServeRepoContents writes the contents of a given file at a given commit.
//
// The repository, file, and commit are given by the 'repo', 'file' and 'commit'
// URL parameters.
func (cache RepoCache) ServeRepoContents(w http.ResponseWriter, r *http.Request) {
	repoDetails, err := cache.getRepoDetails(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	commitParam := r.URL.Query().Get("commit")
	if commitParam == "" {
		http.Error(w, "No commit specified", http.StatusBadRequest)
		return
	}
	if err := checkStringLooksLikeHash(commitParam); err != nil {
		http.Error(w, "Invalid commit specified", http.StatusBadRequest)
		return
	}
	fileParam := r.URL.Query().Get("file")
	if fileParam == "" {
		http.Error(w, "No file specified", http.StatusBadRequest)
		return
	}

	contents, err := repoDetails.Repo.Show(commitParam, fileParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte(contents))
}

func getPageToken(r *http.Request) (page int, err error) {
	pageParam := r.URL.Query().Get("page")
	if pageParam != "" {
		page, err = strconv.Atoi(pageParam)
		if err != nil {
			return 0, err
		}
		if page < 0 {
			return 0, errors.New("Invalid page token")
		}
	}
	return page, nil
}

// ServeClosedReviewsJSON writes a page of the closed reviews list for the given repository to the given writer.
//
// The repository to list reviews for is given by the 'repo' URL parameter.
// The page of the review list to output is given by the 'page' URL parameter.
func (cache RepoCache) ServeClosedReviewsJSON(w http.ResponseWriter, r *http.Request) {
	repoDetails, err := cache.getRepoDetails(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pageToken, err := getPageToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	closedReviews, err := repoDetails.GetClosedReviews(pageToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveJSON(closedReviews, w)
}

// ServeOpenReviewsJSON writes a page of the open reviews list for the given repository to the given writer.
//
// The repository to list reviews for is given by the 'repo' URL parameter.
// The page of the review list to output is given by the 'page' URL parameter.
func (cache RepoCache) ServeOpenReviewsJSON(w http.ResponseWriter, r *http.Request) {
	repoDetails, err := cache.getRepoDetails(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pageToken, err := getPageToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	openReviews, err := repoDetails.GetOpenReviews(pageToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveJSON(openReviews, w)
}

// ServeReviewDetailsJSON writes the details of a review to the given writer.
//
// The enclosing repository is given by the 'repo' URL parameter.
// The review to write is given by the 'review' URL parameter.
func (cache RepoCache) ServeReviewDetailsJSON(w http.ResponseWriter, r *http.Request) {
	reviewDetails, err := cache.getReview(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	serveJSON(reviewDetails, w)
}

// ServeReviewDiff writes the diff summary of a review to the given writer.
//
// The enclosing repository is given by the 'repo' URL parameter.
// The review to write is given by the 'review' URL parameter.
func (cache RepoCache) ServeReviewDiff(w http.ResponseWriter, r *http.Request) {
	reviewDetails, err := cache.getReview(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	lhs := r.URL.Query().Get("lhs")
	rhs := r.URL.Query().Get("rhs")
	if err := checkStringLooksLikeHash(lhs); err != nil {
		http.Error(w, "Invalid left-hand-side commit specified", http.StatusBadRequest)
		return
	}
	if err := checkStringLooksLikeHash(rhs); err != nil {
		http.Error(w, "Invalid right-hand-side commit specified", http.StatusBadRequest)
		return
	}
	diffSummary, err := NewDiffSummary(reviewDetails, lhs, rhs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveJSON(diffSummary, w)
}

// ServeEntryPointRedirect writes the main redirect response to the given writer.
func (cache RepoCache) ServeEntryPointRedirect(w http.ResponseWriter, r *http.Request) {
	if len(cache) == 1 {
		for id := range cache {
			http.Redirect(w, r, "/static/reviews.html#?repo="+id, http.StatusTemporaryRedirect)
			return
		}
	}
	http.Redirect(w, r, "/static/repos.html", http.StatusTemporaryRedirect)
	return
}
