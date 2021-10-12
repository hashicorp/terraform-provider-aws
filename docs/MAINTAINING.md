# Maintaining the Terraform AWS Provider

<!-- TOC depthFrom:2 -->

- [Pull Requests](#pull-requests)
    - [Pull Request Review Process](#pull-request-review-process)
        - [Dependency Updates](#dependency-updates)
            - [Go Default Version Update](#go-default-version-update)
            - [AWS Go SDK Updates](#aws-go-sdk-updates)
            - [golangci-lint Updates](#golangci-lint-updates)
            - [Terraform Plugin SDK Updates](#terraform-plugin-sdk-updates)
            - [tfproviderdocs Updates](#tfproviderdocs-updates)
            - [tfproviderlint Updates](#tfproviderlint-updates)
            - [yaml.v2 Updates](#yaml-v2-updates)
    - [Pull Request Merge Process](#pull-request-merge-process)
- [Breaking Changes](#breaking-changes)
- [Environment Variable Dictionary](#environment-variable-dictionary)
- [Release Process](#release-process)

<!-- /TOC -->

## Community Maintainers

Members of the community who participate in any aspects of maintaining the provider must adhere to the HashiCorp [Community Guidelines](https://www.hashicorp.com/community-guidelines).

## Triage

Incoming issues are classified using labels. These are assigned either by automation, or manually during the triage process. We follow a two-label system where we classify by type and by the area of the provider they affect. A full listing of the labels and how they are used can be found in the [Label Dictionary](#label-dictionary).

## Pull Requests

### Pull Request Review Process

Throughout the review process our first priority is to interact with contributors with kindness, empathy and in accordance with the [Guidelines](https://www.hashicorp.com/community-guidelines) and [Principles](https://www.hashicorp.com/our-principles/) of Hashicorp.

Our contributors are often working within the provider as a hobby, or not in their main line of work so we need to give adequate time for response. By default this is a week, but it is worth considering taking on the work to complete the PR ourselves if the administrative effort of waiting for a response is greater than just resolving the issues ourselves (Don't wait the week, or add a context shift for yourself and the contributor to fix a typo). As long as we use their commits, contributions will be recorded by Github and as always ensure to thank the contributor for their work. Roadmap items are another area where we would consider taking on the work ourselves more quickly in order to meet the commitments made to our users.

Notes for each type of pull request are (or will be) available in subsections below.

- If you plan to be responsible for the pull request through the merge/closure process, assign it to yourself
- Add `bug`, `enhancement`, `new-data-source`, `new-resource`, or `technical-debt` labels to match expectations from change
- Perform a quick scan of open issues and ensure they are referenced in the pull request description (e.g. `Closes #1234`, `Relates #5678`). Edit the description yourself and mention this to the author:

```markdown
This pull request appears to be related to/solve #1234, so I have edited the pull request description to denote the issue reference.
```

- Review the contents of the pull request and ensure the change follows the relevant section of the [Contributing Guide](https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/contribution-checklists.md)
- If the change is not acceptable, leave a long form comment about the reasoning and close the pull request
- If the change is acceptable with modifications, leave a pull request review marked using the `Request Changes` option (for maintainer pull requests with minor modification requests, giving feedback with the `Approve` option is recommended so they do not need to wait for another round of review)
- If the author is unresponsive for changes (by default we give two weeks), determine importance and level of effort to finish the pull request yourself including their commits or close the pull request
- Run relevant acceptance testing ([locally](https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/running-and-writing-acceptance-tests.md) or in TeamCity) against AWS Commercial and AWS GovCloud (US) to ensure no new failures are being introduced
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

##### Go Default Version Update

This project typically upgrades its Go version for development and testing shortly after release to get the latest and greatest Go functionality. Before beginning the update process, ensure that you review the new version release notes to look for any areas of possible friction when updating.

Create an issue to cover the update noting down any areas of particular interest or friction.

Ensure that the following steps are tracked within the issue and completed within the resulting pull request.

- Update go version in `go.mod`
- Verify `make test lint` works as expected
- Verify `goreleaser build --snapshot` succeeds for all currently supported architectures
- Verify `goenv` support for the new version
- Update `docs/DEVELOPMENT.md`
- Update `.go-version`
- Update `CHANGELOG.md` detailing the update and mention any notes practitioners need to be aware of.

See [#9992](https://github.com/hashicorp/terraform-provider-aws/issues/9992) / [#10206](https://github.com/hashicorp/terraform-provider-aws/pull/10206)  for a recent example.

##### AWS Go SDK Updates

Almost exclusively, `github.com/aws/aws-sdk-go` updates are additive in nature. It is generally safe to only scan through them before approving and merging. If you have any concerns about any of the service client updates such as suspicious code removals in the update, or deprecations introduced, run the acceptance testing for potentially affected resources before merging.

Authentication changes:

Occasionally, there will be changes listed in the authentication pieces of the AWS Go SDK codebase, e.g. changes to `aws/session`. The AWS Go SDK `CHANGELOG` should include a relevant description of these changes under a heading such as `SDK Enhancements` or `SDK Bug Fixes`. If they seem worthy of a callout in the Terraform AWS Provider `CHANGELOG`, then upon merging we should include a similar message prefixed with the `provider` subsystem, e.g. `* provider: ...`.

Additionally, if a `CHANGELOG` addition seemed appropriate, this dependency and version should also be updated in the Terraform S3 Backend, which currently lives in Terraform Core. An example of this can be found with https://github.com/hashicorp/terraform-provider-aws/pull/9305 and https://github.com/hashicorp/terraform/pull/22055.

CloudFront changes:

CloudFront service client updates have previously caused an issue when a new field introduced in the SDK was not included with Terraform and caused all requests to error (https://github.com/hashicorp/terraform-provider-aws/issues/4091). As a precaution, if you see CloudFront updates, run all the CloudFront resource acceptance testing before merging (`TestAccAWSCloudFront`).

New Regions:

These are added to the AWS Go SDK `aws/endpoints/defaults.go` file and generally noted in the AWS Go SDK `CHANGELOG` as `aws/endpoints: Updated Regions`. Since April 2019, new regions added to AWS now require being explicitly enabled before they can be used. Examples of this can be found when `me-south-1` was announced:

- [Terraform AWS Provider issue](https://github.com/hashicorp/terraform-provider-aws/issues/9545)
- [Terraform AWS Provider AWS Go SDK update pull request](https://github.com/hashicorp/terraform-provider-aws/pull/9538)
- [Terraform AWS Provider data source update pull request](https://github.com/hashicorp/terraform-provider-aws/pull/9547)
- [Terraform S3 Backend issue](https://github.com/hashicorp/terraform/issues/22254)
- [Terraform S3 Backend pull request](https://github.com/hashicorp/terraform/pull/22253)

Typically our process for new regions is as follows:

- Create new (if not existing) Terraform AWS Provider issue: Support Automatic Region Validation for `XX-XXXXX-#` (Location)
- Create new (if not existing) Terraform S3 Backend issue: backend/s3: Support Automatic Region Validation for `XX-XXXXX-#` (Location)
- [Enable the new region in an AWS testing account](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable) and verify AWS Go SDK update works with the new region with `export AWS_DEFAULT_REGION=XX-XXXXX-#` with the new region and run the `TestAccDataSourceAwsRegion_` acceptance testing or by building the provider and testing a configuration like the following:

```hcl
provider "aws" {
  region = "me-south-1"
}

data "aws_region" "current" {}

output "region" {
  value = data.aws_region.current.name
}
```

- Merge AWS Go SDK update in Terraform AWS Provider and close issue with the following information:

``````markdown
Support for automatic validation of this new region has been merged and will release with version <x.y.z> of the Terraform AWS Provider, later this week.

---

Please note that this new region requires [a manual process to enable](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). Once enabled in the console, it takes a few minutes for everything to work properly.

If the region is not enabled properly, or the enablement process is still in progress, you can receive errors like these:

```console
$ terraform apply

Error: error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid.
    status code: 403, request id: 142f947b-b2c3-11e9-9959-c11ab17bcc63

  on main.tf line 1, in provider "aws":
   1: provider "aws" {
```

---

To use this new region before support has been added to Terraform AWS Provider version in use, you must disable the provider's automatic region validation via:

```hcl
provider "aws" {
  # ... potentially other configuration ...

  region                 = "me-south-1"
  skip_region_validation = true
}
```
``````

- Update the Terraform AWS Provider `CHANGELOG` with the following:

```markdown
NOTES:

* provider: Region validation now automatically supports the new `XX-XXXXX-#` (Location) region. For AWS operations to work in the new region, the region must be explicitly enabled as outlined in the [AWS Documentation](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). When the region is not enabled, the Terraform AWS Provider will return errors during credential validation (e.g. `error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid`) or AWS operations will throw their own errors (e.g. `data.aws_availability_zones.available: Error fetching Availability Zones: AuthFailure: AWS was not able to validate the provided access credentials`). [GH-####]

ENHANCEMENTS:

* provider: Support automatic region validation for `XX-XXXXX-#` [GH-####]
```

- Follow the [Contributing Guide](contributing/contribution-checklists.md#new-region) to submit updates for various data sources to support the new region
- Submit the dependency update to the Terraform S3 Backend by running the following:

```shell
go get github.com/aws/aws-sdk-go@v#.#.#
go mod tidy
```

- Create a S3 Bucket in the new region and verify AWS Go SDK update works with new region by building the Terraform S3 Backend and testing a configuration like the following:

```hcl
terraform {
  backend "s3" {
    bucket = "XXX"
    key    = "test"
    region = "me-south-1"
  }
}

output "test" {
  value = timestamp()
}
```

- After approval, merge AWS Go SDK update in Terraform S3 Backend and close issue with the following information:

``````markdown
Support for automatic validation of this new region has been merged and will release with the next version of the Terraform.

This was verified on a build of Terraform with the update:

```hcl
terraform {
  backend "s3" {
    bucket = "XXX"
    key    = "test"
    region = "me-south-1"
  }
}

output "test" {
  value = timestamp()
}
```

Outputs:

```console
$ terraform init
...
Terraform has been successfully initialized!
```

---

Please note that this new region requires [a manual process to enable](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). Once enabled in the console, it takes a few minutes for everything to work properly.

If the region is not enabled properly, or the enablement process is still in progress, you can receive errors like these:

```console
$ terraform init

Initializing the backend...

Error: error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid.
```

---

To use this new region before this update is released, you must disable the Terraform S3 Backend's automatic region validation via:

```hcl
terraform {
  # ... potentially other configuration ...

  backend "s3" {
    # ... other configuration ...

    region                 = "me-south-1"
    skip_region_validation = true
  }
}
```
``````

- Update the Terraform S3 Backend `CHANGELOG` with the following:

```markdown
NOTES:

* backend/s3: Region validation now automatically supports the new `XX-XXXXX-#` (Location) region. For AWS operations to work in the new region, the region must be explicitly enabled as outlined in the [AWS Documentation](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). When the region is not enabled, the Terraform S3 Backend will return errors during credential validation (e.g. `error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid`). [GH-####]

ENHANCEMENTS:

* backend/s3: Support automatic region validation for `XX-XXXXX-#` [GH-####]
```

##### golangci-lint Updates

Merge if CI passes.

##### Terraform Plugin SDK Updates

Except for trivial changes, run the full acceptance testing suite against the pull request and verify there are no new or unexpected failures.

##### tfproviderdocs Updates

Merge if CI passes.

##### tfproviderlint Updates

Merge if CI passes.

##### yaml.v2 Updates

Run the acceptance testing pattern, `TestAccAWSCloudFormationStack(_dataSource)?_yaml`, and merge if passing.

### Pull Request Merge Process

- Add this pull request to the upcoming release milestone
- Add any linked issues that will be closed by the pull request to the same upcoming release milestone
- Merge the pull request
- Delete the branch (if the branch is on this repository)
- Determine if the pull request should have a CHANGELOG entry by reviewing the [Pull Request Types to CHANGELOG section](#pull-request-types-to-changelog), and follow the CHANGELOG specification [here](https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/pullrequest-submission-and-lifecycle.md#changelog-process)
- Leave a comment on any issues closed by the pull request noting that it has been merged and when to expect the release containing it, e.g.

```markdown
The fix for this has been merged and will release with version X.Y.Z of the Terraform AWS Provider, expected in the XXX timeframe.
```

## Breaking Changes

When breaking changes to the provider are necessary we release them in a major version. If an issue or PR necessitates a breaking change, then the following procedure should be observed:

- Add the `breaking-change` label.
- Add the issue/PR to the next major version milestone.
- Leave a comment why this is a breaking change or otherwise only being considered for a major version update. If possible, detail any changes that might be made for the contributor to accomplish the task without a breaking change.

## Environment Variable Dictionary

Environment variables (beyond standard AWS Go SDK ones) used by acceptance testing. See also the `aws/internal/envvar` package.

| Variable | Description |
|----------|-------------|
| `ACM_CERTIFICATE_ROOT_DOMAIN` | Root domain name to use with ACM Certificate testing. |
| `ACM_CERTIFICATE_MULTIPLE_ISSUED_DOMAIN` | Domain name of ACM Certificate with a multiple issued certificates. **DEPRECATED:** Should be replaced with `aws_acm_certficate` resource usage in tests. |
| `ACM_CERTIFICATE_MULTIPLE_ISSUED_MOST_RECENT_ARN` | Amazon Resource Name of most recent ACM Certificate with a multiple issued certificates. **DEPRECATED:** Should be replaced with `aws_acm_certficate` resource usage in tests. |
| `ACM_CERTIFICATE_SINGLE_ISSUED_DOMAIN` | Domain name of ACM Certificate with a single issued certificate. **DEPRECATED:** Should be replaced with `aws_acm_certficate` resource usage in tests. |
| `ACM_CERTIFICATE_SINGLE_ISSUED_MOST_RECENT_ARN` | Amazon Resource Name of most recent ACM Certificate with a single issued certificate. **DEPRECATED:** Should be replaced with `aws_acm_certficate` resource usage in tests. |
| `ADM_CLIENT_ID` | Identifier for Amazon Device Manager Client in Pinpoint testing. |
| `AMPLIFY_DOMAIN_NAME` | Domain name to use for Amplify domain association testing. |
| `AMPLIFY_GITHUB_ACCESS_TOKEN` | GitHub access token used for AWS Amplify testing. |
| `AMPLIFY_GITHUB_REPOSITORY` | GitHub repository used for AWS Amplify testing. |
| `ADM_CLIENT_SECRET` | Secret for Amazon Device Manager Client in Pinpoint testing. |
| `APNS_BUNDLE_ID` | Identifier for Apple Push Notification Service Bundle in Pinpoint testing. |
| `APNS_CERTIFICATE` | Certificate (PEM format) for Apple Push Notification Service in Pinpoint testing. |
| `APNS_CERTIFICATE_PRIVATE_KEY` | Private key for Apple Push Notification Service in Pinpoint testing. |
| `APNS_SANDBOX_BUNDLE_ID` | Identifier for Sandbox Apple Push Notification Service Bundle in Pinpoint testing. |
| `APNS_SANDBOX_CERTIFICATE` | Certificate (PEM format) for Sandbox Apple Push Notification Service in Pinpoint testing. |
| `APNS_SANDBOX_CERTIFICATE_PRIVATE_KEY` | Private key for Sandbox Apple Push Notification Service in Pinpoint testing. |
| `APNS_SANDBOX_CREDENTIAL` | Credential contents for Sandbox Apple Push Notification Service in SNS Application Platform testing. Conflicts with `APNS_SANDBOX_CREDENTIAL_PATH`. |
| `APNS_SANDBOX_CREDENTIAL_PATH` | Path to credential for Sandbox Apple Push Notification Service in SNS Application Platform testing. Conflicts with `APNS_SANDBOX_CREDENTIAL`. |
| `APNS_SANDBOX_PRINCIPAL` | Principal contents for Sandbox Apple Push Notification Service in SNS Application Platform testing. Conflicts with `APNS_SANDBOX_PRINCIPAL_PATH`. |
| `APNS_SANDBOX_PRINCIPAL_PATH` | Path to principal for Sandbox Apple Push Notification Service in SNS Application Platform testing. Conflicts with `APNS_SANDBOX_PRINCIPAL`. |
| `APNS_SANDBOX_TEAM_ID` | Identifier for Sandbox Apple Push Notification Service Team in Pinpoint testing. |
| `APNS_SANDBOX_TOKEN_KEY` | Token key file content (.p8 format) for Sandbox Apple Push Notification Service in Pinpoint testing. |
| `APNS_SANDBOX_TOKEN_KEY_ID` | Identifier for Sandbox Apple Push Notification Service Token Key in Pinpoint testing. |
| `APNS_TEAM_ID` | Identifier for Apple Push Notification Service Team in Pinpoint testing. |
| `APNS_TOKEN_KEY` | Token key file content (.p8 format) for Apple Push Notification Service in Pinpoint testing. |
| `APNS_TOKEN_KEY_ID` | Identifier for Apple Push Notification Service Token Key in Pinpoint testing. |
| `APNS_VOIP_BUNDLE_ID` | Identifier for VOIP Apple Push Notification Service Bundle in Pinpoint testing. |
| `APNS_VOIP_CERTIFICATE` | Certificate (PEM format) for VOIP Apple Push Notification Service in Pinpoint testing. |
| `APNS_VOIP_CERTIFICATE_PRIVATE_KEY` | Private key for VOIP Apple Push Notification Service in Pinpoint testing. |
| `APNS_VOIP_TEAM_ID` | Identifier for VOIP Apple Push Notification Service Team in Pinpoint testing. |
| `APNS_VOIP_TOKEN_KEY` | Token key file content (.p8 format) for VOIP Apple Push Notification Service in Pinpoint testing. |
| `APNS_VOIP_TOKEN_KEY_ID` | Identifier for VOIP Apple Push Notification Service Token Key in Pinpoint testing. |
| `APPRUNNER_CUSTOM_DOMAIN` | A custom domain endpoint (root domain, subdomain, or wildcard) for AppRunner Custom Domain Association testing. |
| `AWS_ALTERNATE_ACCESS_KEY_ID` | AWS access key ID with access to a secondary AWS account for tests requiring multiple accounts. Requires `AWS_ALTERNATE_SECRET_ACCESS_KEY`. Conflicts with `AWS_ALTERNATE_PROFILE`. |
| `AWS_ALTERNATE_SECRET_ACCESS_KEY` | AWS secret access key with access to a secondary AWS account for tests requiring multiple accounts. Requires `AWS_ALTERNATE_ACCESS_KEY_ID`. Conflicts with `AWS_ALTERNATE_PROFILE`. |
| `AWS_ALTERNATE_PROFILE` | AWS profile with access to a secondary AWS account for tests requiring multiple accounts. Conflicts with `AWS_ALTERNATE_ACCESS_KEY_ID` and `AWS_ALTERNATE_SECRET_ACCESS_KEY`. |
| `AWS_ALTERNATE_REGION` | Secondary AWS region for tests requiring multiple regions. Defaults to `us-east-1`. |
| `AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_BODY` | Certificate body of publicly trusted certificate for API Gateway Domain Name testing. |
| `AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_CHAIN` | Certificate chain of publicly trusted certificate for API Gateway Domain Name testing. |
| `AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_PRIVATE_KEY` | Private key of publicly trusted certificate for API Gateway Domain Name testing. |
| `AWS_API_GATEWAY_DOMAIN_NAME_REGIONAL_CERTIFICATE_NAME_ENABLED` | Flag to enable API Gateway Domain Name regional certificate upload testing. |
| `AWS_CODEBUILD_BITBUCKET_SOURCE_LOCATION` | BitBucket source URL for CodeBuild testing. CodeBuild must have access to this repository via OAuth or Source Credentials. Defaults to `https://terraform@bitbucket.org/terraform/aws-test.git`. |
| `AWS_CODEBUILD_GITHUB_SOURCE_LOCATION` | GitHub source URL for CodeBuild testing. CodeBuild must have access to this repository via OAuth or Source Credentials. Defaults to `https://github.com/hashibot-test/aws-test.git`. |
| `AWS_DEFAULT_REGION` | Primary AWS region for tests. Defaults to `us-west-2`. |
| `AWS_EC2_CLASSIC_REGION` | AWS region for EC2-Classic testing. Defaults to `us-east-1` in AWS Commercial and `AWS_DEFAULT_REGION` otherwise. |
| `AWS_EC2_CLIENT_VPN_LIMIT` | Concurrency limit for Client VPN acceptance tests. [Default is 5](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/limits.html) if not specified. |
| `AWS_EC2_EIP_PUBLIC_IPV4_POOL` | Identifier for EC2 Public IPv4 Pool for EC2 EIP testing. |
| `AWS_GUARDDUTY_MEMBER_ACCOUNT_ID` | Identifier of AWS Account for GuardDuty Member testing. **DEPRECATED:** Should be replaced with standard alternate account handling for tests. |
| `AWS_GUARDDUTY_MEMBER_EMAIL` | Email address for GuardDuty Member testing. **DEPRECATED:** It may be possible to use a placeholder email address instead. |
| `AWS_LAMBDA_IMAGE_LATEST_ID` | ECR repository image URI (tagged as `latest`) for Lambda container image acceptance tests.
| `AWS_LAMBDA_IMAGE_V1_ID` | ECR repository image URI (tagged as `v1`) for Lambda container image acceptance tests.
| `AWS_LAMBDA_IMAGE_V2_ID` | ECR repository image URI (tagged as `v2`) for Lambda container image acceptance tests.
| `DX_CONNECTION_ID` | Identifier for Direct Connect Connection testing. |
| `DX_VIRTUAL_INTERFACE_ID` | Identifier for Direct Connect Virtual Interface testing. |
| `EC2_SECURITY_GROUP_RULES_PER_GROUP_LIMIT` | EC2 Quota for Rules per Security Group. Defaults to 50. **DEPRECATED:** Can be augmented or replaced with Service Quotas lookup. |
| `EVENT_BRIDGE_PARTNER_EVENT_BUS_NAME` | Amazon EventBridge partner event bus name. |
| `EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME` | Amazon EventBridge partner event source name. |
| `GCM_API_KEY` | API Key for Google Cloud Messaging in Pinpoint and SNS Platform Application testing. |
| `GITHUB_TOKEN` | GitHub token for CodePipeline testing. |
| `MACIE_MEMBER_ACCOUNT_ID` | Identifier of AWS Account for Macie Member testing. **DEPRECATED:** Should be replaced with standard alternate account handling for tests. |
| `SAGEMAKER_IMAGE_VERSION_BASE_IMAGE` | Sagemaker base image to use for tests. |
| `SERVICEQUOTAS_INCREASE_ON_CREATE_QUOTA_CODE` | Quota Code for Service Quotas testing (submits support case). |
| `SERVICEQUOTAS_INCREASE_ON_CREATE_SERVICE_CODE` | Service Code for Service Quotas testing (submits support case). |
| `SERVICEQUOTAS_INCREASE_ON_CREATE_VALUE` | Value of quota increase for Service Quotas testing (submits support case). |
| `SES_DOMAIN_IDENTITY_ROOT_DOMAIN` | Root domain name of publicly accessible and Route 53 configurable domain for SES Domain Identity testing. |
| `SWF_DOMAIN_TESTING_ENABLED` | Enables SWF Domain testing (API does not support deletions). |
| `TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN` | Email address for Organizations Account testing. |
| `TEST_AWS_SES_VERIFIED_EMAIL_ARN` | Verified SES Email Identity for use in Cognito User Pool testing. |
| `TF_ACC` | Enables Go tests containing `resource.Test()` and `resource.ParallelTest()`. |
| `TF_ACC_ASSUME_ROLE_ARN` | Amazon Resource Name of existing IAM Role to use for limited permissions acceptance testing. |
| `TF_TEST_CLOUDFRONT_RETAIN` | Flag to disable but dangle CloudFront Distributions during testing to reduce feedback time (must be manually destroyed afterwards) |

## Release Process

- Create a milestone for the next release after this release (generally, the next milestone will be a minor version increase unless previously decided for a major or patch version)
- Check the existing release milestone for open items and either work through them or move them to the next milestone
- Run the HashiCorp (non-OSS) TeamCity release job either via:
    - Slack command: `/tcrelease aws #.#.#` (no `v` prefix)
    - Web interface: With the `DEPLOYMENT_TARGET_VERSION` matching the expected release milestone and `DEPLOYMENT_NEXT_VERSION` matching the next release milestone
- Wait for the TeamCity release job to complete either by watching the build logs or Slack notifications
- Close the release milestone
- Create a new GitHub release with the release title exactly matching the tag and milestone (e.g. `v2.22.0`) and copy the entries from the CHANGELOG to the release notes.
