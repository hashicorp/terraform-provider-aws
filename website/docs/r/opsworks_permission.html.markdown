---
subcategory: "OpsWorks"
layout: "aws"
page_title: "AWS: aws_opsworks_permission"
description: |-
  Provides an OpsWorks permission resource.
---

# Resource: aws_opsworks_permission

Provides an OpsWorks permission resource.

## Example Usage

```terraform
resource "aws_opsworks_permission" "my_stack_permission" {
  allow_ssh  = true
  allow_sudo = true
  level      = "iam_only"
  user_arn   = aws_iam_user.user.arn
  stack_id   = aws_opsworks_stack.stack.id
}
```

## Argument Reference

This resource supports the following arguments:

* `allow_ssh` - (Optional) Whether the user is allowed to use SSH to communicate with the instance
* `allow_sudo` - (Optional) Whether the user is allowed to use sudo to elevate privileges
* `user_arn` - (Required) The user's IAM ARN to set permissions for
* `level` - (Optional) The users permission level. Mus be one of `deny`, `show`, `deploy`, `manage`, `iam_only`
* `stack_id` - (Required) The stack to set the permissions for

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The computed id of the permission. Please note that this is only used internally to identify the permission. This value is not used in aws.
