---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_integration"
description: |-
  Manages an AWS Continuum integration.
---

# Resource: aws_continuum_integration

Manages an AWS Continuum integration. An integration connects Continuum to a code source: an Amazon S3 bucket, a GitLab project or group, or a GitHub account or organization.

~> **NOTE:** The S3 integration role must use the path `/continuum/integrations/`, trust the `continuum.amazonaws.com` service principal with an `sts:ExternalId` condition, and allow `s3:ListBucket` on the bucket. IAM changes are eventually consistent, so the provider retries creation while Continuum cannot yet assume or use a new role.

~> **NOTE:** Continuum does not return external IDs, access tokens or GitHub authorization values. They are kept from configuration and are not populated by import. The first apply after import records them in state without replacing the integration.

## Example Usage

### Amazon S3

```terraform
resource "aws_continuum_integration" "example" {
  name = "example"

  provider_configuration {
    s3 {
      bucket      = aws_s3_bucket.example.id
      role_arn    = aws_iam_role.continuum_integration.arn
      external_id = "example-external-id"
    }
  }
}
```

### GitHub

```terraform
resource "aws_continuum_integration" "example" {
  name = "example-github"

  provider_configuration {
    github {
      organization_name = "example-org"

      authorization {
        authorization_code = var.github_authorization_code
        state              = var.github_authorization_state
      }
    }
  }
}
```

### GitLab

```terraform
resource "aws_continuum_integration" "example" {
  name = "example-gitlab"

  provider_configuration {
    gitlab {
      access_token = var.gitlab_access_token
      instance_url = "https://gitlab.example.com"
      target_path  = "example-group/example-project"
      target_type  = "PROJECT"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) Name of the integration. Must be between 1 and 255 characters.
* `provider_configuration` - (Required) Code source configuration. See [`provider_configuration`](#provider_configuration-block) below.

The following arguments are optional:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `provider_configuration` Block

Exactly one of the following blocks is required. Changing which block is set forces a new resource.

* `github` - (Optional) GitHub configuration. See [`github`](#github-block) below.
* `gitlab` - (Optional) GitLab configuration. See [`gitlab`](#gitlab-block) below.
* `s3` - (Optional) Amazon S3 configuration. See [`s3`](#s3-block) below.

### `github` Block

* `authorization` - (Required) OAuth authorization used to create the integration. See [`authorization`](#authorization-block) below.
* `organization_name` - (Optional, Forces new resource) GitHub organization to install into. Omit for a personal account.

### `authorization` Block

* `authorization_code` - (Required, Sensitive, Forces new resource) Single-use OAuth authorization code returned by GitHub.
* `redirect_url` - (Optional, Forces new resource) Redirect URL used to obtain the authorization code. Must start with `https://`. Defaults to the GitHub App's default redirect URL.
* `state` - (Required, Forces new resource) State returned by GitHub in the OAuth callback. Must be the state returned by `StartIntegrationAuthorization`.

### `gitlab` Block

* `access_token` - (Required, Sensitive, Forces new resource) GitLab personal, group or project access token.
* `instance_url` - (Required, Forces new resource) HTTPS base URL of the GitLab instance.
* `target_path` - (Required, Forces new resource) Path of the GitLab group or project.
* `target_type` - (Required, Forces new resource) Type of the target. Valid values are `GROUP` and `PROJECT`.

### `s3` Block

* `bucket` - (Required, Forces new resource) Name of the bucket that holds the code.
* `external_id` - (Required, Sensitive) External ID that Continuum presents when assuming the role.
* `role_arn` - (Required) ARN of the IAM role that Continuum assumes to read the bucket.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the integration.
* `created_at` - Time the integration was created.
* `id` - ID of the integration.
* `integration_type` - Type of the integration: `GITHUB`, `GITLAB` or `S3`.
* `updated_at` - Time the integration was last updated.

### `gitlab` Block

In addition to the arguments above, the `gitlab` block exports the following attributes:

* `expires_at` - Time the GitLab access token expires. Not set when `token_rotation` is `AUTOMATIC`.
* `token_rotation` - How the access token is rotated. `AUTOMATIC` when the token has the self-rotate scope, in which case Continuum manages the token; otherwise `NONE`.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_continuum_integration.example
  identity = {
    id = "01a17e22728b7d25b3a2294e0ca5c35e"
  }
}

resource "aws_continuum_integration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) ID of the integration.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_continuum_integration` using the ID. For example:

```terraform
import {
  to = aws_continuum_integration.example
  id = "01a17e22728b7d25b3a2294e0ca5c35e"
}
```

Using `terraform import`, import `aws_continuum_integration` using the ID. For example:

```console
% terraform import aws_continuum_integration.example 01a17e22728b7d25b3a2294e0ca5c35e
```
