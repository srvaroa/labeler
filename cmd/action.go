package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/google/go-github/v27/github"
	labeler "github.com/viaduct-ai/labeler/pkg"
	"golang.org/x/oauth2"
)

func main() {

	gh := getGithubClient()
	eventPayload := getEventPayload()
	eventName := os.Getenv("GITHUB_EVENT_NAME")

	// TODO: rethink this.  Currently we'll take the config from the
	// PR's branch, not from master.  My intuition is that one wants
	// to see the rules that are set in the main branch (as those are
	// vetted by the repo's owners).  It seems fairly common in GH
	// actions to use this approach, and I will need to consider
	// whatever branch is set as main in the repo settings, so leaving
	// as this for now.
	configRaw := getRepoFile(gh,
		os.Getenv("GITHUB_REPOSITORY"),
		os.Getenv("INPUT_CONFIG_PATH"),
		os.Getenv("GITHUB_SHA"))

	config := getLabelerConfig(configRaw)

	log.Printf("Re-evaluating labels on %s@%s",
		os.Getenv("GITHUB_REPOSITORY"),
		os.Getenv("GITHUB_SHA"))

	log.Printf("Trigger event: %s", os.Getenv("GITHUB_EVENT_NAME"))

	err := newLabeler(gh, config).HandleEvent(eventName, eventPayload)
	if err != nil {
		log.Fatalf("Unable to execute action: %+v", err)
	}

}

func getRepoFile(gh *github.Client, repo, file, sha string) *[]byte {

	t := strings.Split(repo, "/")
	owner, repoName := t[0], t[1]

	fileContent, _, _, err := gh.Repositories.GetContents(
		context.Background(),
		owner,
		repoName,
		file,
		&github.RepositoryContentGetOptions{Ref: sha})

	var content string
	if err == nil {
		content, err = fileContent.GetContent()
	}

	if err != nil {
		log.Fatalf("Unable to load configuration from %s@%s/%s: %s",
			repo, sha, file, err)
	}

	log.Printf("Loaded config from %s@%s:%s\n--\n%s", repo, sha, file, content)

	raw := []byte(content)
	return &raw
}

// getLabelerConfig builds a LabelerConfig from a raw yaml
func getLabelerConfig(configRaw *[]byte) *labeler.LabelerConfig {

	var c labeler.LabelerConfig

	err := yaml.Unmarshal(*configRaw, &c)
	if err != nil {
		log.Fatalf("Unable to unmarshall config --\n%s\n--, %s", configRaw, err)
	}
	fmt.Printf("The config: %+v", c)

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
