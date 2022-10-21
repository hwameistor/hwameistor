# Release Process

The Hwameistor Project's release process is as follows:

## Before release:

1. Ensure all issues of `priority/critical-urgent` or `priority/important-soon` label have been resolved.
2. Check the latest PeriodCheck based on main branch has passed.

## Core release:

1. Check out master, ensure it's up to date, and ensure you have a clean working directory.
2. Create a new local release branch from master.
3. Edit file `helm/hwameistor/Chart.yaml`
   1. Modify version field to release version 
   2. Modify appVersion field to release version
   > NOTE: version and appVersion are consistent by default 
4. Commit all changes, push the branch, and PR it into master.

## Post release:

1. Checkout release packages(includes helm packages, images) are generated correctly.
2. Write release notes to explain the changes of this update.
