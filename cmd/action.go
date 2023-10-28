package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/google/go-github/v50/github"
	labeler "github.com/srvaroa/labeler/pkg"
	"golang.org/x/oauth2"
)

func main() {

	gh, err := getGithubClient()
	if err != nil {
		log.Printf("Failed to retrieve a GitHub client: %+v", err)
		return
	}
	eventPayload := getEventPayload()
	eventName := os.Getenv("GITHUB_EVENT_NAME")

	// Determine if the user wants to override the upstream config
	// in the main branch with the local one in the checkout
	useLocalConfig, err := strconv.ParseBool(os.Getenv("INPUT_USE_LOCAL_CONFIG"))
	if err != nil {
		useLocalConfig = false
	}

	configFile := os.Getenv("INPUT_CONFIG_PATH")

	var configRaw *[]byte
	if useLocalConfig {
		log.Printf("Reading configuration from local file: %s", configFile)
		contents, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Printf("Error reading configuration from local file: %s", err)
			return
		}
		configRaw = &contents
	} else {
		log.Printf("Reading configuration file from the repository default branch: %s", configFile)
		// TODO: rethink this.  Currently we'll take the config from the
		// PR's branch, not from master.  My intuition is that one wants
		// to see the rules that are set in the main branch (as those are
		// vetted by the repo's owners).  It seems fairly common in GH
		// actions to use this approach, and I will need to consider
		// whatever branch is set as main in the repo settings, so leaving
		// as this for now.
		configRaw, err = getRepoFile(gh,
			os.Getenv("GITHUB_REPOSITORY"),
			configFile,
			os.Getenv("GITHUB_SHA"))

		if err != nil {
			log.Printf("Error reading configuration from default branch: %s", err)
			return
		}

	}

	config, err := getLabelerConfigV1(configRaw)
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
		owner, repo := t[0], t[1]
		l.ProcessAllPRs(owner, repo)
		l.ProcessAllIssues(owner, repo)
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

// getLabelerConfigV1 builds a LabelerConfigV1 from a raw yaml
func getLabelerConfigV1(configRaw *[]byte) (*labeler.LabelerConfigV1, error) {
	var c labeler.LabelerConfigV1
	err := yaml.Unmarshal(*configRaw, &c)
	if err != nil {
		log.Printf("Unable to unmarshall config %s: ", err)
	}
	if c.Version == 0 {
		c, err = getLabelerConfigV0(configRaw)
		if err != nil {
			log.Printf("Unable to unmarshall legacy config %s: ", err)
		}
	}
	return &c, err
}

func getLabelerConfigV0(configRaw *[]byte) (labeler.LabelerConfigV1, error) {

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

func getGithubClient() (*github.Client, error) {
	ghToken := os.Getenv("GITHUB_TOKEN")
	ghApiHost := os.Getenv("GITHUB_API_HOST")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)

	if len(ghApiHost) == 0 {
		log.Printf("Connecting to GitHub.com")
		tc := oauth2.NewClient(ctx, ts)
		return github.NewClient(tc), nil
	} else {
		log.Printf("Connecting to enterprise server at: %s", ghApiHost)
		tc := oauth2.NewClient(ctx, ts)
		return github.NewEnterpriseClient(ghApiHost, ghApiHost, tc)
	}
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
	ctx := context.Background()

	l := labeler.Labeler{

		FetchRepoConfig: func() (*labeler.LabelerConfigV1, error) {
			return config, nil
		},

		ReplaceLabels: func(target *labeler.Target, labels []string) error {
			log.Printf("Setting labels to %s/%s#%d: %s", target.Owner, target.RepoName, target.IssueNo, labels)
			_, _, err := gh.Issues.ReplaceLabelsForIssue(
				context.Background(), target.Owner, target.RepoName, target.IssueNo, labels)
			return err
		},

		GetCurrentLabels: func(target *labeler.Target) ([]string, error) {
			opts := github.ListOptions{} // TODO: ignoring pagination here
			currLabels, _, err := gh.Issues.ListLabelsByIssue(
				context.Background(), target.Owner, target.RepoName, target.IssueNo, &opts)

			labels := []string{}
			for _, label := range currLabels {
				labels = append(labels, *label.Name)
			}
			return labels, err
		},
		GitHubFacade: &labeler.GitHubFacade{
			GetRawDiff: func(owner, repo string, prNumber int) (string, error) {
				diff, _, err := gh.PullRequests.GetRaw(ctx,
					owner, repo, prNumber,
					github.RawOptions{github.Diff})
				return diff, err
			},
			ListIssuesByRepo: func(owner, repo string) ([]*github.Issue, error) {
				issues, _, err := gh.Issues.ListByRepo(ctx,
					owner, repo, &github.IssueListByRepoOptions{})
				return issues, err
			},
			ListPRs: func(owner, repo string) ([]*github.PullRequest, error) {
				prs, _, err := gh.PullRequests.List(ctx,
					owner, repo, &github.PullRequestListOptions{})
				return prs, err
			},
		},
		Client: labeler.NewDefaultHttpClient(),
	}
	return &l
}
