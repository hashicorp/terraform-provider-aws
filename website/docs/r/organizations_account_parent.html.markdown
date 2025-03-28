---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_account_parent"
description: |-
  Manages an AWS Organizations Account's current parent in the Organization.
---
# Resource: aws_organizations_account_parent

Manages an AWS Organizations Account's current parent in the Organization.

~> **Note:** Account management must be done from the organization's management account.

!> **WARNING:** You should not use the `aws_organizations_account_parent` resources in conjunction with the `parent_id` field of the [`aws_organizations_account`](organizations_account.html) resource. Doing so may cause perpetual differences.

## Example Usage

### Basic Usage

```terraform
resource "aws_organizations_account_parent" "example" {
  account_id = "123456789012"
  parent_id  = "ou-c3iy-xbe65nlj"
}
```

## Argument Reference

The following arguments are required:

* `account_id` - (Required) AWS Account ID of the member account.
* `parent_id` - (Required) Parent Organizational Unit ID or Root ID for the account.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the account's existing parent association using the `account_id`. For example:

```terraform
import {
  to = aws_organizations_account_parent.example
  id = "123456789012"
}
```

Using `terraform import`, import Organizations Account Parent using the `account_id`. For example:

```console
% terraform import aws_organizations_account_parent.example 123456789012
```
