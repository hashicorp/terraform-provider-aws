---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_group"
description: |-
  Provides a Cognito User Group resource.
---

# Resource: aws_cognito_user_group

Provides a Cognito User Group resource.

## Example Usage

```terraform
resource "aws_cognito_user_pool" "main" {
  name = "identity pool"
}

data "aws_iam_policy_document" "group_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Federated"
      identifiers = ["cognito-identity.amazonaws.com"]
    }

    actions = ["sts:AssumeRoleWithWebIdentity"]

    condition {
      test     = "StringEquals"
      variable = "cognito-identity.amazonaws.com:aud"
      values   = ["us-east-1:12345678-dead-beef-cafe-123456790ab"]
    }

    condition {
      test     = "ForAnyValue:StringLike"
      variable = "cognito-identity.amazonaws.com:amr"
      values   = ["authenticated"]
    }
  }
}

resource "aws_iam_role" "group_role" {
  name               = "user-group-role"
  assume_role_policy = data.aws_iam_policy_document.group_role.json
}

resource "aws_cognito_user_group" "main" {
  name         = "user-group"
  user_pool_id = aws_cognito_user_pool.main.id
  description  = "Managed by Terraform"
  precedence   = 42
  role_arn     = aws_iam_role.group_role.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the user group.
* `user_pool_id` - (Required) The user pool ID.
* `description` - (Optional) The description of the user group.
* `precedence` - (Optional) The precedence of the user group.
* `role_arn` - (Optional) The ARN of the IAM role to be associated with the user group.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cognito User Groups using the `user_pool_id`/`name` attributes concatenated. For example:

```terraform
import {
  to = aws_cognito_user_group.group
  id = "us-east-1_vG78M4goG/user-group"
}
```

Using `terraform import`, import Cognito User Groups using the `user_pool_id`/`name` attributes concatenated. For example:

```console
% terraform import aws_cognito_user_group.group us-east-1_vG78M4goG/user-group
```
