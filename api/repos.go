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
	"errors"
	"fmt"
	"strconv"

	"github.com/google/git-appraise/repository"
	"github.com/google/git-appraise/review"
)

// RepoListItem represents one entry in the result of calling the API to list repositories.
type RepoListItem struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

// ReposList is the return type for the API to list repositories.
type ReposList []*RepoListItem

func (repos ReposList) Len() int      { return len(repos) }
func (repos ReposList) Swap(i, j int) { repos[i], repos[j] = repos[j], repos[i] }
func (repos ReposList) Less(i, j int) bool {
	return repos[i].ID < repos[j].ID
}

// RepoSummary is the return type for the API to summarize a repository.
type RepoSummary struct {
	Path              string `json:"path"`
	OpenReviewCount   int    `json:"openReviewCount"`
	ClosedReviewCount int    `json:"closedReviewCount"`
}

// RepoDetails encapsulates everything the API server knows about a repository.
type RepoDetails struct {
	ID                string
	Repo              repository.Repo
	RepoState         string
	OpenReviewCount   int
	OpenReviews       [][]review.Summary
	ClosedReviewCount int
	ClosedReviews     [][]review.Summary
}

// Get a fixed-length, obfuscated ID for the given repo.
func getRepoID(repo repository.Repo) string {
	return fmt.Sprintf("%.6x", sha1.Sum([]byte(repo.GetPath())))
}

// NewRepoDetails constructs a RepoDetails instance from the given Repo instance.
func NewRepoDetails(repo repository.Repo) *RepoDetails {
	return &RepoDetails{
		ID:   getRepoID(repo),
		Repo: repo,
	}
}

func (details *RepoDetails) update() error {
	stateHash, err := details.Repo.GetRepoStateHash()
	if err != nil {
		return err
	}
	if stateHash == details.RepoState {
		return nil
	}
	allReviews := review.ListAll(details.Repo)
	var openReviews []review.Summary
	var closedReviews []review.Summary
	for _, review := range allReviews {
		if review.Submitted || review.Request.TargetRef == "" {
			closedReviews = append(closedReviews, review)
		} else {
			openReviews = append(openReviews, review)
		}
	}
	details.OpenReviewCount = len(openReviews)
	details.OpenReviews = paginateReviews(openReviews, 100)
	details.ClosedReviewCount = len(closedReviews)
	details.ClosedReviews = paginateReviews(closedReviews, 100)
	details.RepoState = stateHash
	return nil
}

// GetReview loads the given review details from the repository.
func (details *RepoDetails) GetReview(reviewID string) (*review.Review, error) {
	if err := details.update(); err != nil {
		return nil, err
	}
	reviewDetails, err := review.Get(details.Repo, reviewID)
	if err != nil {
		return nil, errors.New("Invalid review specified")
	}
	return reviewDetails, nil
}

// GetSummary constructs a detailed summary of the repository.
func (details *RepoDetails) GetSummary() (*RepoSummary, error) {
	if err := details.update(); err != nil {
		return nil, err
	}
	return &RepoSummary{
		Path:              details.Repo.GetPath(),
		OpenReviewCount:   details.OpenReviewCount,
		ClosedReviewCount: details.ClosedReviewCount,
	}, nil
}

// GetListItem constructs a concise summary of the repository suitable for including in a list of repositories.
func (details *RepoDetails) GetListItem() *RepoListItem {
	return &RepoListItem{
		ID:   details.ID,
		Path: details.Repo.GetPath(),
	}
}

func getReviewListResponse(pageToken int, reviews [][]review.Summary) *ReviewListResponse {
	var items []review.Summary
	var nextPageToken string
	if pageToken < len(reviews) {
		items = reviews[pageToken]
		if pageToken < len(reviews)-1 {
			nextPageToken = strconv.Itoa(pageToken + 1)
		}
	}
	return &ReviewListResponse{
		Items:         items,
		NextPageToken: nextPageToken,
	}
}

// GetClosedReviews returns the given `page` of the paginated list of closed reviews.
//
// If the page is out of bounds, then an empty response is returned.
func (details *RepoDetails) GetClosedReviews(pageToken int) (*ReviewListResponse, error) {
	if err := details.update(); err != nil {
		return nil, err
	}
	return getReviewListResponse(pageToken, details.ClosedReviews), nil
}

// GetOpenReviews returns the given `page` of the paginated list of open reviews.
//
// If the page is out of bounds, then an empty response is returned.
func (details *RepoDetails) GetOpenReviews(pageToken int) (*ReviewListResponse, error) {
	if err := details.update(); err != nil {
		return nil, err
	}
	return getReviewListResponse(pageToken, details.OpenReviews), nil
}
