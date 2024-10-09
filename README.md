# Label manager for PRs and issues based on configurable conditions

[![labeler release (latest SemVer)](https://img.shields.io/github/v/release/srvaroa/labeler?sort=semver)](https://github.com/srvaroa/labeler/releases)  [![sponsor the project!](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/srvaroa)

Implements an all-in-one [GitHub
Action](https://help.github.com/en/categories/automating-your-workflow-with-github-actions)
that can manage multiple labels for both Pull Requests and Issues using
configurable matching rules. Available conditions:

* [Age](#age): label based on the age of a PR or Issue
* [Author can merge](#author-can-merge): label based on whether the author can merge the PR
* [Author is member of team](#author-in-team): label based on whether the author is an active member of the given team
* [Authors](#authors): label based on the PR/Issue authors
* [Base branch](#base-branch): label based on the PR's base branch name
* [Body](#body): label based on the PR/Issue body
* [Branch](#branch): label based on the PR's branch name
* [Draft](#draft): label based on whether the PR is a draft
* [Files](#files): label based on the files modified in the PR
* [Last modified](#last-modified): label based on the last modification to a PR or Issue
* [Mergeable](#mergeable): label based on whether the PR is mergeable
* [Size](#size): label based on the PR size, allowing file exclusions
* [Title](#title): label based on the PR/Issue title
* [Type](#type): label based on record type (PR or Issue)

## Sponsors

Please consider supporting the project if your organization finds it useful,
you can do this through [GitHub Sponsors](https://github.com/sponsors/srvaroa).
Sponsorships also help speed up bug fixes or new features.

Thanks to [Launchgood](https://github.com/launchgood) and others that
preferred to remain private for supporting this project!

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
page](https://github.com/srvaroa/labeler/releases). We also maintain a
floating tag on the major `v1`. This gets updated whenever a new
minor/patch v1.x.y version is released.

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

<a name="schedule" />A final option is to trigger the action periodically using the
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

Please refer to the [action.yml](action.yml) file in the repository
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

    - name: Checkout your code
      uses: actions/checkout@v3

    - uses: srvaroa/labeler@master
      with:
        config_path: .github/labeler.yml
        use_local_config: false
        fail_on_error: false
      env:
        GITHUB_TOKEN: "${{ secrets.GITHUB_TOKEN }}"
```

Use `config_path` to provide an alternative path for the configuration
file for the action. The default is `.github/labeler.yml`.

Use `use_local_config` to chose where to read the config file from. By
default, the action will read the file from the default branch of your
repository. If you set `use_local_config` to `true`, then the action
will read the config file from the local checkout. Note that you may
need to checkout your branch before the action runs!

Use `fail_on_error` to decide whether an error in the action execution
should trigger a failure of the workflow. By default it's disabled to
prevent the action from disrupting CI pipelines.

## Troubleshooting

To avoid blocking CI pipelines, the action will never return an error
code and just log information about the problem. Typical errors are
related to non-existing configuration file or invalid yaml.

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
alphabetical order. Some important considerations:

* Conditions evaluate only when they are explicitly added in
  configuration. There are no defaults.
* Some conditions are only applicable to pull requests.
* All conditions based on regex rely on [Go's `regexp`
  package](https://pkg.go.dev/regexp), which accepts the syntax accepted
  by RE2 and described at [golang.org](https://golang.org/s/re2syntax).
  You can use tools like [regex101.com](https://regex101.com/?flavor=golang)
  to verify your conditions.

### Age (PRs and Issues) <a name="age" />

This condition is satisfied when the age of the PR or Issue are larger than
the given one. The age is calculated from the creation date.

If you're looking to evaluate on the modification date of the issue or PR, 
check on <a href="#last-modified" ></a>

This condition is best used when with a <a href="#schedule">schedule trigger</a>.

Example:

```yaml
age: 1d
```

The syntax for values is based on a number, followed by a suffix:

* s: seconds
* m: minutes
* h: hours
* d: days
* w: weeks
* y: years

For example, `2d` means 2 days, `4w` means 4 weeks, and so on.

### Author can merge (PRs) <a name="author-can-merge" />

This condition is satisfied when the author of the PR can merge it.
This is implemented by checking if the author is an owner of the repo.

```yaml
author-can-merge: True
```


### Author is member (PRs and Issues) <a name="author-in-team" />

This condition is satisfied when the author of the PR is an active
member of the given team (identified by its url slug).

```yaml
author-in-team: core-team
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
- "cmd\\/.*_tests.go"
- ".*\\/subfolder\\/.*\\.md"
```

> **NOTICE** the double backslash (`\\`) in the example above. This GitHub
Action is coded in Go (Golang), which means you need to pay special attention to
regular expressions (Regex). Special characters need to be escaped with double
backslashes. This is because the backslash in Go strings is an escape character
and therefore must be escaped itself to appear as a literal in the regex.

### Last Modified (PRs and Issues) <a name="last-modified" />

This condition evaluates the modification date of the PR or Issue. 

If you're looking to evaluate on the creation date of the issue or PR, 
check on <a href="#age" ></a>

This condition is best used when with a <a href="#schedule">schedule trigger</a>.

Examples:

```yaml
last-modified:
  at-most: 1d
```
Will label PRs or issues that were last modified at most one day ago

```yaml
last-modified:
  at-least: 1d
```

Will label PRs or issues that were last modified at least one day ago

The syntax for values is based on a number, followed by a suffix:

* s: seconds
* m: minutes
* h: hours
* d: days
* w: weeks
* y: years

For example, `2d` means 2 days, `4w` means 4 weeks, and so on.

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
`go.sum` and such. Use `exclude-files`, which supports both an explicit
file or a Regex expression:

```yaml
- label: "L"
    size:
        exclude-files: ["yarn.lock", "\\/root\\/.+\\/test.md"]
        above: 100
``` 

This condition will apply the `L` label if the diff is above 100 lines,
but NOT taking into account changes in `yarn.lock`, or any `test.md`
file that is in a subdirectory of `root`.

**NOTICE** the double backslash (`\\`) in the example above. This GitHub
Action is coded in Go (Golang), which means you need to pay special attention to
regular expressions (Regex). Special characters need to be escaped with double
backslashes. This is because the backslash in Go strings is an escape character
and therefore must be escaped itself to appear as a literal in the regex.

**NOTICE** the old format for specifying size properties (`size-above`
and `size-below`) has been deprecated. The action will continue
supporting old configs for now, but users are encouraged to migrate to
the new configuration schema.

### Title <a name="title" />

This condition is satisfied when the title matches on the given regex.

```yaml
title: "^WIP:.*"
```

### Type <a name="type" />

By setting the type attribute in your label configuration, you can specify whether a rule applies exclusively to Pull
Requests (PRs) or Issues. This allows for more precise label management based on the type of GitHub record. The
type condition accepts one of two values:

- `pull_request`
- `issue`

This functionality increases the adaptability of this GitHub Action, allowing users to create more tailored labeling
strategies that differentiate between PRs and Issues or apply universally to both.

#### Pull-Request Only:

```yaml
- label: "needs review"
  type: "pull_request"
  name: ".*bug.*"
```
This rule applies the label "needs review" to Pull Requests with "bug" in the title.

#### Issue Only:

```yaml
- label: "needs triage"
  type: "issue"
  name: ".*bug.*"
```

This rule applies the label "needs triage" to Issues with "bug" in the title.
