---
subcategory: "CloudWatch Synthetics"
layout: "aws"
page_title: "AWS: aws_synthetics_runtime_version"
description: |-
  Terraform data source for managing an AWS CloudWatch Synthetics Runtime Version.
---
# Data Source: aws_synthetics_runtime_version

Terraform data source for managing an AWS CloudWatch Synthetics Runtime Version.

## Example Usage

### Latest Runtime Version

```terraform
data "aws_synthetics_runtime_version" "example" {
  prefix = "syn-nodejs-puppeteer"
  latest = true
}
```

### Specific Runtime Version

```terraform
data "aws_synthetics_runtime_version" "example" {
  prefix  = "syn-nodejs-puppeteer"
  version = "9.0"
}
```

## Argument Reference

The following arguments are required:

* `prefix` - (Required) Name prefix of the runtime version (for example, `syn-nodejs-puppeteer`).

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `latest` - (Optional) Whether the latest version of the runtime should be fetched. Conflicts with `version`. Valid values: `true`.
* `version` - (Optional) Version of the runtime to be fetched (for example, `9.0`). Conflicts with `latest`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `deprecation_date` - Date of deprecation if the runtme version is deprecated.
* `description` - Description of the runtime version, created by Amazon.
* `id` - Name of the runtime version. For a list of valid runtime versions, see [Canary Runtime Versions](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Synthetics_Canaries_Library.html).
* `release_date` - Date that the runtime version was released.
* `version_name` - Name of the runtime version. For a list of valid runtime versions, see [Canary Runtime Versions](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Synthetics_Canaries_Library.html).
