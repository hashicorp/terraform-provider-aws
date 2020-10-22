---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_sso_permission_set"
description: |-
  Get information on an AWS Single Sign-On Permission Set.
---

# Data Source: aws_sso_permission_set

Use this data source to get the Single Sign-On Permission Set.

## Example Usage

```hcl
data "aws_sso_instance" "selected" {}

data "aws_sso_permission_set" "example" {
  instance_arn = data.aws_sso_instance.selected.arn
  name         = "Example"
}

output "arn" {
  value = data.aws_sso_permission_set.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `instance_arn` - (Required) The AWS ARN associated with the AWS Single Sign-On Instance.
* `name` - (Required) The name of the AWS Single Sign-On Permission Set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The arn of the permission set.
* `arn` - The arn of the permission set.
* `created_date` - The created date of the permission set.
* `description` - The description of the permission set.
* `session_duration` - The session duration of the permission set.
* `relay_state` - The relay state of the permission set.
* `inline_policy` - The inline policy of the permission set.
* `managed_policy_arns` - The managed policies attached to the permission set.
* `tags` - The tags of the permission set.