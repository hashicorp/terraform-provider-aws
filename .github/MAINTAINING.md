# Maintaining the Terraform AWS Provider

<!-- TOC depthFrom:2 -->

- [Pull Requests](#pull-requests)
    - [Pull Request Review Process](#pull-request-review-process)
        - [Dependency Updates](#dependency-updates)
            - [AWS Go SDK Updates](#aws-go-sdk-updates)
            - [Terraform Updates](#terraform-updates)
    - [Pull Request Merge Process](#pull-request-merge-process)
- [Release Process](#release-process)

<!-- /TOC -->

## Pull Requests

### Pull Request Review Process

Notes for each type of pull request are (or will be) available in subsections below.

- If you plan to be responsible for the pull request through the merge/closure process, assign it to yourself
- Add `bug`, `enhancement`, `new-data-source`, `new-resource`, or `technical-debt` labels to match expectations from change
- Review the contents of the pull request and ensure the change follows the relevant section of the [Contributing Guide](https://github.com/terraform-providers/terraform-provider-aws/blob/master/.github/CONTRIBUTING.md#checklists-for-contribution)
- If the change is not acceptable, leave a long form comment about the reasoning and close the pull request
- If the change is acceptable with modifications, leave a pull request review marked using the `Request Changes` option (for maintainer pull requests with minor modification requests, giving feedback with the `Approve` option is recommended so they do not need to wait for another round of review)
- If the author is unresponsive for changes (by default we give two weeks), determine importance and level of effort to finish the pull request yourself including their commits or close the pull request
- Run relevant acceptance testing ([locally](https://github.com/terraform-providers/terraform-provider-aws/blob/master/.github/CONTRIBUTING.md#running-an-acceptance-test) or in TeamCity) against AWS Commercial and AWS GovCloud (US) to ensure no new failures are being introduced
- Approve the pull request with a comment outlining what steps you took that ensure the change is acceptable, e.g. acceptance testing output

``````markdown
Looks good, thanks @username! :rocket:

Output from acceptance testing in AWS Commercial:

```
--- PASS: TestAcc...
--- PASS: TestAcc...
```

Output from acceptance testing in AWS GovCloud (US):

```
--- PASS: TestAcc...
--- PASS: TestAcc...
```
``````

#### Dependency Updates

##### AWS Go SDK Updates

Almost exclusively, `github.com/aws/aws-sdk-go` updates are additive in nature. It is generally safe to only scan through them before approving and merging. If you have any concerns about any of the service client updates such as suspicious code removals in the update, or deprecations introduced, run the acceptance testing for potentially affected resources before merging.

Other important notes:

- CloudFront service client updates have previously caused an issue when a new field introduced in the SDK was not included with Terraform and caused all requests to error (https://github.com/terraform-providers/terraform-provider-aws/issues/4091). As a precaution, if you see CloudFront updates, run all the CloudFront resource acceptance testing before merging (`TestAccAWSCloudFront`).
- Occassionally, there will be changes listed in the authentication pieces of the AWS Go SDK codebase, e.g. changes to `aws/session`. The AWS Go SDK `CHANGELOG` should include a relevant description of these changes under a heading such as `SDK Enhancements` or `SDK Bug Fixes`. If they seem worthy of a callout in the Terraform AWS Provider `CHANGELOG`, then upon merging we should include a similar message prefixed with the `provider` subsystem, e.g. `* provider: ...`. Additionally, if a `CHANGELOG` addition seemed appropriate, this dependency and version should also be updated in the Terraform S3 Backend, which currently lives in Terraform Core. An example of this can be found with https://github.com/terraform-providers/terraform-provider-aws/pull/9305 and https://github.com/hashicorp/terraform/pull/22055.

##### Terraform Updates

Run the full acceptance testing suite against the pull request and verify there are no new or unexpected failures.

### Pull Request Merge Process

- Add this pull request to the upcoming release milestone
- Add any linked issues that will be closed by the pull request to the same upcoming release milestone
- Merge the pull request
- Delete the branch (if the branch is on this repository)
- Update the repository `CHANGELOG.md` by directly committing to the `master` branch. See also the [Extending Terraform documentation](https://www.terraform.io/docs/extend/best-practices/versioning.html) for more information about the expected CHANGELOG format.
- Leave a comment on any issues closed by the pull request noting that it has been merged and when to expect the release containing it, e.g.

```markdown
The fix for this has been merged and will release with version X.Y.Z of the Terraform AWS Provider, expected in the XXX timeframe.
```

## Release Process

- Create a milestone for the next release after this release (generally, the next milestone will be a minor version increase unless previously decided for a major or patch version)
- Check the existing release milestone for open items and either work through them or move them to the next milestone
- Run the HashiCorp (non-OSS) TeamCity release job with the `DEPLOYMENT_TARGET_VERSION` matching the expected release milestone and `DEPLOYMENT_NEXT_VERSION` matching the next release milestone
- Wait for the TeamCity release job and TeamCity website deployment jobs to complete either by watching the build logs or Slack notifications
- Close the release milestone
- For each item noted in the `CHANGELOG.md` for the release just completed (or milestone as a whole), add a comment to the pull request and any linked closed issues noting that it has been released, e.g.

```markdown
This has been released in [version 2.19.0 of the Terraform AWS provider](https://github.com/terraform-providers/terraform-provider-aws/blob/master/CHANGELOG.md#2190-july-11-2019). Please see the [Terraform documentation on provider versioning](https://www.terraform.io/docs/configuration/providers.html#provider-versions) or reach out if you need any assistance upgrading.

For further feature requests or bug reports with this functionality, please create a [new GitHub issue](https://github.com/terraform-providers/terraform-provider-aws/issues/new/choose) following the template for triage. Thanks!
```
