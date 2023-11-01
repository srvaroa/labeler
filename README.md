# Label manager for PRs and issues based on configurable conditions

[![labeler release (latest SemVer)](https://img.shields.io/github/v/release/srvaroa/labeler?sort=semver)](https://github.com/srvaroa/labeler/releases)  

Implements an all-in-one [GitHub
Action](https://help.github.com/en/categories/automating-your-workflow-with-github-actions)
that can manage multiple labels for both Pull Requests and Issues using
configurable matching rules. Available conditions:

* [Author can merge](#author-can-merge): label based on whether the author can merge the PR
* [Authors](#authors): label based on the PR/Issue authors
* [Base branch](#base-branch): label based on the PR's base branch name
* [Body](#body): label based on the PR/Issue body
* [Branch](#branch): label based on the PR's branch name
* [Draft](#draft): label based on whether the branch is mergeable
* [Files](#files): label based on the files modified in the PR
* [Mergeable](#mergeable): label based on whether the PR is mergeable
* [Size](#size): label based on the PR size
* [Title](#title): label based on the PR/Issue title

## Sponsors

Thanks to [Launchgood](https://github.com/launchgood) for sponsoring
this project.

Please consider supporting the project if your organization is finding
it useful, you can do this through [GitHub Sponsors](https://github.com/sponsors/srvaroa).

## Installing

The action is configured by adding a file `.github/labeler.yml` (which
you can override). The file contains matching rules expanded in the
`Configuration` section below.

The action will strive to maintain backwards compatibility with older
configuration versions. It is nevertheless encouraged to update your
configuration files to benefit from newer features. Please follow our
[releases](https://github.com/srvaroa/labeler/releases) page to stay up
to date.

### GitHub Enterprise support

Add `GITHUB_API_HOST` to your env variables, it should be in the form
`http(s)://[hostname]/`

Please consider [sponsoring the project](https://github.com/sponsors/srvaroa) if you're using Labeler in your organization!

### How to trigger action

To trigger the action on events, add a file `.github/workflows/main.yml`
to your repository: 

```yaml
name: Label PRs

on:
- pull_request
- issues

jobs:
  build:

    runs-on: ubuntu-latest

    steps:
    - uses: srvaroa/labeler@master
      env:
        GITHUB_TOKEN: "${{ secrets.GITHUB_TOKEN }}"
```

Using `@master` will run the latest available release. Feel free to pin
this to a specific version from the [releases
page](https://github.com/srvaroa/labeler/releases).

Use the [`on`
clause](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows)
to control when to run it.

* To trigger on PR events, [use
  `pull_request`](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request).
  to trigger on PR events and run on the merge commit of the PR. Use
  [`pull_request_target`](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request_target)
  instead if you prefer to run on the base.
* To trigger on issue events, add [`issues`](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#issues).

You may combine multiple event triggers.

A final option is to trigger the action periodically using the
[`schedule`](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#schedule)
trigger. For backwards compatibility reasons this will examine all
active pull requests and update their labels. If you wish to examine
issues as well, you'll need to explicitly add the `issues` flag in your
config file:

```yaml
version: 1
issues: True
labels:
- label: "WIP"
  title: "^WIP:.*"
```

### Advanced action settings

Please refer to the (action.yaml)[action.yaml] file in the repository
for the available inputs to the action. Below is an example using all of
them:

```yaml
name: Label PRs

on:
- pull_request
- issues

jobs:
  build:

    runs-on: ubuntu-latest

    steps:
    - uses: srvaroa/labeler@master
      with:
        config_path: .github/labeler.yml
        use_local_config: false
      env:
        GITHUB_TOKEN: "${{ secrets.GITHUB_TOKEN }}"
```

Use `config_path` to provide an alternative path for the configuration
file for the action. The default is `.github/labeler.yaml`.

Use `use_local_config` to chose where to read the config file from. By
default, the action will read the file from the default branch of your
repository. If you set `use_local_config` to `true`, then the action
will read the config file from the local checkout.

## Troubleshooting

To avoid blocking CI pipelines, the action will never return an error
code and just log information about the problem. Typical errors are
related to non-existing configuration file or invalid yaml.

If you wish to make the action fail the pipeline, you can override this
behaviour thus:

    steps:
    - uses: srvaroa/labeler@master
      with:
        fail_on_error: true

When `fail_on_error` is enabled, any failure inside the action will
exit the action process with an error code.

## Configuring matching rules

Configuration can be stored at `.github/labeler.yml` as a plain list of
label matchers, which consist of a label and a set of conditions for
each.  When *all* conditions for a label match, then the Action will set
the given label.  When *any* condition for a label does not match, then
the Action will unset the given label.

All matchers follow this configuration pattern:

```yaml
<label>: "MyLabel"
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

A PR or issue with title "WIP: this is work in progress" would be
labelled as `WIP`.  If the title changes to "This is done", then the
`WIP` label would be removed.

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

A pull request with title "WIP: this is work in progress" *and* not in a
mergeable state would be labelled as `WIP`.  If the title changes to
"This is done", or it becomes mergeable, then the `WIP` label would be
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

### Negate rules

Adding a `negate` property inside the label block will negate the
result of the evaluation of all conditions inside the label. For
example:

```yaml
version: 1
labels:
- label: "unknown"
  negate: True
  branch: "(master|hotfix)"
```

In this case, label `unknown` will be set if the branch does NOT match
`master` or `hotfix`.

The same behaviour occurs with multiple conditions:

```yaml
version: 1
labels:
- label: "unknown"
  negate: True
  branch: "master"
  title: "(feat).*"
```

Only PRs that do NOT match one of the two conditions will get the
`unknown` label.

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

A PR or issue with title 'WIP: my feature' will get the `WIP` label.

Now the title changes to `My feature` the label will get remove. This is
because the labeler configuration includes the `WIP` label, and its rule
does not match anymore.

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

Below are the conditions currently supported in label matchers, in 
alphabetical order. Note that some conditions are only applicable to
pull requests.

All conditions evaluate only when they are explicitly added in
configuration (that is, there are no default values).

### Author can merge (PRs) <a name="author-can-merge" />

This condition is satisfied when the author of the PR can merge it.
This is implemented by checking if the author is an owner of the repo.

```yaml
author-can-merge: True
```
### Authors (PRs and Issues)  <a name="authors" />

This condition is satisfied when the author of the PR or Issue matches
any of the given usernames.

```yaml
authors: ["serubin"]
```

### Base branch (PRs only) <a name="base-branch" />

This condition is satisfied when the PR base branch matches on the given
regex.

```yaml
base-branch: "master"
```

### Body (PRs and Issues) <a name="body" />

This condition is satisfied when the body (description) matches on the
given regex.

``` yaml
body: "^patch.*"
```

### Branch (PRs only) <a name="branch" />

This condition is satisfied when the PR branch matches on the given
regex.

```yaml
branch: "^feature/.*"
```

### Draft status (PRs only) <a name="draft" />

This condition is satisfied when the PR [draft
state](https://developer.github.com/v3/pulls/#response-1) matches that of the
PR.

```yaml
draft: True
```

Matches if the PR is a draft.

```yaml
draft: False
```

Matches if the PR is not a draft.

### Files affected (PRs only) <a name="files" />

This condition is satisfied when any of the PR files matches on the
given regexs.

```yaml
files: 
- "cmd/.*_tests.go"
```

### Mergeable status (PRs only) <a name="mergeable" />

This condition is satisfied when the [mergeable
state](https://developer.github.com/v3/pulls/#response-1) matches that
of the PR. 

```yaml
mergeable: True
```

Will match if the label is mergeable. 

```yaml
mergeable: False
```

Will match if the label is not mergeable. 

### Size (PRs only) <a name="size" />

This condition is satisfied when the total number of changed lines in
the PR is within given thresholds.

The number of changed lines is calculated as the sum of all `additions +
deletions` in the PR.

For example, given this `.github/labeler.yml`:

```yaml
- label: "S"
  size:
      below: 10
- label: "M"
  size:
      above: 9
      below: 100
- label: "L"
  size:
      above: 100
```

These would be the labels assigned to some PRs, based on their size as
reported by the [GitHub API](https://developer.github.com/v3/pulls).

|PR|additions|deletions|Resulting labels|
|---|---|---|---|
|First example|1|1|S|
|Second example|5|42|M|
|Third example|68|148|L|

You can exclude some files so that their changes are not taken into
account for the overall count. This can be useful for `yarn.lock`,
`go.sum` and such. Use `exclude-files`:

```yaml
- label: "L"
    size:
        exclude-files: ["yarn.lock"]
        above: 100
``` 

This condition will apply the `L` label if the diff is above 100 lines,
but NOT taking into account changes in `yarn.lock`.

**NOTICE** the old format for specifying size properties (`size-above`
and `size-below`) has been deprecated. The action will continue
supporting old configs for now, but users are encouraged to migrate to
the new configuration schema.

### Title <a name="title" />

This condition is satisfied when the title matches on the given regex.

```yaml
title: "^WIP:.*"
```
