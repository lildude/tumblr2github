package main

import (
	"fmt"
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

func TestPostToGithub(t *testing.T) {
	//t.Parallel()
	testCases := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "text post not created as already exists",
			content:  `{"total_count":1}`,
			expected: "INFO: Post already exists. Nothing to do.",
		},
		{
			name:     "text post created",
			content:  `{"total_count":0}`,
			expected: "INFO: New post created: 2010-01-04-3723.md",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			//t.Parallel()
			httpClient := NewTestClient(func(req *http.Request) *http.Response {
				fmt.Printf("%v: %v\n", req.Method, req.URL.Path)
				var content string
				resp := http.Response{}
				resp.StatusCode = 200
				switch {
				case strings.Contains(req.URL.Path, "search/code"):
					content = tc.content
				case strings.Contains(req.URL.Path, "git/refs/heads/master"):
					content = `{"ref": "refs/heads/master","object": {"sha": "8613c69c7075ae1e84a6d06054a402d0f3213c1b"}}`
				case strings.Contains(req.URL.Path, "git/trees"):
					content = ``
				case strings.Contains(req.URL.Path, "commits/8613c69c7075ae1e84a6d06054a402d0f3213c1"):
					content = `{
						"sha": "8613c69c7075ae1e84a6d06054a402d0f3213c1b",
						"commit": {
								"author": {
										"name": "lildude",
										"email": "lildood@gmail.com",
										"date": "2018-10-21T16:48:02Z"
								},
								"committer": {
										"name": "lildude",
										"email": "lildood@gmail.com",
										"date": "2018-10-21T16:48:02Z"
								},
								"message": "New note: 2018-10-12-55235.md",
								"tree": {
										"sha": "e1343d30894799dbaf4438d12d9862c0e48d1857",
										"url": "https://api.github.com/repos/lildude/lildude.github.io/git/trees/e1343d30894799dbaf4438d12d9862c0e48d1857"
								},
								"url": "https://api.github.com/repos/lildude/lildude.github.io/git/commits/8613c69c7075ae1e84a6d06054a402d0f3213c1b",
								"comment_count": 0,
								"verification": {
										"verified": false,
										"reason": "unsigned",
										"signature": null,
										"payload": null
								}
						}
					}`
				}
				resp.Body = ioutil.NopCloser(strings.NewReader(content))

				return &resp
			})
			gc := new(GithubClient)
			gc.Client = httpClient
			s.GithubUser = "lildude"
			timeLayout := "2006-01-02 15:04:05 MST"
			time, _ := time.Parse(timeLayout, "2010-01-04 01:02:03 UTC")
			res, _ := gc.postToGithub("foo", &time, "lildude.github.io")
			if res != tc.expected {
				t.Errorf("%v failed, got: %v, want: %v.", tc.name, res, tc.expected)
			}
		})
	}
}

/*
func TestFormatPostFailure(t *testing.T) {

}



*/
