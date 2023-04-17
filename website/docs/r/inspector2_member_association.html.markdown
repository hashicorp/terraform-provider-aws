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

The following arguments are required:

* `account_id` - (Required) ID of the account to associate

## Attributes Reference

No additional attributes are exported.
