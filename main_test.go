package main

import (
	"testing"
	"time"
)

func TestGetRepo(t *testing.T) {
	//t.Parallel()
	testCases := []struct {
		name     string
		tags     []string
		expected string
	}{
		{name: "no tags", tags: []string{}, expected: "colinseymour.co.uk"},
		{name: "run", tags: []string{"foo", "run"}, expected: "gonefora.run"},
		{name: "tech", tags: []string{"foo", "tech"}, expected: "lildude.co.uk"},
		{name: "other", tags: []string{"foo", "boo"}, expected: "colinseymour.co.uk"},
		{name: "tech and run", tags: []string{"tech", "run"}, expected: "lildude.co.uk"},
		{name: "run and tech", tags: []string{"run", "tech"}, expected: "gonefora.run"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			//t.Parallel()
			repo := getRepo(tc.tags)
			if repo != tc.expected {
				t.Errorf("%v failed, got: %s, want: %s.", tc.name, repo, tc.expected)
			}
		})
	}

	s.GithubRepo = "lildude.github.io"
	t.Run("settings override", func(t *testing.T) {
		//t.Parallel()
		repo := getRepo([]string{"foo", "run"})
		if repo != s.GithubRepo {
			t.Errorf("%v failed, got: %s, want: %s.", "settings override", repo, s.GithubRepo)
}
	})
}

func TestCreateSlug(t *testing.T) {
	//t.Parallel()
	testCases := []struct {
		name     string
		expected string
	}{
		{name: "2010-01-04 01:02:03 UTC", expected: "2010-01-04-3723"},
		{name: "1979-09-30 15:16:17 BST", expected: "1979-09-30-51377"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			//t.Parallel()
			timeLayout := "2006-01-02 15:04:05 MST"
			time, _ := time.Parse(timeLayout, tc.name)
			slug := createSlug(&time)
			if slug != tc.expected {
				t.Errorf("%v failed, got: %s, want: %s.", tc.name, slug, tc.expected)
			}
		})
	}
}
