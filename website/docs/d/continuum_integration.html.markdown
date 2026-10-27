---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_integration"
description: |-
  Provides details about an AWS Continuum integration.
---

# Data Source: aws_continuum_integration

Provides details about an AWS Continuum integration.

## Example Usage

```terraform
data "aws_continuum_integration" "example" {
  name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Optional) ID of the integration. Exactly one of `id` or `name` is required.
* `name` - (Optional) Name of the integration.
* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the integration.
* `bucket` - Bucket of an Amazon S3 integration.
* `created_at` - Time the integration was created.
* `instance_url` - URL of the GitLab instance of a GitLab integration.
* `integration_type` - Type of the integration: `GITHUB`, `GITLAB` or `S3`.
* `organization_name` - GitHub organization of a GitHub integration.
* `role_arn` - Role ARN of an Amazon S3 integration.
* `target_path` - Target path of a GitLab integration.
* `target_type` - Target type of a GitLab integration.
* `updated_at` - Time the integration was last updated.

