package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
)

// Settings struct for pulling in our settings from environment variables.
type Settings struct {
	Port        string `required:"true" envconfig:"PORT"`
	GithubToken string `required:"false" split_words:"true"`
	// ConsumerKey    string `required:"true" split_words:"true"`
	// ConsumerSecret string `required:"true" split_words:"true"`
	// Token          string `required:"false"`
	// TokenSecret    string `required:"false" split_words:"true"`
	Secret string `required:"false"`
}

// WebHookPayload defines the structure of the IFTTT webhook payload.
type WebHookPayload struct {
	Secret        string `json:"secret"`
	PostTitle     string `json:"PostTitle"`
	PostURL       string `json:"PostUrl"`
	PostContent   string `json:"PostContent"`
	PostImageURL  string `json:"PostImageUrl"`
	PostTags      string `json:"PostTags"`
	PostPublished string `json:"PostPublished"`
}

var s Settings

func main() {
	err := envconfig.Process("tumblr", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+s.Port, nil)
}

// handler receives the webhook payload and does the magic
func handler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a body", 400)
		return
	}

	// Return OK as soon as we've received the payload - the webhook doesn't care what we do with the payload so no point holding things back.
	//w.WriteHeader(http.StatusOK)

	wh := new(WebHookPayload)
	err := json.NewDecoder(r.Body).Decode(&wh)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Temp log payload so I can verify how it's arriving
	b, err := json.Marshal(wh)
	log.Printf("DEBUG: %v\n", string(b))

	// Store the posturl in an environment variable and use to try catch duplicate deliveries
	lpu, _ := os.LookupEnv("LAST_POST_URL")
	if lpu != "" && lpu == wh.PostURL {
		log.Println("INFO: ignoring duplicate webhook delivery for ", wh.PostURL)
		return
	}
	os.Setenv("LAST_POST_URL", wh.PostURL)

}
