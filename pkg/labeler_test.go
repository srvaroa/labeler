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
			payloads:       []string{"create_pr", "reopen_pr"},
			name:           "Empty config",
			config:         LabelerConfigV1{},
			initialLabels:  []string{"Fix"},
			expectedLabels: []string{"Fix"},
		},
		{
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
			payloads: []string{"small_pr"},
			name:     "Test the branch rule (matching)",
			config: LabelerConfigV1{
				Version: 1,
				Labels: []LabelMatcher{
					{
						Label:  "Branch",
						Branch: "^(?!^feature/.*$)(?!^bugfix/.*$)(?!^enhance/.*$)(?!^style/.*$)(?!^docs/.*$).*$",
					},
				},
			},
			initialLabels:  []string{},
			expectedLabels: []string{"Branch"},
		},
		{
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
		})
	}
}

func NewTestLabeler(t *testing.T, tc TestCase) Labeler {
	return Labeler{
		FetchRepoConfig: func(owner, repoName string) (*LabelerConfigV1, error) {
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
