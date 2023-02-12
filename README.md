# Configurable Pull Request labels based on conditions

[![labeler release (latest SemVer)](https://img.shields.io/github/v/release/srvaroa/labeler?sort=semver)](https://github.com/srvaroa/labeler/releases)  

Implements a [GitHub
Action](https://help.github.com/en/categories/automating-your-workflow-with-github-actions)
that labels Pull Requests based on configurable conditions.

It is inspired by the example [Pull Request
Labeller](https://github.com/actions/labeler), but intends to provide a
richer set of options.

## Installing

Add a file `.github/workflows/main.yml` to your repository with these
contents:

```yaml
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
```

Then add a new file `.github/labeler.yml` with the configuration as
described below in the `Configuration` section.

If you want to run the action on the base of the pull request, rather
than on the merge commit, you should trigger the action on
`pull_request_target`.  Check the [GitHub
documentation](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#pull_request_target)
for more details.

This action will avoid failing in all cases, so if you're experiencing
unexpected behaviour it's worth looking at execution logs just in case.
Typical errors are:

* The configuration file is non existent, or has invalid yaml.
* Running the action from a fork, as the `GITHUB_TOKEN` has not enough
  permissions to label the main repository ([issue for
  solving this](https://github.com/srvaroa/labeler/issues/3))

## Configuration

Configuration can be stored at `.github/labeler.yml` as a plain list of
label matchers, which consist of a label and a set of conditions for
each.  When *all* conditions for a label match, then the Action will set
the given label.  When *any* condition for a label does not match, then
the Action will unset the given label.

Here is an example of a matcher for label "Example":

```yaml
<label>: "Example"
<condition_name>: <condition_parameters>
<condition_name>: <condition_parameters>
```

For example, this `.github/labeler.yml` contains a single matcher with
a single condition:

```yaml
version: 1
labels:
- label: "WIP"
  title: "^WIP:.*"
```

A Pull Request with title "WIP: this is work in progress" would be labelled as
`WIP`.  If the Pull Request title changes to "This is done", then the `WIP`
label would be removed.

Each label may combine multiple conditions.  The action combines all
conditions with an AND operation.  That is, the label will be applied if
*all* conditions are satisfied, removed otherwise.

For example, given this `.github/labeler.yml`:

```yaml
version: 1
labels:
- label: "WIP"
  title: "^WIP:.*"
  mergeable: false
```

A Pull Request with title "WIP: this is work in progress" *and* not in a
mergeable state would be labelled as `WIP`.  If the Pull Request title changes
to "This is done", or it becomes mergeable, then the `WIP` label would be
removed.

If you wish to apply an OR, you may set multiple matchers for the same
label. For example:

```yaml
version: 1
labels:
- label: "WIP"
  title: "^WIP:.*"
- label: "WIP"
  mergeable: false
```

The `WIP` label will be set if the title matches `^WIP:.*` OR the label
is not in a mergeable state.

## Append-only mode

The default behaviour of this action includes *removing* labels that
have a rule configured that does not match anymore. For example, given
this configuration:

```yaml
version: 1
labels:
- label: "WIP"
  title: "^WIP:.*"
```

A PR with title 'WIP: my feature' will get the `WIP` label.

Now the title changes to `My feature`. Since the labeler configuration
includes the `WIP` label, and its rule does not match anymore, the label
will get removed.

In some cases you would prefer that the action adds labels, but never
removes them regardless of the matching status. To achieve this you can
enable the `appendOnly` flag.

```yaml
version: 1
appendOnly: true
labels:
- label: "WIP"
  title: "^WIP:.*"
```

With this config, the behaviour changes:

- A PR with title 'WIP: my feature' will get the `WIP` label.
- When the title changes to `My feature`, even though the labeler has a
  rule for the `WIP` label that does not match, the label will be
  respected.

## Conditions

Below are the conditions currently supported in label matchers. All conditions
evaluate only when they are explicitly added in configuration (that is, there
are no default values).

### Regex on title

This condition is satisfied when the PR title matches on the given regex.

```yaml
title: "^WIP:.*"
```

### Regex on branch

This condition is satisfied when the PR branch matches on the given regex.

```yaml
branch: "^feature/.*"
```

### Regex on base branch

This condition is satisfied when the PR base branch matches on the given regex.

```yaml
base-branch: "master"
```

### Regex on PR body 

This condition is satisfied when the body (description) matches on the given regex.

``` yaml
body: "^patch.*"
```

### Regex on PR files

This condition is satisfied when any of the PR files matches on the given regexs.

```yaml
files: 
- "cmd/.*_tests.go"
```

### Draft status

This condition is satisfied when the PR [draft
state](https://developer.github.com/v3/pulls/#response-1) matches that of the
PR.

```yaml
draft: true
```

Matches if the PR is a draft.

```yaml
draft: false
```

Matches if the PR is not a draft.

### Mergeable status

This condition is satisfied when the [mergeable
state](https://developer.github.com/v3/pulls/#response-1) matches that of the
PR. 

```yaml
mergeable: true
```

Will match if the label is mergeable. 

```yaml
mergeable: false
```

Will match if the label is not mergeable. 

### Match to PR Author

This condition is satisfied when the PR author matches any of the given usernames.

```yaml
author: "serubin"
```

### PR size

This condition is satisfied when the total number of changed lines in
the PR is within given thresholds.

The number of changed lines is calculated as the sum of all `additions +
deletions` in the PR.

For example, given this `.github/labeler.yml`:

```yaml
- label: "S"
  size-below: 10
- label: "M"
  size-above: 9
  size-below: 100
- label: "L"
  size-above: 100
```

These would be the labels assigned to some PRs, based on their size as
reported by the [GitHub API](https://developer.github.com/v3/pulls).

|PR|additions|deletions|Resulting labels|
|---|---|---|---|
|First example|1|1|S|
|Second example|5|42|M|
|Third example|68|148|L|
