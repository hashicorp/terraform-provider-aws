---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_role_membership"
description: |-
  Terraform resource for managing an AWS QuickSight Role Membership.
---
# Resource: aws_quicksight_role_membership

Terraform resource for managing an AWS QuickSight Role Membership.

~> The role membership APIs are disabled for identities managed by QuickSight. This resource can only be used when the QuickSight account subscription uses the Active Directory or IAM Identity Center authentication method.

## Example Usage

### Basic Usage

```terraform
resource "aws_quicksight_role_membership" "example" {
  member_name = "example-group"
  role        = "READER"
}
```

## Argument Reference

The following arguments are required:

* `member_name` - (Required) Name of the group to be added to the role.
* `role` - (Required) Role to add the group to. Valid values are `ADMIN`, `AUTHOR`, `READER`, `ADMIN_PRO`, `AUTHOR_PRO`, and `READER_PRO`.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID. Defaults to the account of the caller identity if not configured.
* `namespace` - (Required) Name of the namespace. Defaults to `default`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight Role Membership using a comma-delimited string combining the `aws_account_id`, `namespace`, `role`, and `member_name`. For example:

```terraform
import {
  to = aws_quicksight_role_membership.example
  id = "012345678901,default,READER,example-group"
}
```

Using `terraform import`, import QuickSight Role Membership using a comma-delimited string combining the `aws_account_id`, `namespace`, `role`, and `member_name`. For example:

```console
% terraform import aws_quicksight_role_membership.example 012345678901,default,READER,example-group
```
