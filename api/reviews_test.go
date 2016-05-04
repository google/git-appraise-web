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

	"testing"
)

func TestDiffSummary(t *testing.T) {
	repo := repository.NewMockRepoForTest()
	repoDetails := NewRepoDetails(repo)
	reviewDetails, err := repoDetails.GetReview(repository.TestCommitG)
	if err != nil {
		t.Fatal(err)
	}

	fullDiffSummary, err := NewDiffSummary(reviewDetails, "", "")
	if len(fullDiffSummary.ReviewCommits) != 5 {
		t.Fatalf("Unexpected list of included diffs: %v", fullDiffSummary.ReviewCommits)
	}

	if fullDiffSummary.LeftHandSide != repository.TestCommitF ||
		fullDiffSummary.RightHandSide != repository.TestCommitI {
		t.Fatalf("Unexpected full diff: %v", fullDiffSummary)
	}

	expectedDiff, err := repo.Diff(repository.TestCommitE, repository.TestCommitG)
	if err != nil {
		t.Fatal(err)
	}

	firstDiffSummary, err := NewDiffSummary(reviewDetails, repository.TestCommitE, repository.TestCommitG)
	if len(firstDiffSummary.ReviewCommits) != 5 {
		t.Fatalf("Unexpected list of included diffs: %v", firstDiffSummary.ReviewCommits)
	}
	if firstDiffSummary.Contents != expectedDiff {
		t.Fatalf("Unexpected first diff: %v", fullDiffSummary)
	}
}
