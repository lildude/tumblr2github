package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

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

/*
func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
*/

func Test_formatPost(t *testing.T) {
	//t.Parallel()
	timeLayout := "2006-01-02 15:04:05 MST"
	timeStamp, _ := time.Parse(timeLayout, "1979-09-30 15:16:17 BST")
	tags := []string{"foo", "run"}

	type args struct {
		content string
		time    *time.Time
		tags    []string
	}
	tests := []struct {
		name string
		args args
		res  string
	}{
		{
			name: "simple text",
			args: args{"This is simple text", &timeStamp, tags},
			res:  "---\nlayout: post\ntags:\n- foo\n- run\ndate: 1979-09-30 15:16:17 +0100\n---\n\nThis is simple text\n",
		},
		{
			name: "markdown text",
			args: args{"This is **markdown** text", &timeStamp, tags},
			res:  "---\nlayout: post\ntags:\n- foo\n- run\ndate: 1979-09-30 15:16:17 +0100\n---\n\nThis is **markdown** text\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPost(tt.args.content, tt.args.time, tt.args.tags)
			if got != tt.res {
				t.Errorf("\ngot:\n%v\n---:=====:---\n\nwant:\n%v", got, tt.res)
			}
		})
	}
}

func Test_parseTextContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "text without html or markdown as html",
			content: "This is simple text",
			want:    "This is simple text",
		},
		{
			name:    "text without html or markdown as markdown",
			content: "This is simple text",
			want:    "This is simple text",
		},
		{
			name:    "text with html as html",
			content: "This is <b>simple</b> text",
			want:    "This is **simple** text",
		},
		{
			name:    "text with html as markdown",
			content: "This is <b>simple</b> text",
			want:    "This is **simple** text",
		},
		{
			name:    "text with markdown as html",
			content: "This is **simple** text",
			want:    "This is **simple** text",
		},
		{
			name:    "text with markdown as markdown",
			content: "This is **simple** text",
			want:    "This is **simple** text",
		},
		{
			name:    "text with markdown wrapped in html",
			content: "<p>This is **simple** text</p>",
			want:    "This is **simple** text",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTextContent(tt.content); got != tt.want {
				t.Errorf("parseTextContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
func TestGithubClient_newGitHubClient(t *testing.T) {
	type fields struct {
		Client *http.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   *github.Client
	}{
		{
			name:   "standard github client without auth",
			fields: fields{nil},
			want:   &github.Client{},
		},
		{
			name:   "standard github client with auth",
			fields: fields{nil},
			want:   &github.Client{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := &GithubClient{
				Client: tt.fields.Client,
			}
			if got := gc.newGitHubClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GithubClient.newGitHubClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/

/*
func TestGithubClient_postToGithub(t *testing.T) {
	type fields struct {
		Client *http.Client
	}
	type args struct {
		content    string
		postDate   *time.Time
		repository string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := &GithubClient{
				Client: tt.fields.Client,
			}
			gotRes, err := gc.postToGithub(tt.args.content, tt.args.postDate, tt.args.repository)
			if (err != nil) != tt.wantErr {
				t.Errorf("GithubClient.postToGithub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRes != tt.wantRes {
				t.Errorf("GithubClient.postToGithub() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
*/

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
			gc := &GithubClient{
				Client: httpClient,
			}
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
func TestGithubClient_repoHasPost(t *testing.T) {
	type fields struct {
		Client *http.Client
	}
	type args struct {
		filename string
		repo     string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := &GithubClient{
				Client: tt.fields.Client,
			}
			if got := gc.repoHasPost(tt.args.filename, tt.args.repo); got != tt.want {
				t.Errorf("GithubClient.repoHasPost() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
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

func Test_createSlug(t *testing.T) {
	timeLayout := "2006-01-02 15:04:05 MST"
	timeStamp, _ := time.Parse(timeLayout, "2010-01-04 01:02:03 UTC")

	type args struct {
		t *time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "slug for 2010-01-04 01:02:03 UTC",
			args: args{&timeStamp},
			want: "2010-01-04-3723",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createSlug(tt.args.t); got != tt.want {
				t.Errorf("createSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRepo(t *testing.T) {
	type args struct {
		tags []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "no tags", args: args{[]string{}}, want: "colinseymour.co.uk"},
		{name: "run", args: args{[]string{"foo", "run"}}, want: "gonefora.run"},
		{name: "tech", args: args{[]string{"foo", "tech"}}, want: "lildude.co.uk"},
		{name: "other", args: args{[]string{"foo", "boo"}}, want: "colinseymour.co.uk"},
		{name: "tech and run", args: args{[]string{"tech", "run"}}, want: "lildude.co.uk"},
		{name: "run and tech", args: args{[]string{"run", "tech"}}, want: "gonefora.run"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRepo(tt.args.tags); got != tt.want {
				t.Errorf("getRepo() = %v, want %v", got, tt.want)
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
