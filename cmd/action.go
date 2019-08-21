package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-github/v26/github"
	"github.com/srvaroa/labeler/pkg"
	"golang.org/x/oauth2"
)

func main() {

	ghToken := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	gh := github.NewClient(tc)

	payloadPath := os.Getenv("GITHUB_EVENT_PATH")
	file, err := os.Open(payloadPath)
	if err != nil {
		log.Fatalf("Failed to open event payload file %s: %s", err)
	}
	eventPayload, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to load event payload from %s: %s", err)
	}
	eventName := os.Getenv("GITHUB_EVENT_NAME")

	l := labeler.NewLabeler(gh)
	l.HandleEvent(eventName, &eventPayload)
}
