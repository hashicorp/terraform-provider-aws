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

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `delegated_admin_account_id` - Account ID of the delegated administrator account
* `relationship_status` - Status of the member relationship
* `updated_at` - Date and time of the last update of the relationship

## Import

Amazon Inspector Member Association can be imported using the `account_id`, e.g.,

```
$ terraform import aws_inspector2_member_association.example 123456789012
```
