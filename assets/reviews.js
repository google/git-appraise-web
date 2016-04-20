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

gitAppraiseWeb.controller("listRepos", function($scope,$http) {
    $http.get("/api/repos").success(
	function(response) {$scope.repositories = response;});
});

gitAppraiseWeb.controller("listReviews", function($scope,$http,$location) {
    var repo = $location.search()['repo'];
    $scope.repo = repo;
    $http.get("/api/open_reviews?repo=" + repo).success(
	function(response) {$scope.openReviews = processListReviewsResponse(response);});
    $http.get("/api/closed_reviews?repo=" + repo).success(
	function(response) {$scope.closedReviews = processListReviewsResponse(response);});

    function processListReviewsResponse(response) {
	var reviews = [];
	for (var i in response) {
	    var revision = response[i].revision;
	    var desc = response[i].request.description;
	    reviews.push(new Review(revision, desc, getSummary(desc)));
	}
	return reviews;
    }

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

    function Review(revision, desc, summary) {
	this.revision = revision;
	this.desc = desc;
	this.summary = summary;
    }
});

gitAppraiseWeb.controller("getReview", function($scope,$http,$location) {
    var repo = $location.search()['repo'];
    var review = $location.search()['review'];
    $http.get("/api/review_details?repo=" + repo + "&review=" + review).success(
	function(response) {$scope.details = response;});
    $http.get("/api/review_diff?repo=" + repo + "&review=" + review).success(
	function(response) {$scope.diff = response;});
});
