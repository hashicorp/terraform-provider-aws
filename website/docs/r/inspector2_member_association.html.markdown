---
subcategory: "Inspector"
layout: "aws"
page_title: "AWS: aws_inspector2_member_association"
description: |-
  Terraform resource for managing an Amazon Inspector Member Association.
---

# Resource: aws_inspector2_member_association

Terraform resource for associating accounts to existing Inspector instances.

## Example Usage

### Basic Usage

```terraform
resource "aws_inspector2_member_association" "example" {
  account_id = "123456789012"
}
```

## Argument Reference

The following argument is required:

* `account_id` - (Required) ID of the account to associate

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `delegated_admin_account_id` - Account ID of the delegated administrator account
* `relationship_status` - Status of the member relationship
* `updated_at` - Date and time of the last update of the relationship

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Inspector Member Association using the `account_id`. For example:

```terraform
import {
  to = aws_inspector2_member_association.example
  id = "123456789012"
}
```

Using `terraform import`, import Amazon Inspector Member Association using the `account_id`. For example:

```console
% terraform import aws_inspector2_member_association.example 123456789012
```
