# Condition based Pull Request Labeler

Implements a [GitHub
Action](https://help.github.com/en/categories/automating-your-workflow-with-github-actions)
that labels Pull Requests based on configurable conditions.

It is inspired by the example [Pull Request
Labeller](https://github.com/actions/labeler), but intends to provide a
richer set of options.

## Installing

Add a .github/workflows/main.yml file to your repository with these
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

Then, add a ./github/labeler.yml with the configuration as described
below.

## Configuration

Configuration can be stored at `./github/labeler.yml` as a plain list of labels
and a set of conditions for each.  When *all* conditions for a label match,
then the Action will set the given label.  When *any* condition for a label
does not match, then the Action will unset the given label.

    <label>:
      <condition_name>: <condition_parameters>
      <condition_name>: <condition_parameters>

For example, given this `./github/labeler.yml`:

      WIP:
        title: "^WIP:.*"

A Pull Request with title "WIP: this is work in progress" would be labelled as
`WIP`.  If the Pull Request title changes to "This is done", then the `WIP`
label would be removed.

Each label may combine multiple conditions.  The label will be applied if *all*
conditions are satisfied, removed otherwise.

For example, given this `./github/labeler.yml`:

      WIP:
        title: "^WIP:.*"
        mergeable: false

A Pull Request with title "WIP: this is work in progress" *and* not in a
mergeable state would be labelled as `WIP`.  If the Pull Request title changes
to "This is done", or it becomes mergeable, then the `WIP` label would be
removed.

## Conditions

Below are the conditions currently supported.

### Regex on title

This condition is satisfied when the PR title matches on the given regex.

    WIP:
      title: "^WIP:.*"

### Mergeable status

This condition is satisfied when the PR is in a [mergeable state](https://developer.github.com/v3/pulls/#response-1).

    MyLabel:
      mergeable: true

### PR size

This condition is satisfied when the total number of changed lines in
the PR is within given thresholds.

The number of changed lines is calculated as the sum of all `additions +
deletions` in the PR.

For example, given this `./github/labeler.yml`:

    S:
      size-below: 10
    M:
      size-above: 9
      size-below: 100
    L:
      size-above: 100

These would be the labels assigned to some PRs, based on their size as
reported by the [GitHub API](https://developer.github.com/v3/pulls).

|PR|additions|deletions|Resulting labels|
|---|---|---|---|
|First example|1|1|S|
|Second example|5|42|M|
|Third example|68|148|L|
