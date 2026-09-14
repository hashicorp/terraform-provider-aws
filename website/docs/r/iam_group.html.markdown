---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_group"
description: |-
  Provides an IAM group.
---

# Resource: aws_iam_group

Provides an IAM group.

~> **NOTE on user management:** Using `aws_iam_group_membership` or `aws_iam_user_group_membership` resources in addition to manually managing user/group membership using the console may lead to configuration drift or conflicts. For this reason, it's recommended to either manage membership entirely with Terraform or entirely within the AWS console.

## Example Usage

```terraform
resource "aws_iam_group" "developers" {
  name = "developers"
  path = "/users/"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Group's name. Must consist of upper and lowercase alphanumeric characters with no spaces. You can also include any of the following characters: `=,.@-_.`. Group names are not distinguished by case. For example, you cannot create groups named both "ADMINS" and "admins".
* `path` - (Optional, default "/") Path in which to create the group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN assigned by AWS for this group.
* `id` - Group's name.
* `name` - Group's name.
* `path` - Path of the group in IAM.
* `unique_id` - [Unique ID](https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html#GUIDs) assigned by AWS.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM Groups using the `name`. For example:

```terraform
import {
  to = aws_iam_group.developers
  id = "developers"
}
```

Using `terraform import`, import IAM Groups using the `name`. For example:

```console
% terraform import aws_iam_group.developers developers
```
