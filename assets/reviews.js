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

var gitAppraiseWeb=angular.module("gitAppraiseWeb", ["ngSanitize"]);

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

// Given a timestamp as the seconds from the unix epoch, return the human-friendly version.
function friendlyTimestamp(timestamp) {
  return new Date(parseInt(timestamp) * 1000).toLocaleString();
}

function parseUnifiedDiff(diff) {
  var fileDiffs = diff.split("\ndiff --git ");
  var files = [];
  for (var f in fileDiffs) {
    files.push(parseFileDiff(fileDiffs[f]));
  }
  return files;

  function parseFileDiff(fileDiff) {
    var fileLines = fileDiff.split("\n");
    var fileNames = fileLines[0].split(" ");
    var rhsFileName = fileNames[fileNames.length-1];
    var lhsFileName = fileNames[fileNames.length-2];
    var lhsNumber = 0;
    var rhsNumber = 0;
    var lines = [];
    for (var i in fileLines) {
      var text = fileLines[i];
      if (text.startsWith("--- ")) {
        lhsFileName = text.substring(4);
      } else if (text.startsWith("+++ ")) {
        rhsFileName = text.substring(4);
      } else if (text.startsWith("@@ ")) {
        var fromLinePart = text.substring(4, text.indexOf(","));
        var toLinePart = text.substring(text.indexOf("+"), text.lastIndexOf(","));
        var nextLhsNumber = parseInt(fromLinePart);
        var omittedCount = nextLhsNumber - lhsNumber;
        lhsNumber = nextLhsNumber;
        rhsNumber = parseInt(toLinePart);
        if (omittedCount > 0) {
          lines.push(
            new DiffLine(
              lhsNumber,
              rhsNumber,
              "omitted",
              "Skipped " + omittedCount + " unchanged lines"));
        }
      } else if (text.startsWith("-")) {
        lines.push(new DiffLine(lhsNumber, rhsNumber, "removed", text.substring(1)));
        lhsNumber++;
      } else if (text.startsWith("+")) {
        lines.push(new DiffLine(lhsNumber, rhsNumber, "added", text.substring(1)));
        rhsNumber++;
      } else if (text.startsWith(" ")) {
        lines.push(new DiffLine(lhsNumber, rhsNumber, "unchanged", text.substring(1)));
        lhsNumber++;
        rhsNumber++;
      }
    }
    var description = getFileDescription(lhsFileName, rhsFileName);
    var id = "file" + files.length;
    return new File(description, id, lines);
  }

  function getFileDescription(lhsFileName, rhsFileName) {
    if (lhsFileName.startsWith("a/") && rhsFileName.startsWith("b/")) {
      lhsFileName = lhsFileName.substring(2);
      rhsFileName = rhsFileName.substring(2);
      if (lhsFileName == rhsFileName) {
        return "Modified " + lhsFileName;
      } else {
        return "Renamed " + lhsFileName + " to " + rhsFileName;
      }
    } else if (lhsFileName.startsWith("a/")) {
      return "Deleted " + lhsFileName.substring(2);
    } else {
      return "Added " + rhsFileName.substring(2);
    }
  }

  function DiffLine(lhsNumber, rhsNumber, status, text) {
    this.lhsNumber = lhsNumber;
    this.rhsNumber = rhsNumber;
    this.status = status;
    this.text = text;
  }

  function File(description, id, diffLines) {
    this.description = description;
    this.id = id;
    this.diffLines = diffLines;
    this.display = true;
  }
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
  $http.get("/api/open_reviews?repo=" + repo).success(
    function(response) {$scope.openReviews = processListReviewsResponse(response);});
  $http.get("/api/closed_reviews?repo=" + repo).success(
    function(response) {$scope.closedReviews = processListReviewsResponse(response);});

  function processListReviewsResponse(response) {
    var reviews = [];
    for (var i in response) {
      var revision = response[i].revision;
      var timestamp = response[i].request.timestamp;
      var desc = response[i].request.description;
      reviews.push(new Review(revision, timestamp, desc, getSummary(desc)));
    }
    return reviews;
  }

  function Review(revision, timestamp, desc, summary) {
    this.revision = revision;
    this.timestamp = friendlyTimestamp(timestamp);
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
    function(response) {$scope.details = processReview(response);});
  $http.get("/api/review_diff?repo=" + repo + "&review=" + review).success(
    function(response) {
      $scope.diff = response;
      $scope.diff.files = parseUnifiedDiff(response.contents);
      $scope.diff.reviewCommits = friendlyCommits($scope.diff.reviewCommits);
    });

  function friendlyCommits(commits) {
    for (var i in commits) {
      var commit = commits[i];
      commit.name = commit.id.substring(0,6);
      commit.details.time = friendlyTimestamp(commit.details.time);
    }
    commits.reverse();
    return commits
  }

  function friendlyCommentTimestamps(commentThread) {
    var timestamp = commentThread.comment.timestamp;
    commentThread.comment.timestamp = friendlyTimestamp(timestamp);
    for (var i in commentThread.children) {
      friendlyCommentTimestamps(commentThread.children[i]);
    }
  }

  function generateCommentHtml(converter, commentThread) {
    commentThread.html = converter.makeHtml(commentThread.comment.description);
    commentThread.display = true;
    if ('resolved' in commentThread) {
      if (commentThread.resolved) {
        commentThread.status = 'lgtm';
      } else {
        commentThread.status = 'nmw';
      }
    } else {
      commentThread.status = 'fyi';
    }
    for (var i in commentThread.children) {
      generateCommentHtml(converter, commentThread.children[i]);
    }
  }

  function friendlyTimestamps(reviewDetails) {
    for (var i in reviewDetails.reports) {
      var report = reviewDetails.reports[i];
      var timestamp = report.timestamp;
      report.timestamp = friendlyTimestamp(timestamp);
    }
    for (var i in reviewDetails.comments) {
      friendlyCommentTimestamps(reviewDetails.comments[i]);
    }
  }

  function processReview(reviewDetails) {
    friendlyTimestamps(reviewDetails);
    converter = new showdown.Converter();
    for (var i in reviewDetails.comments) {
      generateCommentHtml(converter, reviewDetails.comments[i]);
    }
    return reviewDetails;
  }
});
