---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_repository_creation_template"
description: |-
  Provides an Elastic Container Registry Repository Creation Template.
---

# Resource: aws_ecr_repository_creation_template

Provides an Elastic Container Registry Repository Creation Template.

## Example Usage

```terraform
data "aws_iam_policy_document" "example" {
  statement {
    sid    = "new policy"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["123456789012"]
    }

    actions = [
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "ecr:BatchCheckLayerAvailability",
      "ecr:PutImage",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload",
      "ecr:DescribeRepositories",
      "ecr:GetRepositoryPolicy",
      "ecr:ListImages",
      "ecr:DeleteRepository",
      "ecr:BatchDeleteImage",
      "ecr:SetRepositoryPolicy",
      "ecr:DeleteRepositoryPolicy",
    ]
  }
}

resource "aws_ecr_repository_creation_template" "example" {
  prefix               = "example"
  description          = "An example template"
  image_tag_mutability = "IMMUTABLE"
  custom_role_arn      = "arn:aws:iam::123456789012:role/example"

  applied_for = [
    "PULL_THROUGH_CACHE",
  ]

  encryption_configuration {
    encryption_type = "AES256"
  }

  repository_policy = data.aws_iam_policy_document.example.json

  lifecycle_policy = <<EOT
{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Expire images older than 14 days",
      "selection": {
        "tagStatus": "untagged",
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 14
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}
EOT

  resource_tags = {
    Foo = "Bar"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `prefix` - (Required, Forces new resource) The repository name prefix to match against. Use `ROOT` to match any prefix that doesn't explicitly match another template.
* `applied_for` - (Required) Which features this template applies to. Must contain one or more of `PULL_THROUGH_CACHE` or `REPLICATION`.
* `custom_role_arn` - (Optional) A custom IAM role to use for repository creation. Required if using repository tags or KMS encryption.
* `description` - (Optional) The description for this template.
* `encryption_configuration` - (Optional) Encryption configuration for any created repositories. See [below for schema](#encryption_configuration).
* `image_tag_mutability` - (Optional) The tag mutability setting for any created repositories. Must be one of: `MUTABLE` or `IMMUTABLE`. Defaults to `MUTABLE`.
* `lifecycle_policy` - (Optional) The lifecycle policy document to apply to any created repositories. See more details about [Policy Parameters](http://docs.aws.amazon.com/AmazonECR/latest/userguide/LifecyclePolicies.html#lifecycle_policy_parameters) in the official AWS docs. Consider using the [`aws_ecr_lifecycle_policy_document` data_source](/docs/providers/aws/d/ecr_lifecycle_policy_document.html) to generate/manage the JSON document used for the `lifecycle_policy` argument.
* `repository_policy` - (Optional) The registry policy document to apply to any created repositories. This is a JSON formatted string. For more information about building IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).
* `resource_tags` - (Optional) A map of tags to assign to any created repositories.

### encryption_configuration

* `encryption_type` - (Optional) The encryption type to use for any created repositories. Valid values are `AES256` or `KMS`. Defaults to `AES256`.
* `kms_key` - (Optional) The ARN of the KMS key to use when `encryption_type` is `KMS`. If not specified, uses the default AWS managed key for ECR.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `registry_id` - The registry ID the repository creation template applies to.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECR Repository Creation Templates using the `prefix`. For example:

```terraform
import {
  to = aws_ecr_repository_creation_template.example
  id = "example"
}
```

Using `terraform import`, import the ECR Repository Creating Templates using the `prefix`. For example:

```console
% terraform import aws_ecr_repository_creation_template.example example
```
