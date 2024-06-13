package main

import (
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v50/github"
	l "github.com/srvaroa/labeler/pkg"
	labeler "github.com/srvaroa/labeler/pkg"
)

func TestGetLabelerConfigV0(t *testing.T) {

	file, err := os.Open("../test_data/config_v0.yml")
	if err != nil {
		t.Fatal(err)
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	var c *l.LabelerConfigV1
	c, err = getLabelerConfigV1(&contents)
	if err != nil {
		t.Fatal(err)
	}

	if 0 != c.Version {
		t.Fatalf("Expect version: %+v Got: %+v", 0, c.Version)
	}

	expectMatchers := map[string]l.LabelMatcher{
		"WIP": {
			Label: "WIP",
			Title: "^WIP:.*",
		},
		"WOP": {
			Label: "WOP",
			Title: "^WOP:.*",
		},
		"S": {
			Label:     "S",
			SizeBelow: "10",
		},
		"M": {
			Label:     "M",
			SizeAbove: "9",
			SizeBelow: "100",
		},
		"L": {
			Label:     "L",
			SizeAbove: "100",
		},
	}

	if !cmp.Equal(len(expectMatchers), len(c.Labels)) {
		t.Fatalf("Expect same number of matchers: %+v Got: %+v",
			len(expectMatchers),
			len(c.Labels))
	}

	for _, actualMatcher := range c.Labels {
		expectMatcher := expectMatchers[actualMatcher.Label]
		if !cmp.Equal(expectMatcher, actualMatcher) {
			t.Fatalf("Expect matcher: %+v Got: %+v",
				expectMatcher,
				actualMatcher)
		}
	}

}

func TestGetLabelerConfigV1(t *testing.T) {

	file, err := os.Open("../test_data/config_v1.yml")
	if err != nil {
		t.Fatal(err)
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	var c *l.LabelerConfigV1
	c, err = getLabelerConfigV1(&contents)
	if err != nil {
		t.Fatal(err)
	}

	expect := l.LabelerConfigV1{
		Version: 1,
		Labels: []l.LabelMatcher{
			{
				Label:  "WIP",
				Branch: "wip",
			},
			{
				Label: "WIP",
				Title: "^WIP:.*",
			},
			{
				Label: "WOP",
				Title: "^WOP:.*",
			},
			{
				Label:     "S",
				SizeBelow: "10",
			},
			{
				Label:     "M",
				SizeAbove: "9",
				SizeBelow: "100",
			},
			{
				Label:     "L",
				SizeAbove: "100",
			},
			{
				Label: "TestFileMatch",
				Files: []string{
					"cmd\\/.*.go",
					"pkg\\/.*.go",
				},
			},
			{
				Label: "Test",
				Authors: []string{
					"Test1",
					"Test2",
				},
			},
			{
				Label: "TestDraft",
				Draft: "True",
			},
			{
				Label:     "TestMergeable",
				Mergeable: "True",
			},
			{
				Label:          "TestAuthorCanMerge",
				AuthorCanMerge: "True",
			},
		},
	}

	if !cmp.Equal(expect, *c) {
		t.Fatalf("Expect: %+v Got: %+v", expect, c)
	}
}

func TestGetLabelerConfigV1WithIssues(t *testing.T) {

	file, err := os.Open("../test_data/config_v1_issues.yml")
	if err != nil {
		t.Fatal(err)
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	var c *l.LabelerConfigV1
	c, err = getLabelerConfigV1(&contents)
	if err != nil {
		t.Fatal(err)
	}

	expect := l.LabelerConfigV1{
		Version: 1,
		Issues:  true,
		Labels: []l.LabelMatcher{
			{
				Label: "Test",
				Authors: []string{
					"Test1",
					"Test2",
				},
			},
		},
	}

	if !cmp.Equal(expect, *c) {
		t.Fatalf("Expect: %+v Got: %+v", expect, c)
	}
}

func TestGetLabelerConfigV1WithCompositeSize(t *testing.T) {

	file, err := os.Open("../test_data/config_v1_composite_size.yml")
	if err != nil {
		t.Fatal(err)
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	var c *l.LabelerConfigV1
	c, err = getLabelerConfigV1(&contents)
	if err != nil {
		t.Fatal(err)
	}

	expect := l.LabelerConfigV1{
		Version: 1,
		Labels: []l.LabelMatcher{
			{
				Label:     "S",
				SizeAbove: "1",
				SizeBelow: "10",
			},
			{
				Label: "M",
				Size: &labeler.SizeConfig{
					ExcludeFiles: []string{"test.yaml"},
					Above:        "9",
					Below:        "100",
				},
			},
			{
				Label: "L",
				Size: &labeler.SizeConfig{
					ExcludeFiles: []string{"test.yaml", "\\/dir\\/test.+.yaml"},
					Above:        "100",
				},
			},
		},
	}

	if !reflect.DeepEqual(expect, *c) {
		t.Fatalf("\nExpect: %#v \nGot: %#v", expect, *c)
	}
}

func TestGetLabelerConfig2V1(t *testing.T) {

	file, err := os.Open("../test_data/config2_v1.yml")
	if err != nil {
		t.Fatal(err)
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	var c *l.LabelerConfigV1
	c, err = getLabelerConfigV1(&contents)
	if err != nil {
		t.Fatal(err)
	}

	if 1 != c.Version {
		t.Fatalf("Expect version: %+v Got: %+v", 1, c.Version)
	}

	expectMatchers := map[string]l.LabelMatcher{
		"TestLabel": {
			Label: "TestLabel",
			Title: ".*",
		},
		"TestFileMatch": {
			Label: "TestFileMatch",
			Files: []string{"cmd\\/.*.go", "pkg\\/.*.go"},
		},
		"TestTypePullRequest": {
			Label: "TestTypePullRequest",
			Type:  "pull_request",
		},
		"TestTypeIssue": {
			Label: "TestTypeIssue",
			Type:  "issue",
		},
	}

	if !cmp.Equal(len(expectMatchers), len(c.Labels)) {
		t.Fatalf("Expect same number of matchers: %+v Got: %+v",
			len(expectMatchers),
			len(c.Labels))
	}

	for _, actualMatcher := range c.Labels {
		expectMatcher := expectMatchers[actualMatcher.Label]
		if !cmp.Equal(expectMatcher, actualMatcher) {
			t.Fatalf("Expect matcher: %+v Got: %+v",
				expectMatcher,
				actualMatcher)
		}
	}

}

func TestLabelerConfigV1WithLabelSettings(t *testing.T) {

	file, err := os.Open("../test_data/config_with_label_settings_v1.yml")
	if err != nil {
		t.Fatal(err)
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	c, err := getLabelerConfigV1(&contents)
	if err != nil {
		t.Fatal(err)
	}

	if len(c.Labels) != 4 {
		t.Fatalf("configuration was not loaded properly")
	}

	expectMatchers := map[string]l.LabelMatcher{
		"TestLabel": {
			Label:       "TestLabel",
			Title:       ".*",
			Color:       "#ff0000",
			Description: "with color and description",
		},
		"TestFileMatch": {
			Label: "TestFileMatch",
			Files: []string{"cmd\\/.*.go", "pkg\\/.*.go"},
			Color: "#00ff00",
		},
		"TestTypePullRequest": {
			Label:       "TestTypePullRequest",
			Type:        "pull_request",
			Description: "without color",
		},
		"TestTypeIssue": {
			Label: "TestTypeIssue",
			Type:  "issue",
		},
	}

	if !cmp.Equal(len(expectMatchers), len(c.Labels)) {
		t.Fatalf("Expect same number of matchers: %+v Got: %+v",
			len(expectMatchers),
			len(c.Labels))
	}

	for _, actualMatcher := range c.Labels {
		expectMatcher := expectMatchers[actualMatcher.Label]
		if !cmp.Equal(expectMatcher, actualMatcher) {
			t.Fatalf("Expect matcher: %+v Got: %+v",
				expectMatcher,
				actualMatcher)
		}
	}

}

func TestGetOrDefault(t *testing.T) {
	assertEquals(t, "newValue", getOrDefault("fallback", "newValue"))
	assertEquals(t, "fallback", getOrDefault("fallback", ""))
	assertEquals(t, "", getOrDefault("", ""))
}

func TestApplyLabelConfigurationWhenAlreadyExists(t *testing.T) {
	ghLabel := &github.Label{
		Name:        github.String("TestLabel"),
		Color:       github.String("#000000"),
		Description: github.String("Old Description"),
	}

	config := l.LabelerConfigV1{
		Version: 1,
		Labels: []l.LabelMatcher{
			l.LabelMatcher{
				Label:       "TestLabel",
				Color:       "#00ff00",
				Description: "New Description",
			},
		},
	}

	ghMock := &labeler.GitHubFacade{
		ListLabels: func(owner, repo string) ([]*github.Label, error) {
			return []*github.Label{ghLabel}, nil
		},
		EditLabel: func(owner, repo string, label *github.Label) (*github.Label, error) {
			assertEquals(t, "TestLabel", *label.Name)
			assertEquals(t, "#00ff00", *label.Color)
			assertEquals(t, "New Description", *label.Description)
			return label, nil
		},
		CreateLabel: func(owner, repo string, label *github.Label) (*github.Label, error) {
			t.Fatalf("CreateLabel should not be called")
			return nil, nil
		},
	}

	err := applyLabelConfiguration(ghMock, &config, "owner", "repo")
	if err != nil {
		t.Fatalf("Error not expected %+v", err)
	}

}

func TestApplyLabelConfigurationWhenLabelDoesNotExist(t *testing.T) {

	config := l.LabelerConfigV1{
		Version: 1,
		Labels: []l.LabelMatcher{
			l.LabelMatcher{
				Label:       "TestLabel",
				Color:       "#00ff00",
				Description: "New Description",
			},
		},
	}

	ghMock := &labeler.GitHubFacade{
		ListLabels: func(owner, repo string) ([]*github.Label, error) {
			return []*github.Label{}, nil
		},
		CreateLabel: func(owner, repo string, label *github.Label) (*github.Label, error) {
			assertEquals(t, "TestLabel", *label.Name)
			assertEquals(t, "New Description", *label.Description)
			assertEquals(t, "#00ff00", *label.Color)
			return label, nil
		},
		EditLabel: func(owner, repo string, label *github.Label) (*github.Label, error) {
			t.Fatalf("EditLabel should not be called")
			return nil, nil
		},
	}

	err := applyLabelConfiguration(ghMock, &config, "owner", "repo")
	if err != nil {
		t.Fatalf("Error not expected %+v", err)
	}

}

func assertEquals(t *testing.T, expect, actual interface{}) {
	if !cmp.Equal(expect, actual) {
		t.Fatalf("Expect: %+v Got: %+v", expect, actual)
	}
}
