package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
	labeler "labeler/pkg"
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
	configRaw, err := getRepoFile(gh,
		os.Getenv("GITHUB_REPOSITORY"),
		os.Getenv("INPUT_CONFIG_PATH"),
		os.Getenv("GITHUB_SHA"))
	if err != nil {
		return
	}

	config, err := getLabelerConfig(configRaw)
	if err != nil {
		return
	}

	log.Printf("Re-evaluating labels on %s@%s",
		os.Getenv("GITHUB_REPOSITORY"),
		os.Getenv("GITHUB_SHA"))

	log.Printf("Trigger event: %s", os.Getenv("GITHUB_EVENT_NAME"))

	l := newLabeler(gh, config)

	if eventName == "schedule" {
		t := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
		owner, repoName := t[0], t[1]

		prs, _, err := gh.PullRequests.List(context.Background(), owner, repoName, &github.PullRequestListOptions{})
		if err != nil {
			return
		}

		for _, pr := range prs {
			err = l.ExecuteOn(pr)
			log.Printf("Unable to execute action: %+v", err)
		}
	} else {
		err = l.HandleEvent(eventName, eventPayload)
		if err != nil {
			log.Printf("Unable to execute action: %+v", err)
		}
	}

}

func getRepoFile(gh *github.Client, repo, file, sha string) (*[]byte, error) {

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
		log.Printf("Unable to load configuration from %s@%s/%s: %s",
			repo, sha, file, err)
		return nil, err
	}

	log.Printf("Loaded config from %s@%s:%s\n--\n%s", repo, sha, file, content)

	raw := []byte(content)
	return &raw, err
}

// getLabelerConfig builds a LabelerConfigV1 from a raw yaml
func getLabelerConfig(configRaw *[]byte) (*labeler.LabelerConfigV1, error) {
	var c labeler.LabelerConfigV1
	err := yaml.Unmarshal(*configRaw, &c)
	if err != nil {
		log.Printf("Unable to unmarshall config %s: ", err)
	}
	if c.Version == 0 {
		c, err = getLabelerConfigV1(configRaw)
		if err != nil {
			log.Printf("Unable to unmarshall legacy config %s: ", err)
		}
	}
	return &c, err
}

func getLabelerConfigV1(configRaw *[]byte) (labeler.LabelerConfigV1, error) {

	// Load v0
	var oldCfg map[string]labeler.LabelMatcher
	err := yaml.Unmarshal(*configRaw, &oldCfg)
	if err != nil {
		log.Printf("Unable to unmarshall legacy config: %s", err)
		return labeler.LabelerConfigV1{}, err
	}

	// Convert
	var matchers = []labeler.LabelMatcher{}
	for label, matcher := range oldCfg {
		matcher.Label = label
		matchers = append(matchers, matcher)
	}

	return labeler.LabelerConfigV1{
		Version: 0,
		Labels:  matchers,
	}, err
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

func newLabeler(gh *github.Client, config *labeler.LabelerConfigV1) *labeler.Labeler {
	l := labeler.Labeler{

		FetchRepoConfig: func(owner string, repoName string) (*labeler.LabelerConfigV1, error) {
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

		Client: labeler.NewDefaultHttpClient(),
	}
	return &l
}
