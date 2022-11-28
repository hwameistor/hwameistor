# Release Process

This is the process to follow to make a new release. Similar to [Semantic](https://semver.org/) Version. 
The project refers to the respective components of this triple as < major >.< minor >.< patch >.

The Hwameistor Project's release process is as follows:
- [Minor Release Process](#Minor)
- [Patch Release Process](#Patch)

<span id="Minor"> </span>
## Minor Release Process

A minor release is e.g. v0.1.0.

A minor release requires 3 steps are as follows: 

### Before release

1. Ensure all issues of `priority/critical-urgent` or `priority/important-soon` label have been resolved.
2. Ensure all PRs about feature which should be included in this version have been merged.  
3. Check the latest PeriodCheck based on main branch has passed.
4. SmokeTest (optional)

### Core release

1. An issue is proposing a new release with a changelog since the latest release.
2. Check out master, ensure it's upto date, and ensure you have a clean working directory.
   1. Update `CHANGELOG.md` and create changelog under `changelogs/released/<version>` directory
3. Create a new local release branch from master.
4. Edit file `helm/hwameistor/Chart.yaml`
   1. Modify version field to release version 
   2. Modify appVersion field to release version
   > NOTE: version and appVersion are consistent by default
5. Commit all changes, push the branch, and PR it into master.

### Post release

1. Checkout release packages(includes helm packages, images) are generated correctly.
2. Write release notes to explain the changes of this release corresponding to the changelog.

<span id="Patch"> </span>
## Patch Release Process

A minor release is e.g. v0.1.1.

A patch release requires 3 steps are as follows:

### Before release

1. Ensure all issues of `priority/critical-urgent` or `priority/important-soon` label have been resolved.
2. Check the latest PeriodCheck based on main branch has passed.
3. SmokeTest (optional)

### Core release

1. An issue is proposing a new release with a changelog since the latest release.
2. Check out master, ensure it's upto date, and ensure you have a clean working directory.
   1. Update `CHANGELOG.md` and create changelog under `changelogs/released/<version>` directory
3. Create a new local release branch from master.
4. Edit file `helm/hwameistor/Chart.yaml`
   1. Modify version field to release version
   2. Modify appVersion field to release version
   > NOTE: version and appVersion are consistent by default
5. Commit all changes, push the branch, and PR it into master.

### Post release

1. Checkout release packages(includes helm packages, images) are generated correctly.
2. Write release notes to explain the changes of this release corresponding to the changelog.