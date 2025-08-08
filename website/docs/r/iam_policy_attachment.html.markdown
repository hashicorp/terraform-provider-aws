---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_policy_attachment"
description: |-
  Attaches a Managed IAM Policy to user(s), role(s), and/or group(s)
---

# Resource: aws_iam_policy_attachment

Attaches a Managed IAM Policy to user(s), role(s), and/or group(s)

!> **WARNING:** The aws_iam_policy_attachment resource creates **exclusive** attachments of IAM policies. Across the entire AWS account, all of the users/roles/groups to which a single policy is attached must be declared by a single aws_iam_policy_attachment resource. This means that even any users/roles/groups that have the attached policy via any other mechanism (including other Terraform resources) will have that attached policy revoked by this resource. Consider `aws_iam_role_policy_attachment`, `aws_iam_user_policy_attachment`, or `aws_iam_group_policy_attachment` instead. These resources do not enforce exclusive attachment of an IAM policy.

~> **NOTE:** The usage of this resource conflicts with the `aws_iam_group_policy_attachment`, `aws_iam_role_policy_attachment`, and `aws_iam_user_policy_attachment` resources and will permanently show a difference if both are defined.

~> **NOTE:** For a given role, this resource is incompatible with using the [`aws_iam_role` resource](/docs/providers/aws/r/iam_role.html) `managed_policy_arns` argument. When using that argument and this resource, both will attempt to manage the role's managed policy attachments and Terraform will show a permanent difference.

~> **NOTE:** To ensure Terraform correctly manages dependencies during updates, use a reference to the IAM resource when defining the `policy_arn` for `aws_iam_policy_attachment`, rather than constructing the ARN directly. For example, use `policy_arn = aws_iam_policy.example.arn` instead of `policy_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/Example"`. Failing to do so may lead to errors like `DeleteConflict: Cannot delete a policy attached to entities` or `NoSuchEntity`.

## Example Usage

```terraform
resource "aws_iam_user" "user" {
  name = "test-user"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "role" {
  name               = "test-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_group" "group" {
  name = "test-group"
}

data "aws_iam_policy_document" "policy" {
  statement {
    effect    = "Allow"
    actions   = ["ec2:Describe*"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "policy" {
  name        = "test-policy"
  description = "A test policy"
  policy      = data.aws_iam_policy_document.policy.json
}

resource "aws_iam_policy_attachment" "test-attach" {
  name       = "test-attachment"
  users      = [aws_iam_user.user.name]
  roles      = [aws_iam_role.role.name]
  groups     = [aws_iam_group.group.name]
  policy_arn = aws_iam_policy.policy.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `name`    (Required) - Name of the attachment. This cannot be an empty string.
* `users`   (Optional) - User(s) the policy should be applied to.
* `roles`   (Optional) - Role(s) the policy should be applied to.
* `groups`  (Optional) - Group(s) the policy should be applied to.
* `policy_arn`  (Required) - ARN of the policy you want to apply. Typically this should be a reference to the ARN of another resource to ensure dependency ordering, such as `aws_iam_policy.example.arn`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Policy's ID.
* `name` - Name of the attachment.
