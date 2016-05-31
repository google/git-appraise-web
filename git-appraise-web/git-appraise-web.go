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
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/git-appraise-web/api"
	"github.com/google/git-appraise-web/third_party/assets"
	"github.com/google/git-appraise/repository"
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

// Serve our (fixed set of) URL paths
func serveRepos(cache api.RepoCache) {
	http.HandleFunc("/_ah/health",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "ok")
		})
	http.HandleFunc("/static/", serveStaticContent)
	http.HandleFunc("/api/repos", cache.ServeListReposJSON)
	http.HandleFunc("/api/repo_summary", cache.ServeRepoSummaryJSON)
	http.HandleFunc("/api/repo_contents", cache.ServeRepoContents)
	http.HandleFunc("/api/closed_reviews", cache.ServeClosedReviewsJSON)
	http.HandleFunc("/api/open_reviews", cache.ServeOpenReviewsJSON)
	http.HandleFunc("/api/review_details", cache.ServeReviewDetailsJSON)
	http.HandleFunc("/api/review_diff", cache.ServeReviewDiff)
	http.HandleFunc("/", cache.ServeEntryPointRedirect)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// Find all local repositories under the current working directory.
func getLocalRepos() (api.RepoCache, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	repos := make(api.RepoCache)
	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			gitRepo, err := repository.NewGitRepo(path)
			if err == nil {
				repos.AddRepo(gitRepo)
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
