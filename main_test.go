package main

import (
	"testing"
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

}
