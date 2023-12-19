package main

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	l "github.com/srvaroa/labeler/pkg"
	labeler "github.com/srvaroa/labeler/pkg"
)

func TestGetLabelerConfigV0(t *testing.T) {

	file, err := os.Open("../test_data/config_v0.yml")
	if err != nil {
		t.Fatal(err)
	}

	contents, err := ioutil.ReadAll(file)
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

	contents, err := ioutil.ReadAll(file)
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
					"cmd/.*.go",
					"pkg/.*.go",
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

	contents, err := ioutil.ReadAll(file)
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

	contents, err := ioutil.ReadAll(file)
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

	contents, err := ioutil.ReadAll(file)
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
			Files: []string{"cmd/.*.go", "pkg/.*.go"},
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
