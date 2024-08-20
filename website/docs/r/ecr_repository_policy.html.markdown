---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_repository_policy"
description: |-
  Provides an Elastic Container Registry Repository Policy.
---

# Resource: aws_ecr_repository_policy

Provides an Elastic Container Registry Repository Policy.

Note that currently only one policy may be applied to a repository.

## Example Usage

```terraform
resource "aws_ecr_repository" "example" {
  name = "example-repo"
}

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

resource "aws_ecr_repository_policy" "example" {
  repository = aws_ecr_repository.example.name
  policy     = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `repository` - (Required) Name of the repository to apply the policy.
* `policy` - (Required) The policy document. This is a JSON formatted string. For more information about building IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `repository` - The name of the repository.
* `registry_id` - The registry ID where the repository was created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECR Repository Policy using the repository name. For example:

```terraform
import {
  to = aws_ecr_repository_policy.example
  id = "example"
}
```

Using `terraform import`, import ECR Repository Policy using the repository name. For example:

```console
% terraform import aws_ecr_repository_policy.example example
```
