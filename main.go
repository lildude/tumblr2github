package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"text/template"
	"time"

	"github.com/google/go-github/v21/github"
	"github.com/kelseyhightower/envconfig"
	"github.com/lunny/html2md"
	"golang.org/x/oauth2"

	tumblr "github.com/tumblr/tumblr.go"
	tumblrclient "github.com/tumblr/tumblrclient.go"
)

// Settings struct for pulling in our settings from environment variables.
type Settings struct {
	Port              string `required:"true" envconfig:"PORT"`
	GithubToken       string `required:"true" split_words:"true"`
	GithubRepo        string `required:"false" split_words:"true"`
	GithubUser        string `required:"true" split_words:"true"`
	GithubAuthorName  string `required:"true" split_words:"true"`
	GithubAuthorEmail string `required:"true" split_words:"true"`
	ConsumerKey       string `required:"true" split_words:"true"`
	ConsumerSecret    string `required:"true" split_words:"true"`
	UserToken         string `required:"true" split_words:"true"`
	UserTokenSecret   string `required:"true" split_words:"true"`
	BlogID            string `required:"true" split_words:"true"`
}

type poster struct {
	URL    string `json:"url"`
	Type   string `json:"type"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}

type response struct {
	Type        string `json:"type"`
	URL         string `json:"url"`
	DisplayURL  string `json:"display_url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SiteName    string `json:"site_name"`
	Poster      poster
}

type blogPost struct {
	Date    string
	Tags    []string
	Content string
}

type GithubClient struct {
	Client *http.Client
}

var s Settings
var client *github.Client
var ctx = context.Background()

func main() {
	err := envconfig.Process("tumblr", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	cl := tumblrclient.NewClientWithToken(
		s.ConsumerKey,
		s.ConsumerSecret,
		s.UserToken,
		s.UserTokenSecret,
	)

	params := url.Values{}
	params.Add("filter", "raw")
	params.Add("limit", "5")

	posts, err := tumblr.GetPosts(cl, s.BlogID, params)
	if err != nil {
		fmt.Printf("%v\n", err)
		fmt.Printf("%v\n", posts)
	}

	gc := new(GithubClient)
	allPosts, _ := posts.All()
	for _, post := range allPosts {
		//fmt.Printf("%v\n", post)

		switch pt := post.(type) {
		case *tumblr.LinkPost:
			fmt.Printf("INFO: link   %d %v %v\n", pt.Id, pt.Url, pt.Tags)
		case *tumblr.PhotoPost:
			fmt.Printf("INFO: photo   %d %v %v\n", pt.Id, pt.ImagePermalink, pt.Tags)
		case *tumblr.QuotePost:
			fmt.Printf("INFO: quote   %d %v %v\n", pt.Id, pt.Source, pt.Tags)
		case *tumblr.TextPost:
			content := parseTextContent(pt.Trail[0].ContentRaw)
			fmt.Printf("INFO: text   %d %v %v\n", pt.Id, content, pt.Tags)
			timeLayout := "2006-01-02 15:04:05 MST"
			t, err := time.Parse(timeLayout, pt.Date)
			if err != nil {
				log.Fatal(err.Error())
			}
			cont := formatPost(content, &t, pt.Tags)
			res, err := gc.postToGithub(cont, &t, getRepo(pt.Tags))
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Println(res)
		default:
			continue
		}
	}
}

func formatPost(content string, time *time.Time, tags []string) (res string) {
	c := blogPost{time.Format("2006-01-02 15:04:05 -0700"), tags, content}
	fmtTmpl := `---
layout: post
{{- if .Tags }}
tags:
{{- range .Tags }}
- {{.}}
{{- end }}
{{- end }}
date: {{.Date}}
---

{{.Content}}
`
	tmpl := template.Must(template.New("blogpost").Parse(fmtTmpl))

	var out bytes.Buffer
	tmpl.Execute(&out, c)

	return out.String()
}

func parseTextContent(content string) string {
	// If the content contains `data-npf`, it means it's got an embedded link, so lets treat it like we'd like to treat a link post.
	// Interestingly, creating a link post from mobile actually creates a text post. Web creates a link post.
	re := regexp.MustCompile("data-npf='({.*})'")
	linkData := re.FindStringSubmatch(content)
	if len(linkData) > 0 && linkData[1] != "" {
		fmt.Printf("%v\n", linkData)
		ld := &response{}
		json.Unmarshal([]byte(linkData[1]), &ld)
		fmt.Println(ld.URL)
	}

	content = html2md.Convert(content)

	return content
}

func (gc *GithubClient) newGitHubClient() *github.Client {
	tc := gc.Client
	if gc.Client == nil {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: s.GithubToken})
		tc = oauth2.NewClient(ctx, ts)
	}
	return github.NewClient(tc)
}

func (gc *GithubClient) postToGithub(content string, postDate *time.Time, repository string) (res string, err error) {
	client := gc.newGitHubClient()

	filename := fmt.Sprintf("%s.md", createSlug(postDate))

	// Bail early if the post already exists.
	if gc.repoHasPost(filename, repository) {
		return "INFO: Post already exists. Nothing to do.", nil
	}

	ref, _, err := client.Git.GetRef(ctx, s.GithubUser, repository, "refs/heads/master")
	if err != nil {
		return "", err
	}

	// Create a tree with what to commit.
	entries := []github.TreeEntry{}
	entries = append(entries, github.TreeEntry{Path: github.String("_posts/" + filename), Type: github.String("blob"), Content: github.String(string(content)), Mode: github.String("100644")})
	tree, _, err := client.Git.CreateTree(ctx, s.GithubUser, repository, *ref.Object.SHA, entries)
	if err != nil {
		return "", err
	}
	// createCommit creates the commit in the given reference using the given tree.
	// Get the parent commit to attach the commit to.
	parent, _, err := client.Repositories.GetCommit(ctx, s.GithubUser, repository, *ref.Object.SHA)
	if err != nil {
		return "", err
	}

	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	// Create the commit using the tree.
	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: github.String(s.GithubAuthorName), Email: github.String(s.GithubAuthorEmail)}
	commit := &github.Commit{Author: author, Message: github.String("New note: " + filename), Tree: tree, Parents: []github.Commit{*parent.Commit}}
	newCommit, _, err := client.Git.CreateCommit(ctx, s.GithubUser, repository, commit)
	if err != nil {
		return "", err
	}

	// Attach the commit to the master branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = client.Git.UpdateRef(ctx, s.GithubUser, repository, ref, false)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("INFO: New post created: %s", filename), nil
}

func (gc *GithubClient) repoHasPost(filename string, repo string) bool {
	client := gc.newGitHubClient()
	query := fmt.Sprintf("filename:%s repo:%s/%s path:_posts", filename, s.GithubUser, repo)
	opts := &github.SearchOptions{Sort: "forks", Order: "desc"}
	res, _, err := client.Search.Code(ctx, query, opts)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	if res.GetTotal() > 0 {
		return true
	}
	return false
}

func createSlug(t *time.Time) string {
	return fmt.Sprintf("%s-%d", t.Format("2006-01-02"), t.Unix()%(24*60*60))
}

func getRepo(tags []string) string {
	if s.GithubRepo != "" {
		return s.GithubRepo
	}
	for _, n := range tags {
		if n == "run" {
			return "gonefora.run"
		} else if n == "tech" {
			return "lildude.co.uk"
		}
	}
	return "colinseymour.co.uk"
}
