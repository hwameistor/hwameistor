# Backporting changes to release branches

Merge the original PR into `main`, then use a label to request a backport:

| Label on the original PR | Target branch |
| --- | --- |
| `cherry-pick-v1.0` | `release/v1.0.0` |
| `cherry-pick-v1.1` | `release/v1.1` |
| `cherry-pick-v1` (legacy alias) | `release/v1.0.0` |

Labels may also be added before merging: the workflow waits until the original
PR is merged. Add both version labels to backport to both release branches.
Only these labels are supported; adding another version label does not create
a new target automatically.

The **Cherry pick** workflow opens a separate PR for each target, with the
Kubernetes-style title:

```text
Automated cherry pick of #<original PR number>: <original PR title>
```

The new PR targets the release branch, retains the original PR description,
and receives the `cherry-pick` label. It requires normal CI and review; the
workflow does not enable automatic merging. The action reports its results in
a comment on the original PR.

The workflow backports the **whole PR**, using its actual merged commits even
when the label is added later. To backport only selected commits, create a
manual PR instead. Start its description with a source line such as:

```text
Backport of #1864 to `release/v1.1`.
```

Existing open or merged backport PRs are skipped. The workflow recognizes its
generated marker/branch and manual descriptions starting with `Backport of`
followed by the source PR reference (including `Backport of the fix from #1864`).
Only the first `#number` on that line identifies the source. Keep that source
line or the generated marker when editing the description.

## Conflicts and retries

The workflow stops on cherry-pick conflicts and posts instructions on the
original PR. Resolve conflicts locally and open a manual backport PR using
the source line above. Conflicts are never resolved automatically by favoring
one side.

After fixing a transient failure, open **Actions → Cherry pick → Run workflow**,
select `main`, and enter the original PR number. The workflow reads the PR's
current labels and skips existing open or merged backports. If retrying a
closed, unmerged backport, first delete its generated branch
(`backport-<number>-to-<target branch>`), or reopen the existing PR instead.

## Maintainer setup

The workflow becomes active when its configuration is merged into `main`.
Create the labels listed above and the `cherry-pick` label in the repository.

Configure the repository Actions secret `BACKPORT_TOKEN` with a bot PAT (or
an appropriate GitHub App token). The workflow also accepts the existing
`MY_TOKEN` secret as a fallback. The token needs access to this repository
with **Contents: read/write** and **Pull requests: read/write**, including
permission to label/comment on PRs. For a classic PAT, use the `repo` scope.
Backporting workflow files also requires the corresponding workflow write
permission (`workflow` scope for a classic PAT).

The workflow intentionally does not fall back to `GITHUB_TOKEN` for writes:
PRs created with that token do not trigger normal `pull_request` CI. Token
validity, repository access, and CI triggering must be verified in the first
real backport run. Run helper tests locally with:

```sh
node --test .github/scripts/prepare-backport.test.cjs
```
