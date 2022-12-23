---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_organization_admin_account"
description: |-
  Enables the IPAM Service and promotes an account to delegated administrator for the service.
---

# Resource: aws_vpc_ipam_organization_admin_account

Enables the IPAM Service and promotes a delegated administrator.

## Example Usage

Basic usage:

```terraform
resource "aws_vpc_ipam_organization_admin_account" "example" {
  delegated_admin_account_id = data.aws_caller_identity.delegated.account_id
}

data "aws_caller_identity" "delegated" {
  provider = aws.ipam_delegate_account
}

provider "aws" {
  alias = "ipam_delegate_account"

  # authentication arguments omitted
}
```

## Argument Reference

The following arguments are supported:

* `delegated_admin_account_id` - (Required)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Organizations ARN for the delegate account.
* `id` - The Organizations member account ID that you want to enable as the IPAM account.
* `email` - The Organizations email for the delegate account.
* `name` - The Organizations name for the delegate account.
* `service_principal` - The AWS service principal.

## Import

IPAMs can be imported using the `delegate account id`, e.g.

```
$ terraform import aws_vpc_ipam_organization_admin_account.example 12345678901
```
