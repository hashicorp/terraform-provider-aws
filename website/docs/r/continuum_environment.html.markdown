---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_environment"
description: |-
  Manages an AWS Continuum environment.
---

# Resource: aws_continuum_environment

Manages an AWS Continuum environment. An environment ties one or more code sources to a scan schedule and patch policy.

~> **NOTE:** Creating an environment starts a scan immediately. Scans are billed per issue worked. Set `scan_frequency` to `PT0H` to turn off automatic scans; a scheduled scan of unchanged code adds a duplicate set of issues.

## Example Usage

```terraform
resource "aws_continuum_environment" "example" {
  name           = "example"
  scan_frequency = "PT0H"

  s3_source {
    integration_id = aws_continuum_integration.example.id
    code_prefix    = "code/"
  }
}
```

### GitHub repositories

```terraform
resource "aws_continuum_environment" "example" {
  name = "example"

  repository_source {
    integration_id = aws_continuum_integration.github.id
    repository     = "example-org/example-repo"
    default_branch = "main"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `allowed_domains` - (Optional) Set of domains that code and context in this environment may reach over the network.
* `name` - (Required) Name of the environment.
* `patch_creation_policy` - (Optional) Whether Continuum creates patches. Valid values are `ENABLED` and `DISABLED`. Computed when omitted.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `repository_source` - (Optional) Repository code sources, up to 10. See [`repository_source`](#repository_source-block) below.
* `s3_source` - (Optional) Amazon S3 code sources. See [`s3_source`](#s3_source-block) below.
* `scan_frequency` - (Optional) ISO 8601 duration between automatic scans, such as `P1D` or `P7D`. `PT0H` turns automatic scans off. Computed when omitted.

### `repository_source` Block

* `default_branch` - (Optional) Branch to scan. Computed when omitted.
* `integration_id` - (Required) ID of the GitHub or GitLab integration.
* `repository` - (Required) Repository to scan.

### `s3_source` Block

* `code_prefix` - (Required) Prefix under which the code `.zip` package is stored.
* `context_prefix` - (Optional) Prefix of business-context documents. Computed when omitted.
* `integration_id` - (Required) ID of the Amazon S3 integration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the environment.
* `collector_data_access_role_arn` - ARN of the role that Continuum assumes to read data from the associated collector.
* `created_at` - Time the environment was created.
* `id` - ID of the environment.
* `status` - Environment status.
* `status_message` - Additional detail about the status.
* `updated_at` - Time the environment was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_continuum_environment.example
  identity = {
    id = "01a17e22728b7d25b3a2294e0ca5c35e"
  }
}

resource "aws_continuum_environment" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) ID of the environment.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_continuum_environment` using the ID. For example:

```terraform
import {
  to = aws_continuum_environment.example
  id = "ID_PLACEHOLDER"
}
```

Using `terraform import`, import `aws_continuum_environment` using the ID. For example:

```console
% terraform import aws_continuum_environment.example ID_PLACEHOLDER
```
