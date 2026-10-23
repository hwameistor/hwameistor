# Release preparation

Run **Prepare release** with a target branch, explicit version (including prereleases), starting tag and optional ending ref. Install `RELEASE_ASSISTANT_TOKEN` (App token or fine-grained PAT with Contents and Pull requests write permission) and `COPILOT_GITHUB_TOKEN` (Copilot Requests permission), or configure a compatible model endpoint and `RELEASE_ASSISTANT_API_KEY`.

Policy, template and model configuration are read from the latest main branch on every run. **Re-run all jobs** regenerates the same open PR and replaces generated files, including manual edits. Its source range is frozen before inference. If the target branch advances, start a new run with its new full SHA as `to_ref`. Closed/merged PRs and existing tags cannot be reused.

The PR contains the cumulative Changelog, version log, Chart version/appVersion and provenance. Dependency/tool image versions are managed separately. Review included changes, exclusions and uncertain statements, then mark the draft ready. Require **Validate release PR / validate** and an up-to-date branch before merging; install this workflow on each release branch as well.

Merging the Chart version update triggers the existing publication pipeline. It reads the reviewed manifest and version log from the exact merged commit. Missing or changed notes stop publication. New manual version bumps must also provide matching reviewed release metadata; the pipeline no longer silently publishes the Chart description as release notes.

Chart-releaser still packages the Chart and updates the index. The publication companion creates the Release with its prerelease/latest flags already set and uploads the Chart. Retries verify the tag, notes and existing archive content, retaining previously published assets. Binary and image jobs use the same source SHA and explicit version. For Operator manual dispatch, select the same workflow branch as the branch input.

This integration does not change the Chart version and does not publish a product release when merged. Configure and test the credentials and generation first. See https://github.com/hwameistor/release-assistant for the specification, tests and CLI usage.
