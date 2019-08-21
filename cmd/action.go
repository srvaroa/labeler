package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-github/v27/github"
	labeler "github.com/srvaroa/labeler/pkg"
	"golang.org/x/oauth2"
)

func main() {

	gh := getGithubClient()
	eventPayload := getEventPayload()
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	config := getLabelerConfig()

	log.Printf("Re-evaluating labels on %s@%s",
		os.Getenv("GITHUB_REPOSITORY"),
		os.Getenv("GITHUB_SHA"))
	log.Printf("Trigger event: %s", os.Getenv("GITHUB_EVENT_NAME"))

	err := newLabeler(gh, config).HandleEvent(eventName, eventPayload)
	if err != nil {
		log.Fatalf("Unable to execute action: %+v", err)
	}

}

func getLabelerConfig() *labeler.LabelerConfig {
	c := labeler.LabelerConfig{
		"WIP": labeler.LabelMatcher{
			Title: "^WIP:.*",
		},
	}
	return &c
}

func getGithubClient() *github.Client {
	ghToken := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func getEventPayload() *[]byte {
	payloadPath := os.Getenv("GITHUB_EVENT_PATH")
	file, err := os.Open(payloadPath)
	if err != nil {
		log.Fatalf("Failed to open event payload file %s: %s", payloadPath, err)
	}
	eventPayload, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to load event payload from %s: %s", payloadPath, err)
	}
	return &eventPayload
}

func newLabeler(gh *github.Client, config *labeler.LabelerConfig) *labeler.Labeler {
	l := labeler.Labeler{

		FetchRepoConfig: func(owner string, repoName string) (*labeler.LabelerConfig, error) {
			return config, nil
		},

		ReplaceLabelsForPr: func(owner string, repoName string, prNumber int, labels []string) error {
			log.Printf("Setting labels to %s/%s#%d: %s", owner, repoName, prNumber, labels)
			_, _, err := gh.Issues.ReplaceLabelsForIssue(
				context.Background(), owner, repoName, prNumber, labels)
			return err
		},

		GetCurrentLabels: func(owner string, repoName string, prNumber int) ([]string, error) {
			opts := github.ListOptions{} // TODO: ignoring pagination here
			currLabels, _, err := gh.Issues.ListLabelsByIssue(
				context.Background(), owner, repoName, prNumber, &opts)

			labels := []string{}
			for _, label := range currLabels {
				labels = append(labels, *label.Name)
			}
			return labels, err
		},
	}
	return &l
}
