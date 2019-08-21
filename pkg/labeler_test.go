package labeler

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"
)

func loadPayload(name string) ([]byte, error) {
	file, err := os.Open("../test_data/" + name + "_payload")
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(file)
}

type TestCase struct {
	name           string
	config         LabelerConfig
	initialLabels  []string
	expectedLabels []string
}

func TestHandleEvent(t *testing.T) {
	payload, err := loadPayload("create_pr")
	if err != nil {
		t.Fatal(err)
	}

	// These all use the payload in the create_pr_payload file
	// referenced above
	testCases := []TestCase{
		TestCase{
			name: "Add a label when not set and config matches",
			config: LabelerConfig{
				"WIP": LabelMatcher{
					Title: "^WIP.*",
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"WIP"},
		},
		TestCase{
			name: "Remove a label when set and config does not match",
			config: LabelerConfig{
				"Fix": LabelMatcher{
					Title: "Fix: .*",
				},
			},
			initialLabels:  []string{"Fix"},
			expectedLabels: []string{},
		},
		TestCase{
			name: "Respect a label when set, and not present in config",
			config: LabelerConfig{
				"Fix": LabelMatcher{
					Title: "^Fix.*",
				},
			},
			initialLabels:  []string{"SomeLabel"},
			expectedLabels: []string{"SomeLabel"},
		},
		TestCase{
			name: "A combination of all cases",
			config: LabelerConfig{
				"WIP": LabelMatcher{
					Title: "^WIP.*",
				},
				"ShouldRemove": LabelMatcher{
					Title: "^MEH.*",
				},
			},
			initialLabels:  []string{"ShouldRemove", "ShouldRespect"},
			expectedLabels: []string{"WIP", "ShouldRespect"},
		},
	}

	for _, tc := range testCases {
		labeler := Labeler{
			fetchRepoConfig: func(owner string, repoName string) (LabelerConfig, error) {
				return tc.config, nil
			},
			getCurrentLabels: func(owner string, repoName string, prNumber int) ([]string, error) {
				return tc.initialLabels, nil
			},
			replaceLabelsForPr: func(owner string, repoName string, prNumber int, labels []string) error {
				sort.Strings(tc.expectedLabels)
				sort.Strings(labels)
				if reflect.DeepEqual(tc.expectedLabels, labels) {
					return nil
				}
				return fmt.Errorf("%s: Expecting %+v, got %+v",
					tc.name, tc.expectedLabels, labels)
			},
		}
		err = labeler.HandleEvent("pull_request", &payload)
		if err != nil {
			t.Fatal(err)
		}
	}

}
