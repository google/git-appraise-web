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
	"sort"
	"testing"
)

func TestReposList(t *testing.T) {
	var reposList ReposList

	firstRepo := &RepoListItem{
		ID:   "0000",
		Path: "/b/c/d",
	}
	secondRepo := &RepoListItem{
		ID:   "1111",
		Path: "/a/b/c/d",
	}
	thirdRepo := &RepoListItem{
		ID:   "2222",
		Path: "/c/d",
	}
	reposList = append(reposList, secondRepo, firstRepo, thirdRepo)

	sort.Stable(reposList)
	if reposList[0] != firstRepo || reposList[1] != secondRepo || reposList[2] != thirdRepo {
		t.Fatalf("Unexpected repository ordering: %v", reposList)
	}
}
