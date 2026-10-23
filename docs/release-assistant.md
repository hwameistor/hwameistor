# Release preparation

Run **Prepare release** with a target branch, explicit version (including prereleases), starting tag and optional ending ref. Install `RELEASE_ASSISTANT_TOKEN` (App token or fine-grained PAT with Contents and Pull requests write permission) and `COPILOT_GITHUB_TOKEN` (Copilot Requests permission), or configure a compatible model endpoint and `RELEASE_ASSISTANT_API_KEY`.

Policy, template and model configuration are read from the latest main branch on every run. **Re-run all jobs** regenerates the same open PR and replaces generated files, including manual edits. Its source range is frozen before inference. If the target branch advances, start a new run with its new full SHA as `to_ref`. Closed/merged PRs and existing tags cannot be reused.

The PR contains the cumulative Changelog, version log, Chart version/appVersion and provenance. Dependency/tool image versions are managed separately. Review included changes, exclusions and uncertain statements, then mark the draft ready. Require **Validate release PR / validate** and an up-to-date branch before merging; install this workflow on each release branch as well.

After review, merging the Chart version update follows the repository's existing publication workflow and its existing triggers. This integration adds release preparation and PR validation only.

Generated notes are reviewed and committed in the Release PR. Automatically using them as the GitHub Release body, or changing prerelease/latest flags, would require a separately reviewed publication change. Accepting a version such as `v1.1.2-alpha1` in Prepare release does not change how the existing publication workflow handles it. Keep test Release PRs as drafts and do not merge them until an actual release is intended.

This integration does not change the Chart version and does not publish a product release when merged. Configure and test the credentials and generation first. See https://github.com/hwameistor/release-assistant for the specification, tests and CLI usage.
