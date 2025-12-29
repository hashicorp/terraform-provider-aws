---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_pull_time_update_exclusion"
description: |-
  Manages an AWS ECR (Elastic Container Registry) Pull Time Update Exclusion.
---

# Resource: aws_ecr_pull_time_update_exclusion

Manages an AWS ECR (Elastic Container Registry) Pull Time Update Exclusion.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "example" {
  name = "example-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "example" {
  name = "example-role-policy"
  role = aws_iam_role.example.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_ecr_pull_time_update_exclusion" "example" {
  principal_arn = aws_iam_role.example.arn
}
```

### With IAM User

```terraform
resource "aws_iam_user" "example" {
  name = "example-user"
}

resource "aws_iam_user_policy" "example" {
  name = "example-user-policy"
  user = aws_iam_user.example.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_ecr_pull_time_update_exclusion" "example" {
  principal_arn = aws_iam_user.example.arn
}
```

## Argument Reference

The following arguments are required:

* `principal_arn` - (Required, Forces new resource) ARN of the IAM principal to exclude from having image pull times recorded.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ARN of the IAM principal to exclude from having image pull times recorded.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECR (Elastic Container Registry) Pull Time Update Exclusion using the `principal_arn`. For example:

```terraform
import {
  to = aws_ecr_pull_time_update_exclusion.example
  id = "arn:aws:iam::123456789012:role/example-role"
}
```

Using `terraform import`, import ECR (Elastic Container Registry) Pull Time Update Exclusion using the `principal_arn`. For example:

```console
% terraform import aws_ecr_pull_time_update_exclusion.example arn:aws:iam::123456789012:role/example-role
```
