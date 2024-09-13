# Dependency Updates

Generally, dependency updates are handled by maintainers.

## Go Default Version Update

This project typically upgrades its Go version for development and testing shortly after release to get the latest and greatest Go functionality. Before beginning the update process, ensure that you review the new version release notes to look for any areas of possible friction when updating.

Create an issue to cover the update noting down any areas of particular interest or friction.

Ensure that the following steps are tracked within the issue and completed within the resulting pull request.

- Update go version in `go.mod`
- Verify `make test lint` works as expected
- Verify `goreleaser build --snapshot` succeeds for all currently supported architectures
- Verify `goenv` support for the new version
- Update `docs/development-environment.md`
- Update `.go-version`
- Update `CHANGELOG.md` detailing the update and mention any notes practitioners need to be aware of.

See [#9992](https://github.com/hashicorp/terraform-provider-aws/issues/9992) / [#10206](https://github.com/hashicorp/terraform-provider-aws/pull/10206)  for a recent example.

## AWS Go SDK Updates

Almost exclusively, `github.com/aws/aws-sdk-go` and `github.com/aws/aws-sdk-go-v2` updates are additive in nature. It is generally safe to only scan through them before approving and merging. If you have any concerns about any of the service client updates such as suspicious code removals in the update, or deprecations introduced, run the acceptance testing for potentially affected resources before merging.

### Authentication changes

Occasionally, there will be changes listed in the authentication pieces of the AWS Go SDK codebase, e.g., changes to `aws/session`. The AWS Go SDK `CHANGELOG` should include a relevant description of these changes under a heading such as `SDK Enhancements` or `SDK Bug Fixes`. If they seem worthy of a callout in the Terraform AWS Provider `CHANGELOG`, then upon merging we should include a similar message prefixed with the `provider` subsystem, e.g., `* provider: ...`.

Additionally, if a `CHANGELOG` addition seemed appropriate, this dependency and version should also be updated in the Terraform S3 Backend, which currently lives in Terraform Core. An example of this can be found at https://github.com/hashicorp/terraform-provider-aws/pull/9305 and https://github.com/hashicorp/terraform/pull/22055.

### CloudFront changes

CloudFront service client updates have previously caused an issue when a new field introduced in the SDK was not included with Terraform and caused all requests to error (https://github.com/hashicorp/terraform-provider-aws/issues/4091). As a precaution, if you see CloudFront updates, run all the CloudFront resource acceptance testing before merging (`TestAccCloudFront`).

## golangci-lint Updates

Merge if CI passes.

## Terraform Plugin Development Package Updates (SDK V2 or Framework)

Except for trivial changes, run the full acceptance testing suite against the pull request and verify there are no new or unexpected failures.

## tfproviderdocs Updates

Merge if CI passes.

## tfproviderlint Updates

Merge if CI passes.

## yaml.v2 Updates

Run the acceptance testing pattern, `TestAccCloudFormationStack(_dataSource)?_yaml`, and merge if passing.
