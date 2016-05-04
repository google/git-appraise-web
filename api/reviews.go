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
	"github.com/google/git-appraise/repository"
	"github.com/google/git-appraise/review"
)

// CommitOverview encapsulates the fine-grained details of a commit.
type CommitOverview struct {
	ID      string                    `json:"id"`
	Details *repository.CommitDetails `json:"details"`
}

// DiffSummary summarizes one of the diffs included in a review.
//
// This summary includes the list of all commits that can be used to construct such a diff.
type DiffSummary struct {
	ReviewCommits []CommitOverview `json:"reviewCommits,omitEmpty"`
	LeftHandSide  string           `json:"leftHandSide"`
	RightHandSide string           `json:"rightHandSide"`
	Contents      string           `json:"contents"`
}

// ReviewListResponse represents a single `page` in a list of reviews.
type ReviewListResponse struct {
	Items         []review.Summary `json:"items"`
	NextPageToken string           `json:"nextPageToken,omitEmpty"`
}

func paginateReviews(reviews []review.Summary, maxPerPage int) [][]review.Summary {
	var results [][]review.Summary
	var currID int
	var currPage []review.Summary
	for _, r := range reviews {
		if currID >= maxPerPage {
			results = append(results, currPage)
			currPage = nil
			currID = 0
		}
		currPage = append(currPage, r)
		currID++
	}
	results = append(results, currPage)
	return results
}

// getReviewBase gets the earliest commit that can be used as a left-hand-side
// when generating diffs for a review.
//
// This is different from the `GetBaseCommit` method because that one only
// focuses on the current diff, whereas this method also includes historical diffs.
func getReviewBase(r *review.Review) (string, error) {
	if r.Request.BaseCommit != "" {
		return r.Request.BaseCommit, nil
	}
	submittedBase, err := r.Repo.GetLastParent(r.Revision)
	if err != nil {
		return "", err
	}
	if r.Submitted {
		return submittedBase, nil
	}
	targetRefBase, err := r.Repo.ResolveRefCommit(r.Request.TargetRef)
	if err != nil {
		return "", err
	}
	return r.Repo.MergeBase(targetRefBase, submittedBase)
}

// NewDiffSummary constructs a new instance of DiffSummary.
//
// If the `lhs` or `rhs` arguments are empty or out of bounds, then the
// review's base or head commits are used instead.
func NewDiffSummary(reviewDetails *review.Review, lhs, rhs string) (*DiffSummary, error) {
	base, err := getReviewBase(reviewDetails)
	if err != nil {
		return nil, err
	}
	head, err := reviewDetails.GetHeadCommit()
	if err != nil {
		return nil, err
	}
	reviewCommits := []string{base}
	subsequentCommits, err := reviewDetails.Repo.ListCommitsBetween(base, head)
	if err != nil {
		return nil, err
	}
	reviewCommits = append(reviewCommits, subsequentCommits...)
	commitsMap := make(map[string]interface{})
	var commitOverviews []CommitOverview
	for _, commit := range reviewCommits {
		commitsMap[commit] = nil
		details, err := reviewDetails.Repo.GetCommitDetails(commit)
		if err != nil {
			return nil, err
		}
		commitOverviews = append(commitOverviews, CommitOverview{
			ID:      commit,
			Details: details,
		})
	}
	if _, ok := commitsMap[lhs]; !ok {
		lhs, err = reviewDetails.GetBaseCommit()
		if err != nil {
			return nil, err
		}
	}
	if _, ok := commitsMap[rhs]; !ok {
		rhs = head
	}
	diff, err := reviewDetails.Repo.Diff(lhs, rhs)
	if err != nil {
		return nil, err
	}
	return &DiffSummary{
		ReviewCommits: commitOverviews,
		LeftHandSide:  lhs,
		RightHandSide: rhs,
		Contents:      diff,
	}, nil
}
