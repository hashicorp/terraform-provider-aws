---
subcategory: "CodeBuild"
layout: "aws"
page_title: "AWS: aws_codebuild_webhook"
description: |-
  Provides a CodeBuild Webhook resource.
---

# Resource: aws_codebuild_webhook

Manages a CodeBuild webhook, which is an endpoint accepted by the CodeBuild service to trigger builds from source code repositories. Depending on the source type of the CodeBuild project, the CodeBuild service may also automatically create and delete the actual repository webhook as well.

## Example Usage

### Bitbucket and GitHub

When working with [Bitbucket](https://bitbucket.org) and [GitHub](https://github.com) source CodeBuild webhooks, the CodeBuild service will automatically create (on `aws_codebuild_webhook` resource creation) and delete (on `aws_codebuild_webhook` resource deletion) the Bitbucket/GitHub repository webhook using its granted OAuth permissions. This behavior cannot be controlled by Terraform.

~> **Note:** The AWS account that Terraform uses to create this resource *must* have authorized CodeBuild to access Bitbucket/GitHub's OAuth API in each applicable region. This is a manual step that must be done *before* creating webhooks with this resource. If OAuth is not configured, AWS will return an error similar to `ResourceNotFoundException: Could not find access token for server type github`. More information can be found in the CodeBuild User Guide for [Bitbucket](https://docs.aws.amazon.com/codebuild/latest/userguide/sample-bitbucket-pull-request.html) and [GitHub](https://docs.aws.amazon.com/codebuild/latest/userguide/sample-github-pull-request.html).

~> **Note:** Further managing the automatically created Bitbucket/GitHub webhook with the `bitbucket_hook`/`github_repository_webhook` resource is only possible with importing that resource after creation of the `aws_codebuild_webhook` resource. The CodeBuild API does not ever provide the `secret` attribute for the `aws_codebuild_webhook` resource in this scenario.

```terraform
resource "aws_codebuild_webhook" "example" {
  project_name = aws_codebuild_project.example.name
  build_type   = "BUILD"
  filter_group {
    filter {
      type    = "EVENT"
      pattern = "PUSH"
    }

    filter {
      type    = "BASE_REF"
      pattern = "master"
    }
  }
}
```

### GitHub Enterprise

When working with [GitHub Enterprise](https://enterprise.github.com/) source CodeBuild webhooks, the GHE repository webhook must be separately managed (e.g., manually or with the `github_repository_webhook` resource).

More information creating webhooks with GitHub Enterprise can be found in the [CodeBuild User Guide](https://docs.aws.amazon.com/codebuild/latest/userguide/sample-github-enterprise.html).

```terraform
resource "aws_codebuild_webhook" "example" {
  project_name = aws_codebuild_project.example.name
}

resource "github_repository_webhook" "example" {
  active     = true
  events     = ["push"]
  name       = "example"
  repository = github_repository.example.name

  configuration {
    url          = aws_codebuild_webhook.example.payload_url
    secret       = aws_codebuild_webhook.example.secret
    content_type = "json"
    insecure_ssl = false
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `project_name` - (Required) The name of the build project.
* `build_type` - (Optional) The type of build this webhook will trigger. Valid values for this parameter are: `BUILD`, `BUILD_BATCH`.
* `branch_filter` - (Optional) A regular expression used to determine which branches get built. Default is all branches are built. We recommend using `filter_group` over `branch_filter`.
* `filter_group` - (Optional) Information about the webhook's trigger. Filter group blocks are documented below.
* `scope_configuration` - (Optional) Scope configuration for global or organization webhooks. Scope configuration blocks are documented below.

`filter_group` supports the following:

* `filter` - (Required) A webhook filter for the group. Filter blocks are documented below.

`filter` supports the following:

* `type` - (Required) The webhook filter group's type. Valid values for this parameter are: `EVENT`, `BASE_REF`, `HEAD_REF`, `ACTOR_ACCOUNT_ID`, `FILE_PATH`, `COMMIT_MESSAGE`, `WORKFLOW_NAME`, `TAG_NAME`, `RELEASE_NAME`. At least one filter group must specify `EVENT` as its type.
* `pattern` - (Required) For a filter that uses `EVENT` type, a comma-separated string that specifies one event: `PUSH`, `PULL_REQUEST_CREATED`, `PULL_REQUEST_UPDATED`, `PULL_REQUEST_REOPENED`. `PULL_REQUEST_MERGED`, `WORKFLOW_JOB_QUEUED` works with GitHub & GitHub Enterprise only. For a filter that uses any of the other filter types, a regular expression.
* `exclude_matched_pattern` - (Optional) If set to `true`, the specified filter does *not* trigger a build. Defaults to `false`.

`scope_configuration` supports the following:

* `name` - (Required) The name of either the enterprise or organization.
* `scope` - (Required) The type of scope for a GitHub webhook. Valid values for this parameter are: `GITHUB_ORGANIZATION`, `GITHUB_GLOBAL`.
* `domain` - (Optional) The domain of the GitHub Enterprise organization. Required if your project's source type is GITHUB_ENTERPRISE.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the build project.
* `payload_url` - The CodeBuild endpoint where webhook events are sent.
* `secret` - The secret token of the associated repository. Not returned by the CodeBuild API for all source types.
* `url` - The URL to the webhook.

~> **Note:** The `secret` attribute is only set on resource creation, so if the secret is manually rotated, terraform will not pick up the change on subsequent runs.  In that case, the webhook resource should be tainted and re-created to get the secret back in sync.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeBuild Webhooks using the CodeBuild Project name. For example:

```terraform
import {
  to = aws_codebuild_webhook.example
  id = "MyProjectName"
}
```

Using `terraform import`, import CodeBuild Webhooks using the CodeBuild Project name. For example:

```console
% terraform import aws_codebuild_webhook.example MyProjectName
```
