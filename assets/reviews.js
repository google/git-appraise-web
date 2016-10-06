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
    function(response) {
      $scope.details = response;
      loadSnippets(response.comments);
    });
  $http.get("/api/review_diff?repo=" + repo + "&review=" + review).success(
    function(response) {$scope.diff = response;});

  function loadSnippets(commentThreads) {
    var commentLocations = {};
    for (var i in commentThreads) {
      var commentThread = commentThreads[i];
      if ('location' in commentThread.comment) {
        var location = commentThread.comment.location;
        if (('commit' in location) && ('path' in location)) {
          var commit = location.commit;
          var path = location.path;
          if (!(commit in commentLocations)) {
            commentLocations[commit] = {};
          }
          var commitPaths = commentLocations[commit];
          if (!(path in commitPaths)) {
            commitPaths[path] = {};
          }
          var pathLines = commitPaths[path];
          if ('range' in location) {
            var range = location.range;
            if ('startLine' in range) {
              var line = range.startLine;
              if (!(line in pathLines)) {
                pathLines[line] = [];
              }
              var lineThreads = pathLines[line];
              lineThreads.push(commentThread.hash);
            }
          }
        }
      }
    }

    for (var commit in commentLocations) {
      var commitPaths = commentLocations[commit];
      for (var path in commitPaths) {
        var pathLines = commitPaths[path];
        var reader = snippetReader(path, commit, pathLines, commentThreads);
        $http.get("/api/repo_contents?repo=" + repo + "&commit=" + commit + "&file=" + path).success(reader);
      }
    }
  }

  function snippetReader(path, commit, pathLines, commentThreads) {
    return function(response) {
      var contentLines = response.split("\n");
      for (var line in pathLines) {
        if (line > 0 && line <= contentLines.length) {
          var startingLine = Math.max(0, line - 5);
          var endingLine = Math.max(0, line - 1);
          var snippetLines = [];
          for (var i = startingLine; i <= endingLine; i++) {
            snippetLines.push(new SnippetLine(i+1, contentLines[i]));
          }
          var snippet = new Snippet(commit, path, snippetLines);
          for (var h in pathLines[line]) {
            var hash = pathLines[line][h];
            for (var i in commentThreads) {
              var commentThread = commentThreads[i];
              if (commentThread.hash == hash) {
                commentThread.snippet = snippet;
              }
            }
          }
        }
      }
    };
  }

  function SnippetLine(lineNumber, contents) {
    this.lineNumber = lineNumber;
    this.contents = contents;
  }

  function Snippet(commit, path, lines) {
    this.commit = commit;
    this.friendlyCommit = commit.substring(0,6);
    this.path = path;
    this.lines = lines;
  }
});
