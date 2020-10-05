---
subcategory: "Identity Store"
layout: "aws"
page_title: "AWS: aws_identity_store_group"
description: |-
  Get information on an AWS Identity Store Group
---

# Data Source: aws_identity_store_group

Use this data source to get an Identity Store Group.

## Example Usage

```hcl
data "aws_sso_instance" "selected" {}

data "aws_identity_store_group" "example" {
  identity_store_id = data.aws_sso_instance.selected.identity_store_id
  display_name      = "ExampleGroup@example.com"
}

output "group_id" {
  value = data.aws_identity_store_group.example.group_id
}
```

## Argument Reference

The following arguments are supported:

* `identity_store_id` - (Required) The Identity Store ID associated with the AWS Single Sign-On Instance.
* `group_id` - (Optional) An Identity Store group ID.
* `display_name` - (Optional) An Identity Store group display name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Identity Store group ID.
* `group_id` - The Identity Store group ID.
* `display_name` - The Identity Store group display name.
