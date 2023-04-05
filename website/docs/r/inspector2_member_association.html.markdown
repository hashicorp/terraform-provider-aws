---
subcategory: "Inspector V2"
layout: "aws"
page_title: "AWS: aws_inspector2_member_association"
description: |-
  Terraform resource for managing an AWS Inspector V2 Member Association.
---

# Resource: aws_inspector2_member_association

Terraform resource for associating accounts to existing Inspector2 instances.

## Example Usage

### Basic Usage

```terraform
resource "aws_inspector2_member_association" "example" {
  account_id = "012345678901"
}
```

## Argument Reference

The following arguments are required:

* `account_id` - (Required) ID of the account to associate

## Attributes Reference

In addition to all arguments above, the following attributes are exported:
