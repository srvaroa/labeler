package labeler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	gh "github.com/google/go-github/v50/github"
	"github.com/waigani/diffparser"
)

func FilesCondition(l *Labeler) Condition {
	prFiles := []string{}

	return Condition{
		GetName: func() string {
			return "File matches regex"
		},
		CanEvaluate: func(target *Target) bool {
			return target.ghPR != nil
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {

			if len(matcher.Files) <= 0 {
				return false, fmt.Errorf("Files are not set in config")
			}

			if len(prFiles) == 0 {
				var err error
				prFiles, err = l.getPrFileNames(target.ghPR)
				if err != nil {
					return false, err
				}
			}

			log.Printf("Matching `%s` against: %s", strings.Join(matcher.Files, ", "), strings.Join(prFiles, ", "))
			for _, fileMatcher := range matcher.Files {
				for _, prFile := range prFiles {
					isMatched, _ := regexp.Match(fileMatcher, []byte(prFile))
					if isMatched {
						log.Printf("Matched `%s` against: `%s`", prFile, fileMatcher)
						return isMatched, nil
					}
				}
			}
			return false, nil
		},
	}
}

// getPrFileNames returns all of the file names (old and new) of files changed in the given PR
func (l *Labeler) getPrFileNames(pr *gh.PullRequest) ([]string, error) {
	log.Printf("getPrFileNames for pr - %s", pr.GetURL())
	ghToken := os.Getenv("GITHUB_TOKEN")
	diffReq, err := http.NewRequest("GET", pr.GetURL(), nil)

	if err != nil {
		return nil, err
	}

	if ghToken != "" {
		diffReq.Header.Add("Authorization", "Bearer "+ghToken)
	} else {
		log.Printf("Env var GITHUB_TOKEN is missing, using annonymous request")
	}
	diffReq.Header.Add("Accept", "application/vnd.github.v3.diff")
	diffRes, err := l.Client.Do(diffReq)

	if err != nil {
		return nil, err
	}

	defer diffRes.Body.Close()

	var diffRaw []byte
	prFiles := make([]string, 0)
	if diffRes.StatusCode == http.StatusOK {
		diffRaw, err = ioutil.ReadAll(diffRes.Body)
		if err != nil {
			return nil, err
		}

		diff, err := diffparser.Parse(string(diffRaw))
		if err != nil {
			return nil, err
		}

		log.Printf("got diff %s, parsed %+v", string(diffRaw), diff)
		prFilesSet := map[string]struct{}{}
		// Place in a set to remove duplicates
		for _, file := range diff.Files {
			prFilesSet[file.OrigName] = struct{}{}
			prFilesSet[file.NewName] = struct{}{}
		}
		// Convert to list to make it easier to consume
		for k := range prFilesSet {
			prFiles = append(prFiles, k)
		}
		log.Printf("diff files %s", prFiles)
	} else {
		log.Printf("failed with status %s", diffRes.Status)
	}

	return prFiles, nil
}
