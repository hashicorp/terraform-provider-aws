---
subcategory: "Identity Store"
layout: "aws"
page_title: "AWS: aws_identity_store_user"
description: |-
  Get information on an AWS Identity Store User
---

# Data Source: aws_identity_store_user

Use this data source to get an Identity Store User.

## Example Usage

```hcl
data "aws_sso_instance" "selected" {}

data "aws_identity_store_user" "example" {
  identity_store_id = data.aws_sso_instance.selected.identity_store_id
  user_name         = "example@example.com"
}

output "user_id" {
  value = data.aws_identity_store_user.example.user_id
}
```

## Argument Reference

The following arguments are supported:

* `identity_store_id` - (Required) The Identity Store ID associated with the AWS Single Sign-On Instance.
* `user_id` - (Optional) An Identity Store user ID.
* `user_name` - (Optional) An Identity Store user name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Identity Store user ID.
* `user_id` - The Identity Store user ID.
* `user_name` - The Identity Store user name.
