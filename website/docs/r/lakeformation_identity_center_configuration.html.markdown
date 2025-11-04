---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_identity_center_configuration"
description: |-
  Manages an AWS Lake Formation Identity Center Configuration.
---

# Resource: aws_lakeformation_identity_center_configuration

Manages an AWS Lake Formation Identity Center Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_lakeformation_identity_center_configuration" "example" {
  instance_arn = local.identity_center_instance_arn
}

locals {
  identity_center_instance_arn = data.aws_ssoadmin_instances.example.arns[0]
}

data "aws_ssoadmin_instances" "example" {}
```

## Argument Reference

The following arguments are required:

* `instance_arn` - (Required) ARN of the IAM Identity Center Instance to associate.

The following arguments are optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog.
  By default, the account ID.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `application_arn` - ARN of the Lake Formation applicated integrated with IAM Identity Center.
* `resource_share` - ARN of the Resource Access Manager (RAM) resource share.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lake Formation Identity Center Configuration using the `catalog_id`. For example:

```terraform
import {
  to = aws_lakeformation_identity_center_configuration.example
  id = "123456789012"
}
```

Using `terraform import`, import Lake Formation Identity Center Configuration using the `catalog_id`. For example:

```console
% terraform import aws_lakeformation_identity_center_configuration.example 123456789012
```
