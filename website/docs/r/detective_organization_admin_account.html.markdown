---
subcategory: "Detective"
layout: "aws"
page_title: "AWS: aws_detective_organization_admin_account"
description: |-
  Manages a Detective Organization Admin Account
---

# Resource: aws_detective_organization_admin_account

Manages a Detective Organization Admin Account. The AWS account utilizing this resource must be an Organizations primary account. More information about Organizations support in Detective can be found in the [Detective User Guide](https://docs.aws.amazon.com/detective/latest/adminguide/accounts-orgs-transition.html).

## Example Usage

```terraform
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["detective.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_detective_organization_admin_account" "example" {
  depends_on = [aws_organizations_organization.example]

  account_id = "123456789012"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) AWS account identifier to designate as a delegated administrator for Detective.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS account identifier.

## Import

Detective Organization Admin Account can be imported using the AWS account ID, e.g.,

```
$ terraform import aws_detective_organization_admin_account.example 123456789012
```
