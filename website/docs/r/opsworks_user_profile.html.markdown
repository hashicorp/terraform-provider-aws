---
subcategory: "OpsWorks"
layout: "aws"
page_title: "AWS: aws_opsworks_user_profile"
description: |-
  Provides an OpsWorks User Profile resource.
---

# Resource: aws_opsworks_user_profile

Provides an OpsWorks User Profile resource.

## Example Usage

```terraform
resource "aws_opsworks_user_profile" "my_profile" {
  user_arn     = aws_iam_user.user.arn
  ssh_username = "my_user"
}
```

## Argument Reference

This resource supports the following arguments:

* `user_arn` - (Required) The user's IAM ARN
* `allow_self_management` - (Optional) Whether users can specify their own SSH public key through the My Settings page
* `ssh_username` - (Required) The ssh username, with witch this user wants to log in
* `ssh_public_key` - (Optional) The users public key

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Same value as `user_arn`
