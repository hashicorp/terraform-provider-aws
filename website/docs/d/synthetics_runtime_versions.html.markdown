---
subcategory: "CloudWatch Synthetics"
layout: "aws"
page_title: "AWS: aws_synthetics_runtime_versions"
description: |-
  Terraform data source for managing an AWS CloudWatch Synthetics Runtime Versions.
---

# Data Source: aws_synthetics_runtime_versions

Terraform data source for managing an AWS CloudWatch Synthetics Runtime Versions.

## Example Usage

### Basic Usage

```terraform
data "aws_synthetics_runtime_versions" "example" {
  skip_deprecated = true
}
```

## Argument Reference

The following arguments are optional:

* `skip_deprecated` - (Optional) Whether deprecated runtime versions should be skipped. Default to `false`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Name of the AWS region from which runtime versions are fetched.
* `version_names` - List of runtime version names.
