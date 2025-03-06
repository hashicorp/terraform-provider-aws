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

### Create User With IAM Identity Type Using an IAM Role

```terraform
resource "aws_quicksight_user" "example" {
  email         = "author1@example.com"
  identity_type = "IAM"
  user_role     = "AUTHOR"
  iam_arn       = "arn:aws:iam::123456789012:role/AuthorRole"
  session_name  = "author1"
}
```

### Create User With IAM Identity Type Using an IAM User

```terraform
resource "aws_quicksight_user" "example" {
  email         = "authorpro1@example.com"
  identity_type = "IAM"
  user_role     = "AUTHOR_PRO"
  iam_arn       = "arn:aws:iam::123456789012:user/authorpro1"
}
```

### Create User With QuickSight Identity Type in Non-Default Namespace

```terraform
resource "aws_quicksight_user" "example" {
  email         = "reader1@example.com"
  identity_type = "QUICKSIGHT"
  user_role     = "READER"
  namespace     = "example"
  user_name     = "reader1"
}
```

## Argument Reference

The following arguments are required:

* `email` - (Required) Email address of the user that you want to register.
* `identity_type` - (Required) Identity type that your Amazon QuickSight account uses to manage the identity of users. Valid values: `IAM`, `QUICKSIGHT`.
* `user_role` - (Required) Amazon QuickSight role for the user. Value values: `READER`, `AUTHOR`, `ADMIN`, `READER_PRO`, `AUTHOR_PRO`, `ADMIN_PRO`.

The following arguments are optional:

* `aws_account_id` - (Optional) ID for the AWS account that the user is in. Use the ID for the AWS account that contains your Amazon QuickSight account.
* `iam_arn` - (Optional) ARN of the IAM user or role that you are registering with Amazon QuickSight. Required only for users with an identity type of `IAM`.
* `namespace`  - (Optional) The Amazon Quicksight namespace to create the user in. Defaults to `default`.
* `session_name` - (Optional) Name of the IAM session to use when assuming roles that can embed QuickSight dashboards. Only valid for registering users using an assumed IAM role. Additionally, if registering multiple users using the same IAM role, each user needs to have a unique session name.
* `user_name` - (Optional) Amazon QuickSight user name that you want to create for the user you are registering. Required only for users with an identity type of `QUICKSIGHT`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` -  Amazon Resource Name (ARN) for the user.
* `id` - Unique identifier consisting of the account ID, the namespace, and the user name separated by `/`s.
* `user_invitation_url` - URL the user visits to complete registration and provide a password. Returned only for users with an identity type of `QUICKSIGHT`.

## Import

You cannot import this resource.
