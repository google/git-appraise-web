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

var gitAppraiseWeb=angular.module("gitAppraiseWeb", []);

// Get a repository name from the full path.
function getLastPathElement(path) {
  var slashIndex = path.lastIndexOf("/");
  if (slashIndex > 0) {
    return path.substring(slashIndex+1, path.length);
  }
  return path;
}

// Get a review summary from the full description.
function getSummary(desc) {
  var result = desc;
  var newlineIndex = desc.indexOf("\n");
  if (newlineIndex > 0) {
    result = desc.substring(0, newlineIndex);
  }
  if (result.length > 80) {
    result = result.substring(0, 80);
  }
  return result;
}

gitAppraiseWeb.controller("listRepos", function($scope,$http) {
  $http.get("/api/repos").success(
    function(response) {$scope.repositories = processListReposResponse(response);});

  function processListReposResponse(response) {
    var repos = [];
    for (var i in response) {
      var path = response[i].path;
      repos.push(new Repo(response[i].id, getLastPathElement(path)));
    }
    return repos;
  }

  function Repo(id, name) {
    this.id = id;
    this.name = name;
  }
});

gitAppraiseWeb.controller("listReviews", function($scope,$http,$location) {
  var repo = $location.search()['repo'];
  $scope.repo = repo;
  $http.get("/api/repo_summary?repo=" + repo).success(
    function(response) {$scope.path = getLastPathElement(response.path);});
  listAllReviews("/api/open_reviews?repo=" + repo, function(response){
    $scope.openReviews = response;
  });
  listAllReviews("/api/closed_reviews?repo=" + repo, function(response){
    $scope.closedReviews = response;
  });

  function addPageItems(page, reviews) {
    for (var i in page.items) {
      var revision = page.items[i].revision;
      var timestamp = page.items[i].request.timestamp;
      var desc = page.items[i].request.description;
      reviews.push(new Review(revision, timestamp, desc, getSummary(desc)));
    }
  }

  function listReviewsResponseProcessor(baseQuery, callback) {
    var reviews = [];
    function processor(response) {
      addPageItems(response, reviews);
      if (response.nextPageToken) {
        $http.get(baseQuery + "&page=" + response.nextPageToken).success(processor);
      } else {
        callback(reviews);
      }
    }
    return processor;
  }

  function listAllReviews(baseQuery, callback) {
    $http.get(baseQuery).success(listReviewsResponseProcessor(baseQuery, callback));
  }

  function Review(revision, timestamp, desc, summary) {
    this.revision = revision;
    this.timestamp = timestamp;
    this.desc = desc;
    this.summary = summary;
  }
});

gitAppraiseWeb.controller("getReview", function($scope,$http,$location) {
  var repo = $location.search()['repo'];
  var review = $location.search()['review'];
  $http.get("/api/repo_summary?repo=" + repo).success(
    function(response) {$scope.path = getLastPathElement(response.path);});
  $http.get("/api/review_details?repo=" + repo + "&review=" + review).success(
    function(response) {$scope.details = response;});
  $http.get("/api/review_diff?repo=" + repo + "&review=" + review).success(
    function(response) {
      $scope.diff = response;
      $scope.diff.reviewCommits = friendlyCommits($scope.diff.reviewCommits);
    });

  function friendlyCommits(commits) {
    for (var i in commits) {
      var commit = commits[i];
      commit.name = commit.id.substring(0,6);
    }
    commits.reverse();
    return commits
  }
});
