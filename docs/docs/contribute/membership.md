# HwameiStor Membership Roles

This document describes the set of roles individuals may have within the HwameiStor community, the requirements of each role, and the privileges that each role grants.

- [Role summary](#role-summary)
- [Collaborator](#collaborator)
- [Member](#member)
- [Reviewer](#reviewer)
- [Maintainer](#maintainer)
- [Administrator](#administrator)

## Role summary

Here is the set of roles we use within the HwameiStor community, the general responsibilities expected
by individuals in each role, the requirements necessary to join or stay in a given role, and the concrete
manifestation of the role in terms of permissions and privileges.

<table>
    <tr>
      <td>Role</td>
      <td>Responsibilities</td>
      <td>Requirements</td>
      <td>Privileges</td>
    </tr>
    <tr>
      <td><a href="#collaborator">Collaborator</a></td>
      <td>Casual contributor to the project</td>
      <td>n/a</td>
      <td>
          <p>Outside collaborator of the GitHub HwameiStor organization</p>
          <p>Can submit PRs and issues</p>
          <p>Read and commenting permission on the HwameiStor Team drive</p>
      </td>
    </tr>
    <tr>
      <td><a href="#member">Member</a></td>
      <td>Regular active contributor in the community</td>
      <td>
          <p>Has pushed at least one PR to an HwameiStor repository</p>
      </td>
      <td>
          <p>Member of the GitHub HwameiStor organization</p>
          <p>Edit permission on the HwameiStor Team drive</p>
          <p>Triage permission on the HwameiStor repos, allowing issues to be manipulated.</p>
      </td>
    </tr>
    <tr>
      <td><a href="#documentation-reviewer">Reviewer</a></td>
      <td>Content expert helping improve code and documentation.</td>
      <td>
        Highly experienced contributor to the HwameiStor documentation.
      </td>
      <td>Like a member, plus:
          <p>Maintainers and Administrator prioritize approval of the content they reviewed.</p>
      </td>
    </tr>
    <tr>
      <td><a href="#maintainer">Maintainer</a></td>
      <td>Approve contributions from other members</td>
      <td>Highly experienced and active reviewer and contributor to an area</td>
      <td>Like a member, plus:
          <p>Able to approve code changes in GitHub</p>
          <p>Voting rights in the context of working group decision-making</p>
          <p>Responsible for making sure that release notes and upgrade notes get added to pull requests with user facing changes</p>
      </td>
    </tr>
    <tr>
      <td><a href="#administrator">Administrator</a></td>
      <td>Manage and control permissions</td>
      <td>Appointed by the HwameiStor organization</td>
      <td>
          <p>Admin privileges on varous HwameiStor-related resources</p>
      </td>
    </tr>
</table>

## Collaborator

Individuals may be added as an outside collaborator (with READ access) to a repo in the HwameiStor GitHub
organization without becoming a member. This allows them to be assigned issues and PRs until they become a member,
but will not allow tests to be run against their PRs automatically nor allow them to interact with the PR bot.

### Requirements

Working on some contribution to the project that would benefit from the ability to have PRs or Issues to be assigned to the contributor.

## Member

Established community members are expected to demonstrate their adherence to the principles in this document,
familiarity with project organization, roles, policies, procedures, conventions, etc., and technical and/or writing ability.

Members are continuously active contributors in the community. They can have issues and PRs assigned to them,
participate in working group meetings, and pre-submit tests are automatically run for their PRs.
Members are expected to remain active contributors to the community.

All members are encouraged to help with the code review burden, although each PR must be reviewed by one or more
official reviewers and maintainers for the area before being accepted into the source base.

### Requirements

- Has pushed **at least one PR** to the HwameiStor repositories within the last 6 months.
- Actively contributing to one or more areas.

Members are expected to be active participants in the project on an on-going basis.
If an individual doesn't contribute to the project for a 180 day period,
that individual may lose membership. On-going contributions include:

- Successfully merging pull requests
- Triaging issues or pull requests
- Commenting on issues or pull requests
- Closing issues or pull requests

### Becoming a member

If you are interested in becoming a member and meet the requirements above, you can join the organization
by adding yourself to the members list under [`members.yaml`](./members.yaml). Once that has been done,
submit a Pull Request with the change and fill out the pull request template with all information requested.

### Responsibilities and privileges

- Responsive to issues and PRs assigned to them
- Active owner of code they have contributed (unless ownership is explicitly transferred)
  - Code is well tested
  - Tests consistently pass
  - Addresses bugs or issues discovered after code is accepted

Members who frequently contribute code are expected to proactively perform code reviews for the area that they are active in.

## Reviewer

A **Reviewer** is trusted to only approve content that meets the acceptance criteria described in the
[contribution guides](./CONTRIBUTING.md).

### Requirements

To become a **Reviewer**, contributors must meet the following **requirements**:

- Be a **Member** of the HwameiStor community.
- Perform 5 substantial contributions to the `HwameiStor.io` repo. Substantial
  contributions include the following examples:
  - New content
  - Content reviews
  - Content improvements
- Demonstrate a solid commitment to documentation quality and use of our style guide.
- Be sponsored by an HwameiStor Maintainer or WG Lead.

### Responsibilities

- Review PRs in `hwameistor/hwameistor`.
- Ensure the relevant technical Working Group is added as a reviewer and ensure
  a maintainer or administrator has approved the PR.

### Privileges

- Content approved by a **Reviewer** gets prioritized by Maintainers or Administrator.
- Reviewers can place a `/lgtm` label to notify Maintainers to expedite publication of the reviewed content.

Reviewers can't merge content into the `hwameistor/hwameistor` main; only Maintainers and Administrator can merge content into main.

## Maintainer

Maintainers review and approve code contributions. While code review is focused on code quality and correctness,
approval is focused on holistic acceptance of a contribution including: backwards / forwards compatibility,
adhering to API and flag conventions, subtle performance and correctness issues, interactions with other parts of the system, etc.
Maintainer status is scoped to a part of the codebase and is reflected in a CODEOWNERS file.

### Requirements

The following apply to the part of the codebase for which one would be a maintainer:

- Member for at least 3 months
- Contributed at least 30 substantial PRs to the codebase
- Must remain an active participant in the community by contributing
  code, performing reviews, triaging issues, etc.
- Knowledgeable about the codebase
- Sponsored by a working group lead with no objections from other leads

If a maintainer becomes inactive in the project for an extended period of time, the individual will transition to being an
emeritus maintainer. Emeritus maintainers lose their ability to approve code contributions, but retain their voting rights
for up to one year. After one year, emeritus maintainers revert back to being normal members with no voting rights.

Maintainers contribute to the parts of the project they are responsible for by:

- Successfully merging pull requests
- Triaging issues or pull requests
- Closing issues or pull requests

### Responsibilities and privileges

The following apply to the part of the codebase for which one would be a maintainer:

- Maintainer status may be a precondition to accepting large code contributions
- Demonstrates sound technical judgement
- Responsible for project quality control via code reviews
  - Focus on code quality and correctness, including testing and factoring
  - Focus on holistic acceptance of contribution such as dependencies with other features, backwards / forwards compatibility, API and flag definitions, etc
- Expected to be responsive to review requests as per community expectations
- May approve code contributions for acceptance
- Maintainers in an area get a vote when a working group needs to make decisions.

## Administrator

Administrators are responsible for the bureaucratic aspects of the project.

### Requirements

Appointed by the HwameiStor organization.

### Responsibilities and privileges

- Manage a variety of infrastructure support for the HwameiStor project
- Although admins may have the authority to override any policy and cut corners, we expect admins to generally abide
  by the overall rules of the project. For example, unless strictly necessary, admins should not approve and/or commit
  PRs they aren't entitled to if they were not admins.
