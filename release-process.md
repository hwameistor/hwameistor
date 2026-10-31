# Release Process

[Release V2](.github/workflows/release.yaml) publishes Hwameistor when the root
`VERSION` file changes on a supported branch. The version in that file determines
the image tags, Git tag and GitHub Release name. All builds and the final tag use
the exact commit that triggered the workflow, even if the branch advances later.

## Before release

1. Merge the changes intended for this release and resolve release-blocking issues.
2. Select the target branch. The workflow currently supports `main`,
   `release/v1.0.0` and `release/v1.1`. Patch releases use the corresponding
   maintenance branch; its major/minor version must match `VERSION`. Add any new
   release branch to the workflow's branch filter before using it.
3. Ensure required CI and release testing pass for the target branch. Vulnerability
   checks are managed separately: Release V2 does not run CVE scans.
4. Ensure Release V2 is installed on the target branch. When first introducing
   `VERSION`, initialize it to that branch's current version. Initial file creation,
   branch creation and whitespace-only version changes do not publish a release.

## Prepare reviewed release notes

1. Run **Prepare release** for the target branch, desired version and starting tag.
   Review its generated PR, including `CHANGELOG.md`,
   `changelogs/released/<version>/changelog` and the release provenance files.
2. Wait for **Validate release PR / validate** and the required checks to pass,
   then merge the reviewed preparation PR into the target branch.

Prepare release currently updates Chart metadata but does **not** update `VERSION`.
After Release V2 is installed, merging those generated files alone does not publish
a release. Use the separate version update below to trigger publication.

The version log and matching `.release-assistant/manifest.json` must already be
present at the triggering commit. Release V2 checks the manifest's repository,
version and log content hash. Missing or mismatched notes stop the workflow before
building. Publication uses the reviewed log verbatim; it does not generate new notes
with AI. Confirm that the notes still describe the changes on the target branch
before advancing `VERSION`.

## Trigger publication

1. Open a separate PR against the target branch that changes only the root
   `VERSION` file to the prepared version, for example:

   ```text
   v1.1.3
   ```

   Use a `v`-prefixed [Semantic Version](https://semver.org/), such as `v1.1.3` or
   `v1.1.3-alpha1`, greater than the previous value. Build metadata (`+...`) is not
   supported. Do not create the Git tag or GitHub Release in advance.
2. Review and merge the version PR. The resulting push starts **Release V2**.
3. Follow the run in Actions. The workflow executes these stages in order:

   - Validate the version and read the reviewed release notes at the triggering SHA.
   - Build `hwameictl` for Linux, macOS and Windows, each for amd64 and arm64.
   - Build all 11 component images for linux/amd64 and linux/arm64, then push them
     to `ghcr.io/<repository-owner>` and verify their manifests against this build.
   - Recheck the published image digests, create the version tag at the triggering
     SHA, and create a draft GitHub Release.
   - Upload the six CLI binaries, `images.json` and `SHA256SUMS`, then publish the
     Release with the reviewed version log as its description.

The tag is created only after all images have been pushed and verified. Prerelease
versions are marked as prereleases; prereleases and older maintenance releases do
not replace a newer stable release as **Latest**. This workflow does not publish
Helm charts; changing `helm/hwameistor/Chart.yaml` no longer triggers publication.

## Recover a failed run

- For a transient failure before tag creation, use **Re-run failed jobs** on the
  original run. Its source SHA and version remain fixed; changing `VERSION` again
  is unnecessary. A failed push can leave some images in the registry, but the
  workflow does not create a tag or Release before all images pass verification.
- If `publish-release` fails after creating the tag or draft, rerun only that failed
  job. It reuses the run's artifacts and existing tag/draft and retries attachment
  upload. Rebuilding an already-tagged version is blocked. Keep the run artifacts
  available for this retry.
- A rerun does not pick up later repository fixes. If recovery requires code or
  workflow changes, prepare a new release with a higher `VERSION`; do not move an
  existing tag to another commit.

Release V2 currently has no manual dispatch entry. The existing **Release Image**
workflow remains a separate image-only operation, with its existing CVE scanning;
it does not finish Release V2's tag, notes or asset publication.

## Verify the release

Check that the tag resolves to the run's source SHA, all component images have both
supported architectures, and the published Release contains the expected notes and
eight assets. Use `images.json` for the source SHA and image digests, and
`SHA256SUMS` for the CLI binary and image-report checksums. The Release URL appears
in the `publish-release` job summary.
