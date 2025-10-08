---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_user"
description: |-
  Provides an IAM user.
---

# Resource: aws_iam_user

Provides an IAM user.

~> *NOTE:* If policies are attached to the user via the [`aws_iam_policy_attachment` resource](/docs/providers/aws/r/iam_policy_attachment.html) and you are modifying the user `name` or `path`, the `force_destroy` argument must be set to `true` and applied before attempting the operation otherwise you will encounter a `DeleteConflict` error. The [`aws_iam_user_policy_attachment` resource (recommended)](/docs/providers/aws/r/iam_user_policy_attachment.html) does not have this requirement.

## Example Usage

```terraform
resource "aws_iam_user" "lb" {
  name = "loadbalancer"
  path = "/system/"

  tags = {
    tag-key = "tag-value"
  }
}

resource "aws_iam_access_key" "lb" {
  user = aws_iam_user.lb.name
}

data "aws_iam_policy_document" "lb_ro" {
  statement {
    effect    = "Allow"
    actions   = ["ec2:Describe*"]
    resources = ["*"]
  }
}

resource "aws_iam_user_policy" "lb_ro" {
  name   = "test"
  user   = aws_iam_user.lb.name
  policy = data.aws_iam_policy_document.lb_ro.json
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The user's name. The name must consist of upper and lowercase alphanumeric characters with no spaces. You can also include any of the following characters: `=,.@-_.`. User names are not distinguished by case. For example, you cannot create users named both "TESTUSER" and "testuser".
* `path` - (Optional, default "/") Path in which to create the user.
* `permissions_boundary` - (Optional) The ARN of the policy that is used to set the permissions boundary for the user.
* `force_destroy` - (Optional, default false) When destroying this user, destroy even if it
  has non-Terraform-managed IAM access keys, login profile or MFA devices. Without `force_destroy`
  a user with non-Terraform-managed access keys and login profile will fail to be destroyed.
* `tags` - Key-value map of tags for the IAM user. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN assigned by AWS for this user.
* `id` - The user's name.
* `name` - The user's name.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `unique_id` - The [unique ID][1] assigned by AWS.

  [1]: https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html#GUIDs

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM Users using the `name`. For example:

```terraform
import {
  to = aws_iam_user.lb
  id = "loadbalancer"
}
```

Using `terraform import`, import IAM Users using the `name`. For example:

```console
% terraform import aws_iam_user.lb loadbalancer
```
