package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
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
			{
				Label:        "TestIsAuthorInTeam",
				AuthorInTeam: "team1",
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

func TestIsUserMemberOfTeam_404(t *testing.T) {
	// Simulate the GitHub API returning 404 when the token lacks
	// read:org scope. Verify that newLabeler's IsUserMemberOfTeam
	// produces an actionable error message mentioning permissions.
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not Found"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	ghClient, err := github.NewEnterpriseClient(
		server.URL+"/",
		server.URL+"/",
		server.Client(),
	)
	if err != nil {
		t.Fatal(err)
	}

	config := &labeler.LabelerConfigV1{Version: 1}
	labelerInstance := newLabeler(ghClient, config)

	isMember, err := labelerInstance.GitHubFacade.IsUserMemberOfTeam(
		"myorg", "someuser", "myteam")

	if isMember {
		t.Error("Expected isMember to be false on 404")
	}
	if err == nil {
		t.Fatal("Expected an error on 404")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("Expected error to mention HTTP 404, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "read:org") {
		t.Errorf("Expected error to mention read:org scope, got: %s", err.Error())
	}
}

func TestIsUserMemberOfTeam_ActiveMember(t *testing.T) {
	// Simulate a successful membership check returning active state.
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"state":"active","role":"member"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	ghClient, err := github.NewEnterpriseClient(
		server.URL+"/",
		server.URL+"/",
		server.Client(),
	)
	if err != nil {
		t.Fatal(err)
	}

	config := &labeler.LabelerConfigV1{Version: 1}
	labelerInstance := newLabeler(ghClient, config)

	isMember, err := labelerInstance.GitHubFacade.IsUserMemberOfTeam(
		"myorg", "someuser", "myteam")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !isMember {
		t.Error("Expected isMember to be true for active member")
	}
}

func TestIsUserMemberOfTeam_500(t *testing.T) {
	// Verify that non-404 errors are returned as-is without the
	// permissions guidance message.
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Internal Server Error"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	ghClient, err := github.NewEnterpriseClient(
		server.URL+"/",
		server.URL+"/",
		server.Client(),
	)
	if err != nil {
		t.Fatal(err)
	}

	config := &labeler.LabelerConfigV1{Version: 1}
	labelerInstance := newLabeler(ghClient, config)

	isMember, err := labelerInstance.GitHubFacade.IsUserMemberOfTeam(
		"myorg", "someuser", "myteam")

	if isMember {
		t.Error("Expected isMember to be false on 500")
	}
	if err == nil {
		t.Fatal("Expected an error on 500")
	}
	if strings.Contains(err.Error(), "read:org") {
		t.Errorf("Non-404 errors should not mention read:org, got: %s", err.Error())
	}
}
