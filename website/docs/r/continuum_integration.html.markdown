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

~> **NOTE:** Continuum does not return external IDs, access tokens or GitHub authorization codes. They are kept from configuration and are not populated by import.

## Example Usage

### Amazon S3

```terraform
resource "aws_continuum_integration" "example" {
  name = "example"

  s3 {
    bucket      = aws_s3_bucket.example.id
    role_arn    = aws_iam_role.continuum_integration.arn
    external_id = "example-external-id"
  }
}
```

### GitLab

```terraform
resource "aws_continuum_integration" "example" {
  name = "example-gitlab"

  gitlab {
    access_token = var.gitlab_access_token
    instance_url = "https://gitlab.example.com"
    target_path  = "example-group/example-project"
    target_type  = "PROJECT"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required, Forces new resource) Name of the integration.
* `github` - (Optional, Forces new resource) GitHub provider configuration. See [`github`](#github) below. Exactly one of `github`, `gitlab` or `s3` is required.
* `gitlab` - (Optional, Forces new resource) GitLab provider configuration. See [`gitlab`](#gitlab) below.
* `s3` - (Optional) Amazon S3 provider configuration. See [`s3`](#s3) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `github`

* `authorization` - (Required) Single-use OAuth authorization. See [`authorization`](#authorization) below.
* `organization_name` - (Optional) GitHub organization to install into. Omit for a personal account.

### `authorization`

* `authorization_code` - (Required, Sensitive) Single-use OAuth authorization code returned by GitHub.
* `redirect_url` - (Required) Redirect URL used for the authorization.
* `state` - (Required) State value from `StartIntegrationAuthorization`.

### `gitlab`

* `access_token` - (Required, Sensitive) GitLab access token.
* `instance_url` - (Required) URL of the GitLab instance.
* `target_path` - (Required) Path of the GitLab project or group.
* `target_type` - (Required) Type of the target. Valid values are `PROJECT`.

### `s3`

* `bucket` - (Required, Forces new resource) Name of the bucket that holds the code.
* `role_arn` - (Required) ARN of the IAM role that Continuum assumes to read the bucket. Can be updated in place.
* `external_id` - (Optional, Sensitive) External ID that Continuum presents when assuming the role. Can be updated in place.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the integration.
* `created_at` - Time the integration was created.
* `id` - ID of the integration.
* `integration_type` - Type of the integration: `GITHUB`, `GITLAB` or `S3`.
* `gitlab.0.expires_at` - Time the GitLab access token expires.
* `gitlab.0.token_rotation` - Token rotation mode of the GitLab integration.
* `updated_at` - Time the integration was last updated.

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
  id = "ID_PLACEHOLDER"
}
```

Using `terraform import`, import `aws_continuum_integration` using the ID. For example:

```console
% terraform import aws_continuum_integration.example ID_PLACEHOLDER
```

