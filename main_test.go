package main

import (
	"io/ioutil"
	"net/http"
	"strings"
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

func TestFormatPost(t *testing.T) {
	//t.Parallel()
	testCases := []struct {
		name     string
		date     string
		content  string
		expected string
	}{
		{
			name:     "simple text",
			date:     "2010-01-04 01:02:03 UTC",
			content:  "This is simple text",
			expected: "---\nlayout: post\ntags:\n- foo\n- run\ndate: 2010-01-04 01:02:03 +0000 UTC\n---\n\nThis is simple text\n",
		},
		{
			name:     "markdown text",
			date:     "1979-09-30 15:16:17 BST",
			content:  "This is **markdown** text",
			expected: "---\nlayout: post\ntags:\n- foo\n- run\ndate: 1979-09-30 15:16:17 +0100 BST\n---\n\nThis is **markdown** text\n",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			//t.Parallel()
			timeLayout := "2006-01-02 15:04:05 MST"
			time, _ := time.Parse(timeLayout, tc.date)
			post, _ := formatPost(tc.content, &time, []string{"foo", "run"})
			if post != tc.expected {
				t.Errorf("%v failed, got: %s, want: %s.", tc.name, post, tc.expected)
			}
		})
	}
}

func TestParseTextContent(t *testing.T) {
	//t.Parallel()
	testCases := []struct {
		name     string
		content  string
		format   string
		expected string
	}{
		{
			name:     "text without html or markdown as html",
			content:  "This is simple text",
			format:   "html",
			expected: "This is simple text",
		},
		{
			name:     "text without html or markdown as markdown",
			content:  "This is simple text",
			format:   "markdown",
			expected: "This is simple text",
		},
		{
			name:     "text with html as html",
			content:  "This is <b>simple</b> text",
			format:   "html",
			expected: "This is **simple** text",
		},
		{
			name:     "text with html as markdown",
			content:  "This is <b>simple</b> text",
			format:   "markdown",
			expected: "This is <b>simple</b> text",
		},
		{
			name:     "text with markdown as html",
			content:  "This is **simple** text",
			format:   "html",
			expected: "This is **simple** text",
		},
		{
			name:     "text with markdown as markdown",
			content:  "This is **simple** text",
			format:   "markdown",
			expected: "This is **simple** text",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			//t.Parallel()
			c := parseTextContent(tc.content, tc.format)
			if c != tc.expected {
				t.Errorf("%v failed, got: %s, want: %s.", tc.name, c, tc.expected)
			}
		})
	}
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestRepoHasPost(t *testing.T) {
	//t.Parallel()
	testCases := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "text post does not exist",
			content:  `{"total_count":0}`,
			expected: false,
		},
		{
			name:     "text post exists already",
			content:  `{"total_count":1}`,
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			//t.Parallel()
			httpClient := NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: 200,
					// Send response to be tested
					Body: ioutil.NopCloser(strings.NewReader(tc.content)),
				}
			})
			gc := new(GithubClient)
			gc.Client = httpClient
			res := gc.repoHasPost("foo", "lildude.github.io")
			if res != tc.expected {
				t.Errorf("%v failed, got: %v, want: %v.", tc.name, res, tc.expected)
			}
		})
	}
}

/*
func TestFormatPostFailure(t *testing.T) {

}

func TestPostToGithub(t *testing.T) {

}

*/
