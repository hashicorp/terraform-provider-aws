<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Raising a Pull Request

1. [Fork the GitHub repository](https://help.github.com/en/articles/fork-a-repo) allowing you to make the changes in your own copy of the repository.

1. Create a branch using the following naming prefixes:

    - f = feature
    - b = bug fix
    - d = documentation
    - t = tests
    - td = technical debt
    - v = dependencies ("vendoring" previously)

    Some indicative example branch names would be `f-aws_emr_instance_group-refactor` or `td-staticcheck-st1008`

1. Make the changes you would like to include in the provider, add new tests as required, and make sure that all relevant existing tests are passing.

1. [Create a pull request](https://help.github.com/en/articles/creating-a-pull-request-from-a-fork). Please ensure (if possible) that the 'Allow edits from maintainers' checkbox is checked. This will allow the maintainers to make changes and merge the PR without requiring action from the contributor.
   You are welcome to submit your pull request for commentary or review before
   it is fully completed by creating a [draft pull request](https://help.github.com/en/articles/about-pull-requests#draft-pull-requests).
   Please include specific questions or items you'd like feedback on.

1. Pull Request Best Practices
    - **Descriptive Titles:**  When creating a pull request (PR), use a clear and descriptive title that highlights the primary change. If the change pertains to a specific resource or data source, include its name in the title.
    - **Detailed Descriptions:**  Provide a comprehensive description that explains the reasoning behind the change, what was modified, and any expected changes to the user experience (if applicable).
    - **Focused and Manageable Scope:**
        * Keep pull requests small and focused on a single change.
        * For resource or data source additions, each PR should contain only one item and its corresponding tests.
        * Avoid bundling multiple resources or combining service client additions with resource changes in a single PR. Such combinations are significantly harder and more time-consuming for maintainers to review.

1. Create a changelog entry following the process outlined [here](changelog-process.md)

1. Once you believe your pull request is ready to be reviewed, ensure the
   pull request is not a draft pull request by [marking it ready for review](https://help.github.com/en/articles/changing-the-stage-of-a-pull-request)
   or removing `[WIP]` from the pull request title if necessary, and a
   maintainer will review it. Follow [the checklists below](#resource-contribution-guidelines)
   to help ensure that your contribution can be easily reviewed and potentially
   merged.

1. One of the Terraform AWS provider team members will look over your contribution and
   either approve it or provide comments letting you know if there is anything
   left to do. We'll try to give you the opportunity to make the required changes yourself, but in some cases, we may perform the changes ourselves if it makes sense to (minor changes, or for urgent issues).  We do our best to keep up with the volume of PRs waiting for review, but it may take some time depending on the complexity of the work.

1. Once all outstanding comments and checklist items have been addressed, your
   contribution will be merged! Merged PRs will be included in the next
   Terraform AWS provider release.

1. In some cases, we might decide that a PR should be closed without merging.
   We'll make sure to provide clear reasoning when this happens.

### AI Usage

The Terraform AWS Provider follows the AI usage policy documented in the Terraform Core [contributor guide](https://github.com/hashicorp/terraform/blob/main/.github/CONTRIBUTING.md#ai-usage).
See [AI Usage](ai-usage.md) for the full policy.
