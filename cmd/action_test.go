package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	l "github.com/srvaroa/labeler/pkg"
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
	c, err = getLabelerConfig(&contents)
	if err != nil {
		t.Fatal(err)
	}

	if 0 != c.Version {
		t.Fatalf("Expect version: %+v Got: %+v", 0, c.Version)
	}

	expectMatchers := map[string]l.LabelMatcher{
		"WIP": l.LabelMatcher{
			Label: "WIP",
			Title: "^WIP:.*",
		},
		"WOP": l.LabelMatcher{
			Label: "WOP",
			Title: "^WOP:.*",
		},
		"S": l.LabelMatcher{
			Label:     "S",
			SizeBelow: "10",
		},
		"M": l.LabelMatcher{
			Label:     "M",
			SizeAbove: "9",
			SizeBelow: "100",
		},
		"L": l.LabelMatcher{
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
	c, err = getLabelerConfig(&contents)
	if err != nil {
		t.Fatal(err)
	}

	expect := l.LabelerConfigV1{
		Version: 1,
		Labels: []l.LabelMatcher{
			l.LabelMatcher{
				Label:  "WIP",
				Branch: "wip",
			},
			l.LabelMatcher{
				Label: "WIP",
				Title: "^WIP:.*",
			},
			l.LabelMatcher{
				Label: "WOP",
				Title: "^WOP:.*",
			},
			l.LabelMatcher{
				Label:     "S",
				SizeBelow: "10",
			},
			l.LabelMatcher{
				Label:     "M",
				SizeAbove: "9",
				SizeBelow: "100",
			},
			l.LabelMatcher{
				Label:     "L",
				SizeAbove: "100",
			},
		},
	}

	if !cmp.Equal(expect, *c) {
		t.Fatalf("Expect: %+v Got: %+v", expect, c)
	}
}
