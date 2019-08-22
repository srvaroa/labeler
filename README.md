# Rule based Pull Request Labeler

Implements a [GitHub
Action](https://help.github.com/en/categories/automating-your-workflow-with-github-actions)
that labels Pull Requests based on configurable rules.

It is inspired by the example [Pull Request
Labeller](https://github.com/actions/labeler), but intends to provide a
richer set of options.

Configuration can be stored at `./github/labeler.yml` as a plain list of
labels and a set of rules for each.  When *all* rules for a label match,
then the Action will set the given label.  When *any* rule for a label
does not match, then the Action will unset the given label.

    <label>:
      <rule_name>: <rule_parameters>
      <rule_name>: <rule_parameters>

For example, given this `./github/labeler.yml`:

      WIP:
        title: "^WIP:.*"

A Pull Request with title "WIP: this is work in progress" would be
labelled as `WIP`.  If the Pull Request title changes to "This is done",
then the `WIP` label would be removed.

## Rules

Below are the rules currently supported.  The general format is:

### Regex on title

    WIP:
      title: "^WIP:.*"
