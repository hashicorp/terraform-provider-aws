<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Dependency Updates

Generally, dependency updates are handled by maintainers.

## Changelog Entries for Dependency Updates

**Inlcude a changelog entry for dependency updates that address:**

1. Security vulnerabilities
2. Significant changes (e.g., AWS SDK authentication changes)

### Go Updates

Go is transparent about disclosing security fixes and significant changes in updates. To find specifics, see the version milestone. For example, the [Go 1.25.7 milestone](https://github.com/golang/go/issues?q=milestone%3AGo1.25.7) lists four security updates. NOTE: Security updates don't always include the "Security" label. Highlights and links to milestones can also be found in the [Release History](https://go.dev/doc/devel/release).

### AWS SDK for Go V2 and Other Dependency Updates

Other teams aren't as transparent as Go about security fixes. But, a quick way to find disclosed fixes is reviewing the Release Notes and commits in dependabot PRs. For the AWS SDK, you can also review the [CHANGELOG](https://raw.githubusercontent.com/aws/aws-sdk-go-v2/refs/heads/main/CHANGELOG.md).

### Changelog Format

``````
```release-note:note
provider: Updated Go version to v1.25.7 (addresses GO-2026-4337, Unexpected session resumption)
```
``````

## Go Version Update

The Terraform AWS provider is written in Go and is compiled into an executable binary that communicates with [Terraform Core over a local RPC interface](https://developer.hashicorp.com/terraform/plugin).

A new version of Go is [released every 6 months](https://go.dev/wiki/Go-Release-Cycle#overview). [Minor releases](https://go.dev/wiki/MinorReleases), fixing serious problems and security issues, are done [regularly](https://go.dev/doc/devel/release) for the current and previous versions.

<!-- markdownlint-disable MD046 -->

!!! note

    Go versions differ from [semver](https://semver.org/spec/v2.0.0.html). Major versions of Go differ in the second component (upgrading from 1.24.11 to 1.25.5 is a major version upgrade), while minor versions of Go differ just in the third component (upgrading from 1.24.10 to 1.24.11 is a minor version upgrade).

<!-- markdownlint-enable MD046 -->

### When To Upgrade Major Version

The Terraform AWS provider aims to switch to the newest Go version after 1 or 2 minor releases of that version, unless there is an urgent reason to upgrade sooner.

* Upgrading too soon risks tooling (linters etc.) being incompatible with the new Go version
* Upgrading too late risks tooling (linters etc.) being incompatible with the old Go version

### When To Upgrade Minor Version

The Terraform AWS provider should switch to the latest minor version for the next scheduled provider release. If the minor release addresses a critical security issue then a patch release of the provider can be considered.

### To Upgrade The Go Version

#### To Upgrade Major Version

* Review [release notes](https://go.dev/doc/devel/release) for breaking changes and security fixes
* Edit `.go-version`
* Edit each `go.mod`, e.g. `find . -name 'go.mod' -exec sed -i '' 's/go 1.24.10/go 1.25.5/' {} +`
* Run smoke tests, e.g. `make sane`
* If a new version of [`modernize`](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize) has been released supporting the new Go version, update the `make modern-check` and `make modern-fix` [makefile targets](https://hashicorp.github.io/terraform-provider-aws/makefile-cheat-sheet/#cheat-sheet) and fix any new issues
* Create a PR with the changes. See [#45653](https://github.com/hashicorp/terraform-provider-aws/pull/45653) for a recent example
* Note any material changes, such as Go dropping support for very old OS versions, in a CHANGELOG entry

Support for new language and standard library features should be done in separate PRs.

#### To Upgrade Minor Version

* Review [release notes](https://go.dev/doc/devel/release) for breaking changes and security fixes
* Edit `.go-version`
* Edit each `go.mod`, e.g. `find . -name 'go.mod' -print | xargs ruby -p -i -e 'gsub(/go 1.24.10/, "go 1.24.11")'`
* Run smoke tests, e.g. `make sane`
* Create a PR with the changes. See [#45379](https://github.com/hashicorp/terraform-provider-aws/pull/45379) for a recent example

## AWS Go SDK Updates

Almost exclusively, `github.com/aws/aws-sdk-go-v2` updates are additive in nature. It is generally safe to only scan through them before approving and merging. If you have any concerns about any of the service client updates such as suspicious code removals in the update, or deprecations introduced, run the acceptance testing for potentially affected resources before merging.

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
