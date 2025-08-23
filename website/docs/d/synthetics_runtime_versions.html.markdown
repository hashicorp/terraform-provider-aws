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
data "aws_synthetics_runtime_versions" "example" {}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Name of the AWS region from which runtime versions are fetched.
* `runtime_versions` - List of runtime versions. See [`runtime_versions` attribute reference](#runtime_versions-attribute-reference).

### `runtime_versions` Attribute Reference

* `deprecation_date` - Date of deprecation if the runtme version is deprecated.
* `description` - Description of the runtime version, created by Amazon.
* `release_date` - Date that the runtime version was released.
* `version_name` - Name of the runtime version.
For a list of valid runtime versions, see [Canary Runtime Versions](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Synthetics_Canaries_Library.html).
