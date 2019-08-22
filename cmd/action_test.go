package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	l "github.com/srvaroa/labeler/pkg"
)

func TestGetLabelerConfig(t *testing.T) {

	file, err := os.Open("../test_data/config.yml")
	if err != nil {
		t.Fatal(err)
	}

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	c := getLabelerConfig(&contents)
	expect := l.LabelerConfig{
		"WIP": l.LabelMatcher{
			Title: "^WIP:.*",
		},
		"WOP": l.LabelMatcher{
			Title: "^WOP:.*",
		},
	}

	if !cmp.Equal(expect, *c) {
		t.Fatalf("Expect: %+v Got: %+v", expect, c)
	}
}
