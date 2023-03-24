package main

import (
	"testing"
)

var (
	testsRan    = 0
	testsPassed = 0
)

//"github.com/google/go-github/v50/github"
//"golang.org/x/oauth2"

func TestGithubPullReq(t *testing.T) {
	repoNames := []string{
		"ach1ntya/461-project",
		"golang/go",
		"kubernetes/kubernetes",
		"tensorflow/tensorflow",
		"facebook/react",
		"nodejs/node",
		"django/django",
		"vuejs/vue",
		"angular/angular",
		"ruby/ruby",
		"rails/rails",
		"expressjs/express",
		"meteor/meteor",
		"laravel/laravel",
		"spring-projects/spring-framework",
		"microsoft/dotnet",
		"aspnet/aspnetcore",
		"flutter/flutter",
		"apple/swift",
		"dart-lang/sdk",
	}
	expectedCounts := []int{
		24,
		2811,
		72782,
		22519,
		13063,
		29421,
		16421,
		2357,
		23109,
		7211,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
	}

	for i, repoName := range repoNames {
		testsRan++
		count := githubPullReq(repoName)
		if count != expectedCounts[i] {
			t.Errorf("For repo %s, expected count to be %d, but got %d", repoName, expectedCounts[i], count)
		} else {
			testsPassed++
		}
	}
	// outString := string(out)
	// var coverage int
	// fmt.Sscanf(outString, "coverage: %d%%\n", &coverage)
	// fmt.Printf("%d/%d test cases passed. %d%% line coverage achieved.", testsPassed, testsRan, coverage)
}
