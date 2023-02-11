package labeler

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
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
	event          string // issues or pull_request
	payloads       []string
	name           string
	config         LabelerConfigV1
	initialLabels  []string
	expectedLabels []string
}

func TestHandleEvent(t *testing.T) {

	// These all use the payload in payload files
	testCases := []TestCase{
		{
			event:          "pull_request",
			payloads:       []string{"create_pr", "reopen_pr"},
			name:           "Empty config",
			config:         LabelerConfigV1{},
			initialLabels:  []string{"Fix"},
			expectedLabels: []string{"Fix"},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Config with no rules",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "WIP",
					},
				},
			},
			initialLabels:  []string{"Fix"},
			expectedLabels: []string{"Fix"},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Add a label when not set and config matches",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "WIP",
						Title: "^WIP:.*",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"WIP"},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Remove a label when set and config does not match",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "Fix",
						Title: "Fix: .*",
					},
				},
			},
			initialLabels:  []string{"Fix"},
			expectedLabels: []string{},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Respect a label when set, and not present in config",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "Fix",
						Title: "^Fix.*",
					},
				},
			},
			initialLabels:  []string{"SomeLabel"},
			expectedLabels: []string{"SomeLabel"},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "A combination of all cases",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "WIP",
						Title: "^WIP:.*",
					},
					{
						Label: "ShouldRemove",
						Title: "^MEH.*",
					},
				},
			},
			initialLabels:  []string{"ShouldRemove", "ShouldRespect"},
			expectedLabels: []string{"WIP", "ShouldRespect"},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Add a label with two conditions, both matching",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:     "WIP",
						Title:     "^WIP:.*",
						Mergeable: "False",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"WIP"},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr", "create_draft_pr"},
			name:     "Draft PR without explicit value is not evaluated",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "NotADraft",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr"},
			name:     "Non draft PR with explicit config is evaluated",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "NotADraft",
						Draft: "False",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"NotADraft"},
		},

		{
			event:    "pull_request",
			payloads: []string{"create_draft_pr"},
			name:     "Draft PR",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "ThisIsADraft",
						Draft: "True",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"ThisIsADraft"},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Add a label with two conditions, one not matching (1)",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:     "WIP",
						Title:     "^WIP:.*",
						Mergeable: "True",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{},
		},
		{
			// covers evaluation order making a True in the last
			// condition, while previous ones are false
			event:    "pull_request",
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "Add a label with two conditions, one not matching (2)",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:     "WIP",
						Title:     "^DOES NOT MATCH:.*",
						Mergeable: "False",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{},
		},
		{
			event:    "pull_request",
			payloads: []string{"small_pr"},
			name:     "Test the size_below rule",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:     "S",
						SizeBelow: "10",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"S"},
		},
		{
			event:    "pull_request",
			payloads: []string{"mid_pr"},
			name:     "Test the size_below and size_above rules",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:     "M",
						SizeAbove: "9",
						SizeBelow: "100",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"M"},
		},
		{
			event:    "pull_request",
			payloads: []string{"big_pr"},
			name:     "Test the size_above rule",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:     "L",
						SizeAbove: "100",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"L"},
		},
		{
			event:    "pull_request",
			payloads: []string{"small_pr"},
			name:     "Test the branch rule (matching)",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:  "Branch",
						Branch: "^srvaroa-patch.*",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"Branch"},
		},
		{
			event:    "pull_request",
			payloads: []string{"small_pr"},
			name:     "Test the branch rule (not matching)",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:  "Branch",
						Branch: "^does/not-match/*",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr"},
			name:     "Test the base branch rule (matching)",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:      "Branch",
						BaseBranch: "^master",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"Branch"},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr"},
			name:     "Test the base branch rule (not matching)",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:      "Branch",
						BaseBranch: "^does/not-match/*",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr"},
			name:     "Test the body rule (matching)",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "Body",
						Body:  "^Signed-off.*",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"Body"},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr"},
			name:     "Test the body rule (not matching)",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "Body",
						Body:  "/patch/",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{},
		},
		{
			event:    "pull_request",
			payloads: []string{"diff_pr"},
			name:     "Test the files rule",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "Files",
						Files: []string{
							"^.*.md",
						},
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"Files"},
		},
		{
			event:    "pull_request",
			payloads: []string{"small_pr"},
			name:     "Multiple conditions for the same tag function as OR",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:  "Branch",
						Branch: "^srvaroa-patch.*",
					},
					{
						Label:  "Branch",
						Branch: "WONT MATCH",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"Branch"},
		},
		{
			event:    "pull_request",
			payloads: []string{"small_pr"},
			name:     "Multiple conditions for the same tag function as OR",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:   "Author",
						Authors: []string{"srvaroa"},
					},
					{
						Label:  "Branch",
						Branch: "WONT MATCH",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"Author"},
		},
		{
			event:    "pull_request",
			payloads: []string{"create_pr", "reopen_pr"},
			name:     "AppendOnly enabled forbids deletions",
			config: LabelerConfigV1{
				Version:    1,
				AppendOnly: true,
				Labels: []LabelMatcher{
					{
						Label: "Fix",
						Title: "THIS DOES NOT MATCH",
					},
				},
			},
			initialLabels: []string{"Fix"},
			// We have a rule for label Fix, it does not match
			// BUT because AppendOnly is set, we do not erase it
			expectedLabels: []string{"Fix"},
		},

		// Issues

		{
			event:    "issues",
			payloads: []string{"issue_open"},
			name:     "Add a label to issue when title matches",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "Test",
						Title: "^Testy.*t",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"Test"},
		},
		{
			event:    "issues",
			payloads: []string{"issue_open"},
			name:     "Remove a label from issue when title does not match",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "Test",
						Title: "Wontmatch",
					},
				},
			},
			initialLabels:  []string{"Test", "Meh"},
			expectedLabels: []string{"Meh"},
		},
		{
			event:    "issues",
			payloads: []string{"issue_open"},
			name:     "Add label to issue when author matches",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:   "Test",
						Authors: []string{"spiderman", "srvaroa"},
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"Test"},
		},
		{
			event:    "issues",
			payloads: []string{"issue_open"},
			name:     "Remove label from issue when author does not match",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:   "Test",
						Authors: []string{"spiderman"},
					},
				},
			},
			initialLabels:  []string{"Test", "Meh"},
			expectedLabels: []string{"Meh"},
		},
		{
			event:    "issues",
			payloads: []string{"issue_open"},
			name:     "Add label to issue when body matches",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "Test",
						Body:  ".+ descr.+on!$",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"Test"},
		},
		{
			event:    "issues",
			payloads: []string{"issue_open"},
			name:     "Remove label from issue when body does not match",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label: "Test",
						Body:  "will_not_match",
					},
				},
			},
			initialLabels:  []string{"Test", "Meh"},
			expectedLabels: []string{"Meh"},
		},
		{
			event:    "issues",
			payloads: []string{"issue_open"},
			name:     "Leave tags untouched when rule not supported in issues",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:     "ShouldNotAppear1",
						Mergeable: "False",
					},
					{
						Label:     "ShouldNotAppear2",
						SizeBelow: "10",
						SizeAbove: "0",
					},
					{
						Label: "ShouldNotAppear3",
						Files: []string{
							"^.*.md",
						},
					},
					{
						Label:  "ShouldNotAppear4",
						Branch: "master",
					},
					{
						Label:      "ShouldNotAppear5",
						BaseBranch: "master",
					},
				},
			},
			initialLabels:  []string{"Test", "WIP"},
			expectedLabels: []string{"Test", "WIP"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, file := range tc.payloads {
				payload, err := loadPayload(file)
				if err != nil {
					fmt.Printf("Test `%s`: failed to load %s\n", tc.name, file)
					t.Fatal(err)
				}

				fmt.Printf("--> TEST: %s \n", tc.name)
				l := NewTestLabeler(t, tc)
				err = l.HandleEvent(tc.event, &payload)
				if err != nil {
					fmt.Printf("Test failed: %s\n", tc.name)
					t.Fatal(err)
				}
			}
		})
	}
}

func NewTestLabeler(t *testing.T, tc TestCase) Labeler {
	return Labeler{
		FetchRepoConfig: func() (*LabelerConfigV1, error) {
			return &tc.config, nil
		},
		GetCurrentLabels: func(target *Target) ([]string, error) {
			return tc.initialLabels, nil
		},
		ReplaceLabels: func(target *Target, labels []string) error {
			sort.Strings(tc.expectedLabels)
			sort.Strings(labels)
			if reflect.DeepEqual(tc.expectedLabels, labels) {
				return nil
			}
			return fmt.Errorf("%s: Expecting %+v, got %+v",
				tc.name, tc.expectedLabels, labels)
		},
		Client: &FakeHttpClient{},
	}
}

type FakeHttpClient struct {
}

func (f *FakeHttpClient) Do(req *http.Request) (*http.Response, error) {
	file, err := os.Open("../test_data/diff_response")
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	diffReader := strings.NewReader(string(data))
	diffReadCloser := io.NopCloser(diffReader)

	response := http.Response{
		StatusCode: http.StatusOK,
		Body:       diffReadCloser,
	}
	return &response, nil
}
