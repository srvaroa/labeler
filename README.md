# Condition based Pull Request Labeler

Implements a [GitHub
Action](https://help.github.com/en/categories/automating-your-workflow-with-github-actions)
that labels Pull Requests based on configurable conditions.

It is inspired by the example [Pull Request
Labeller](https://github.com/actions/labeler), but intends to provide a
richer set of options.

## Installing

Add a file `.github/workflows/main.yml` to your repository with these
contents:

	name: Label PRs

	on:
	  - pull_request

	jobs:
	  build:

		runs-on: ubuntu-latest
		
		steps:
		- uses: srvaroa/labeler@master
		  env:
			GITHUB_TOKEN: "${{ secrets.GITHUB_TOKEN }}"

Then add a new file `./github/labeler.yml` with the configuration as
described below in the `Configuration` section.

This action will avoid failing in all cases, so if you're experiencing
unexpected behaviour it's worth looking at execution logs just in case.
Typical errors are:

* The configuration file is non existent, or has invalid yaml.
* Running the action from a fork, as the `GITHUB_TOKEN` has not enough
  permissions to label the main repository ([issue for
  solving this](https://github.com/srvaroa/labeler/issues/3))

## Configuration

Configuration can be stored at `./github/labeler.yml` as a plain list of
label matchers, which consist of a label and a set of conditions for
each.  When *all* conditions for a label match, then the Action will set
the given label.  When *any* condition for a label does not match, then
the Action will unset the given label.

Here is an example of a matcher for label "Example":

      <label>: "Example"
      <condition_name>: <condition_parameters>
      <condition_name>: <condition_parameters>

For example, this `./github/labeler.yml` contains a single matcher with
a single condition:

    version: 1
    labels:
      - label: "WIP"
	title: "^WIP:.*"

A Pull Request with title "WIP: this is work in progress" would be labelled as
`WIP`.  If the Pull Request title changes to "This is done", then the `WIP`
label would be removed.

Each label may combine multiple conditions.  The action combines all
conditions with an AND operation.  That is, the label will be applied if
*all* conditions are satisfied, removed otherwise.

For example, given this `./github/labeler.yml`:

    version: 1
    labels:
      - label: "WIP"
	title: "^WIP:.*"
        mergeable: false

A Pull Request with title "WIP: this is work in progress" *and* not in a
mergeable state would be labelled as `WIP`.  If the Pull Request title changes
to "This is done", or it becomes mergeable, then the `WIP` label would be
removed.

If you wish to apply an OR, you may set multiple matchers for the same
label. For example:

    version: 1
    labels:
      - label: "WIP"
	title: "^WIP:.*"
      - label: "WIP"
        mergeable: false

The `WIP` label will be set if the title matches `^WIP:.*` OR the label
is not in a mergeable state.

## Conditions

Below are the conditions currently supported in label matchers.

### Regex on title

This condition is satisfied when the PR title matches on the given regex.

    title: "^WIP:.*"

### Regex on branch

This condition is satisfied when the PR branch matches on the given regex.

    branch: "^feature/.*"

### Regex on PR files

This condition is satisfied when any of the PR files matches on the given regexs.

    files: 
      - "cmd/.*_tests.go"

### Mergeable status

This condition is satisfied when the PR is in a [mergeable state](https://developer.github.com/v3/pulls/#response-1).

    mergeable: true

### PR size

This condition is satisfied when the total number of changed lines in
the PR is within given thresholds.

The number of changed lines is calculated as the sum of all `additions +
deletions` in the PR.

For example, given this `./github/labeler.yml`:

    - label: "S"
      size-below: 10
    - label: "M":
      size-above: 9
      size-below: 100
    - label: "L":
      size-above: 100

These would be the labels assigned to some PRs, based on their size as
reported by the [GitHub API](https://developer.github.com/v3/pulls).

|PR|additions|deletions|Resulting labels|
|---|---|---|---|
|First example|1|1|S|
|Second example|5|42|M|
|Third example|68|148|L|
