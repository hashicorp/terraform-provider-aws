---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_role"
description: |-
  Get information on a Single Sign-On (SSO) Role.
---

# Data Source: aws_ssoadmin_role

Use this data source to get a Single Sign-On (SSO) Role.

## Example Usage

```hcl
data "aws_ssoadmin_instances" "example" {}

data "aws_ssoadmin_permission_set" "example" {
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  name         = "AWSReadOnlyAccess"
}

data "aws_identitystore_group" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]

  filter {
    attribute_path  = "DisplayName"
    attribute_value = "ExampleGroup"
  }
}

resource "aws_ssoadmin_account_assignment" "example" {
  instance_arn       = data.aws_ssoadmin_permission_set.example.instance_arn
  permission_set_arn = data.aws_ssoadmin_permission_set.example.arn

  principal_id   = data.aws_identitystore_group.example.group_id
  principal_type = "GROUP"

  target_id   = "012347678910"
  target_type = "AWS_ACCOUNT"
}

data "aws_ssoadmin_role" "example" {
  permission_set_name = data.aws_ssoadmin_permission_set.example.name
}

output "arn" {
  value = data.aws_ssoadmin_role.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `permission_set_name` - (Required) The name of the SSO Permission Set the role was created for.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The friendly IAM role name to match.
* `arn` - The Amazon Resource Name (ARN) specifying the role.
* `assume_role_policy` - The policy document associated with the role.
* `create_date` - Creation date of the role in RFC 3339 format.
* `description` - Description for the role.
* `max_session_duration` - Maximum session duration.
* `path` - The path to the role.
* `permissions_boundary` - The ARN of the policy that is used to set the permissions boundary for the role.
* `unique_id` - The stable and unique string identifying the role.
* `tags` - The tags attached to the role.
