---
subcategory: "User Experience Customization"
layout: "aws"
page_title: "AWS: aws_uxc_account_customizations"
description: |-
  Manages UXC Account Customizations for an AWS Account.
---

# Resource: aws_uxc_account_customizations

Manages UXC Account Customizations for an AWS Account. This resource controls the console experience for the account, including the account color and which AWS regions and services are visible in the AWS Management Console.

~> **Note:** There is only a single set of account customizations per AWS account.

~> **Note:** This resource operates globally and always targets the `us-east-1` region regardless of the provider region configuration.

~> **Note:** The UXC API does not provide a delete operation. Destroying this resource resets all customizations to their defaults: `account_color` is set to `none`, and both `visible_regions` and `visible_services` are cleared to allow all regions and services.

## Example Usage

### Set Account Color

```terraform
resource "aws_uxc_account_customizations" "example" {
  account_color = "lightBlue"
}
```

### Restrict Visible Regions and Services

```terraform
resource "aws_uxc_account_customizations" "example" {
  account_color    = "green"
  visible_regions  = ["us-east-1", "us-west-2", "eu-west-1"]
  visible_services = ["ec2", "s3", "rds"]
}
```

## Argument Reference

This resource supports the following arguments:

* `account_color` - (Optional) Color used to identify the account in the AWS Management Console. Valid values are `none`, `red`, `darkBlue`, `lightBlue`, `green`, `yellow`, `orange`, `pink`, `purple`, and `teal`. Defaults to `none`.
* `visible_regions` - (Optional) Set of AWS region codes to display in the console. When omitted or empty, all regions are visible.
* `visible_services` - (Optional) Set of AWS service identifiers to display in the console. When omitted or empty, all services are visible.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to       = aws_uxc_account_customizations.example
  identity = {}
}

resource "aws_uxc_account_customizations" "example" {}
```

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import UXC Account Customizations using the AWS account ID. For example:

```terraform
import {
  to = aws_uxc_account_customizations.example
  id = "123456789012"
}
```

Using `terraform import`, import UXC Account Customizations using the AWS account ID. For example:

```console
% terraform import aws_uxc_account_customizations.example 123456789012
```
