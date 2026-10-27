---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_environment"
description: |-
  Provides details about an AWS Continuum environment.
---

# Data Source: aws_continuum_environment

Provides details about an AWS Continuum environment.

## Example Usage

```terraform
data "aws_continuum_environment" "example" {
  name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Optional) ID of the environment. Exactly one of `id` or `name` is required.
* `name` - (Optional) Name of the environment.
* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `allowed_domains` - Set of allowed domains.
* `arn` - ARN of the environment.
* `collector_data_access_role_arn` - ARN of the collector data access role.
* `created_at` - Time the environment was created.
* `patch_creation_policy` - Patch creation policy.
* `repository_source` - Repository sources. Each has `default_branch`, `integration_id` and `repository`.
* `s3_source` - Amazon S3 sources. Each has `code_prefix`, `context_prefix` and `integration_id`.
* `scan_frequency` - Scan frequency.
* `status` - Environment status.
* `status_message` - Additional detail about the status.
* `updated_at` - Time the environment was last updated.

