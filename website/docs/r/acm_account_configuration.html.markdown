---
subcategory: "ACM (Certificate Manager)"
layout: "aws"
page_title: "AWS: aws_acm_account_configuration"
description: |-
  Manages AWS ACM (Certificate Manager) Account Configuration for certificate expiry events.
---

# Resource: aws_acm_account_configuration

Manages AWS ACM (Certificate Manager) Account Configuration. This resource allows you to configure account-level settings for ACM, such as certificate expiry event notifications.

~> Each AWS account may only have one ACM Account Configuration per region. Multiple configurations of the resource against the same AWS account and region will cause perpetual differences.

~> Deletion of this resource will not modify any settings, only remove the resource from the state.

## Example Usage

### Basic Usage

```terraform
resource "aws_acm_account_configuration" "example" {
  expiry_events {
    days_before_expiry = 14
  }
}
```

## Argument Reference

The following arguments are required:

* `expiry_events` - (Required) Configuration block for certificate expiry events. See [expiry_events](#expiry_events-block) for details.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `expiry_events` Block

The `expiry_events` configuration block supports the following arguments:

* `days_before_expiry` - (Optional) Number of days before certificate expiry to send notifications. Must be between `1` and `45`. Defaults to `45`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS account ID where the ACM account configuration is managed.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ACM Account Configuration using the `id`. For example:

```terraform
import {
  to = aws_acm_account_configuration.example
  id = "123456789012"
}
```

Using `terraform import`, import ACM Account Configuration using the `id`. For example:

```console
% terraform import aws_acm_account_configuration.example 123456789012
```
