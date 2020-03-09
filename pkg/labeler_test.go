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
	payloads       []string
	name           string
	config         LabelerConfig
	initialLabels  []string
	expectedLabels []string
}

func TestHandleEvent(t *testing.T) {

	// These all use the payload in payload files
	testCases := []TestCase{
		TestCase{
			payloads:       []string{"create_pr", "reopen_pr"},
			name:           "Empty config",
			config:         LabelerConfig{},
			initialLabels:  []string{"Fix"},
			expectedLabels: []string{"Fix"},
		},
		TestCase{
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Config with no rules",
			config: LabelerConfig{
				"WIP": LabelMatcher{},
			},
			initialLabels:  []string{"Fix"},
			expectedLabels: []string{"Fix"},
		},
		TestCase{
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Add a label when not set and config matches",
			config: LabelerConfig{
				"WIP": LabelMatcher{
					Title: "^WIP:.*",
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"WIP"},
		},
		TestCase{
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Remove a label when set and config does not match",
			config: LabelerConfig{
				"Fix": LabelMatcher{
					Title: "Fix: .*",
				},
			},
			initialLabels:  []string{"Fix"},
			expectedLabels: []string{},
		},
		TestCase{
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Respect a label when set, and not present in config",
			config: LabelerConfig{
				"Fix": LabelMatcher{
					Title: "^Fix.*",
				},
			},
			initialLabels:  []string{"SomeLabel"},
			expectedLabels: []string{"SomeLabel"},
		},
		TestCase{
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "A combination of all cases",
			config: LabelerConfig{
				"WIP": LabelMatcher{
					Title: "^WIP:.*",
				},
				"ShouldRemove": LabelMatcher{
					Title: "^MEH.*",
				},
			},
			initialLabels:  []string{"ShouldRemove", "ShouldRespect"},
			expectedLabels: []string{"WIP", "ShouldRespect"},
		},
		TestCase{
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Add a label with two conditions, both matching",
			config: LabelerConfig{
				"WIP": LabelMatcher{
					Title:     "^WIP:.*",
					Mergeable: "False",
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"WIP"},
		},
		TestCase{
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Add a label with two conditions, one not matching",
			config: LabelerConfig{
				"WIP": LabelMatcher{
					Title:     "^WIP:.*",
					Mergeable: "True",
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{},
		},
		TestCase{
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Add a label with two conditions, one not matching",
			config: LabelerConfig{
				"WIP": LabelMatcher{
					Title: "^((?!WIP).)*$",
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{},
		},
		TestCase{
			payloads: []string{"small_pr"},
			name:     "Test the size_below rule",
			config: LabelerConfig{
				"S": LabelMatcher{
					SizeBelow: "10",
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"S"},
		},
		TestCase{
			payloads: []string{"mid_pr"},
			name:     "Test the size_below and size_above rules",
			config: LabelerConfig{
				"M": LabelMatcher{
					SizeAbove: "9",
					SizeBelow: "100",
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"M"},
		},
		TestCase{
			payloads: []string{"big_pr"},
			name:     "Test the size_above rule",
			config: LabelerConfig{
				"L": LabelMatcher{
					SizeAbove: "100",
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"L"},
		},
	}

	for _, tc := range testCases {
		for _, file := range tc.payloads {
			payload, err := loadPayload(file)
			if err != nil {
				t.Fatal(err)
			}

			fmt.Println(tc.name)
			l := NewTestLabeler(t, tc)
			err = l.HandleEvent("pull_request", &payload)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func NewTestLabeler(t *testing.T, tc TestCase) Labeler {
	return Labeler{
		FetchRepoConfig: func(owner, repoName string) (*LabelerConfig, error) {
			return &tc.config, nil
		},
		GetCurrentLabels: func(owner, repoName string, prNumber int) ([]string, error) {
			return tc.initialLabels, nil
		},
		ReplaceLabelsForPr: func(owner, repoName string, prNumber int, labels []string) error {
			sort.Strings(tc.expectedLabels)
			sort.Strings(labels)
			if reflect.DeepEqual(tc.expectedLabels, labels) {
				return nil
			}
			return fmt.Errorf("%s: Expecting %+v, got %+v",
				tc.name, tc.expectedLabels, labels)
		},
	}
}
