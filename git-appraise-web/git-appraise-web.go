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

package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/git-appraise-web/assets"
	"github.com/google/git-appraise/repository"
	"github.com/google/git-appraise/review"
)

var port int

func init() {
	flag.IntVar(&port, "port", 8080, "Port on which to start the server.")
}

func serveStaticContent(w http.ResponseWriter, r *http.Request) {
	resourceName := "assets/" + r.URL.Path[8:]
	contents, err := assets.Asset(resourceName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var contentType string
	if strings.HasSuffix(resourceName, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(resourceName, ".html") {
		contentType = "text/html"
	} else if strings.HasSuffix(resourceName, ".js") {
		contentType = "text/javascript"
	} else {
		contentType = http.DetectContentType(contents)
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(contents)
}

type repoListItem struct {
	ID   string
	Path string
}

type repoCache struct {
	Repo        repository.Repo
	RepoState   string
	OpenReviews []review.Summary
	AllReviews  []review.Summary
}

func (r *repoCache) update() error {
	stateHash, err := r.Repo.GetRepoStateHash()
	if err != nil {
		return err
	}
	if stateHash == r.RepoState {
		return nil
	}
	r.AllReviews = review.ListAll(r.Repo)
	var openReviews []review.Summary
	for _, review := range r.AllReviews {
		if !review.Submitted {
			openReviews = append(openReviews, review)
		}
	}
	r.OpenReviews = openReviews
	r.RepoState = stateHash
	return nil
}

func serveJson(v interface{}, w http.ResponseWriter) {
	json, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

func serveReposJson(repos map[string]*repoCache, w http.ResponseWriter, r *http.Request) {
	var reposList []repoListItem
	for id, cache := range repos {
		reposList = append(reposList, repoListItem{
			ID:   id,
			Path: cache.Repo.GetPath(),
		})
	}
	serveJson(reposList, w)
}

func serveAllReviewsJson(cache *repoCache, w http.ResponseWriter, r *http.Request) {
	if err := cache.update(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveJson(cache.AllReviews, w)
}

func serveOpenReviewsJson(cache *repoCache, w http.ResponseWriter, r *http.Request) {
	if err := cache.update(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveJson(cache.OpenReviews, w)
}

func getRepoCache(repos map[string]*repoCache, r *http.Request) (*repoCache, error) {
	repoParam := r.URL.Query().Get("repo")
	if repoParam == "" {
		return nil, errors.New("No repository specified")
	}
	cache, ok := repos[repoParam]
	if !ok {
		return nil, errors.New("Invalid repository specified")
	}
	return cache, nil
}

// Type for repo-specific HTTP handlers.
type repoFunc func(repo *repoCache, w http.ResponseWriter, r *http.Request)

// Convert a repo-specific function into an HTTP handler function.
func handleRepoFunc(repos map[string]*repoCache, f repoFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cache, err := getRepoCache(repos, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f(cache, w, r)
	}
}

// Serve our (fixed set of) URL paths
func serveRepos(repos map[string]*repoCache) {
	http.HandleFunc("/_ah/health",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "ok")
		})
	http.HandleFunc("/static/", serveStaticContent)
	http.HandleFunc("/repos", func(w http.ResponseWriter, r *http.Request) {
		serveReposJson(repos, w, r)
	})
	http.HandleFunc("/all_reviews", handleRepoFunc(repos, serveAllReviewsJson))
	http.HandleFunc("/open_reviews", handleRepoFunc(repos, serveOpenReviewsJson))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if len(repos) == 1 {
			for id := range repos {
				http.Redirect(w, r,
					"/static/reviews.html?repo="+id,
					http.StatusTemporaryRedirect)
				return
			}
		}
		http.Redirect(w, r,
			"/static/repos.html",
			http.StatusTemporaryRedirect)
		return
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// Get a fixed-length, obfuscated ID for the given repo.
func getRepoId(repo repository.Repo) string {
	return fmt.Sprintf("%.6x", sha1.Sum([]byte(repo.GetPath())))
}

// Find all local repositories under the current working directory.
func getLocalRepos() (map[string]*repoCache, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	repos := make(map[string]*repoCache)
	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			gitRepo, err := repository.NewGitRepo(path)
			if err == nil {
				repos[getRepoId(gitRepo)] = &repoCache{
					Repo: gitRepo,
				}
				return filepath.SkipDir
			}
		}
		return nil
	})
	return repos, nil
}

func main() {
	flag.Parse()
	repos, err := getLocalRepos()
	if err != nil {
		log.Fatal(err.Error())
	}
	if len(repos) == 0 {
		log.Fatal("Unable to find any local repositories under the current directory")
	}
	serveRepos(repos)
}
