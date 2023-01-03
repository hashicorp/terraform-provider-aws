---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_user"
description: |-
  Manages a Resource QuickSight User.
---

# Resource: aws_quicksight_user

Resource for managing QuickSight User

## Example Usage

```terraform
resource "aws_quicksight_user" "example" {
  session_name  = "an-author"
  email         = "author@example.com"
  namespace     = "foo"
  identity_type = "IAM"
  iam_arn       = "arn:aws:iam::123456789012:user/Example"
  user_role     = "AUTHOR"
}
```

## Argument Reference

The following arguments are supported:

* `email` - (Required) The email address of the user that you want to register.
* `identity_type` - (Required) Amazon QuickSight supports several ways of managing the identity of users. This parameter accepts either  `IAM` or `QUICKSIGHT`. If `IAM` is specified, the `iam_arn` must also be specified.
* `user_role` - (Required) The Amazon QuickSight role of the user. The user role can be one of the following: `READER`, `AUTHOR`, or `ADMIN`
* `user_name` - (Optional) The Amazon QuickSight user name that you want to create for the user you are registering. Only valid for registering a user with `identity_type` set to `QUICKSIGHT`.
* `aws_account_id` - (Optional) The ID for the AWS account that the user is in. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.
* `iam_arn` - (Optional) The ARN of the IAM user or role that you are registering with Amazon QuickSight.
* `namespace`  - (Optional) The Amazon Quicksight namespace to create the user in. Defaults to `default`.
* `session_name` - (Optional) The name of the IAM session to use when assuming roles that can embed QuickSight dashboards. Only valid for registering users using an assumed IAM role. Additionally, if registering multiple users using the same IAM role, each user needs to have a unique session name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the user

## Import

Importing is currently not supported on this resource.
