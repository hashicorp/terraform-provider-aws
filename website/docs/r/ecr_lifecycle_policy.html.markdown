---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_lifecycle_policy"
description: |-
  Manages an ECR repository lifecycle policy.
---

# Resource: aws_ecr_lifecycle_policy

Manages an ECR repository lifecycle policy.

~> **NOTE:** Only one `aws_ecr_lifecycle_policy` resource can be used with the same ECR repository. To apply multiple rules, they must be combined in the `policy` JSON.

~> **NOTE:** The AWS ECR API seems to reorder rules based on `rulePriority`. If you define multiple rules that are not sorted in ascending `rulePriority` order in the Terraform code, the resource will be flagged for recreation every `terraform plan`.

## Example Usage

### Policy on untagged image

```terraform
resource "aws_ecr_repository" "example" {
  name = "example-repo"
}

resource "aws_ecr_lifecycle_policy" "example" {
  repository = aws_ecr_repository.example.name

  policy = <<EOF
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
EOF
}
```

### Policy on tagged image

```terraform
resource "aws_ecr_repository" "example" {
  name = "example-repo"
}

resource "aws_ecr_lifecycle_policy" "example" {
  repository = aws_ecr_repository.example.name

  policy = <<EOF
{
    "rules": [
        {
            "rulePriority": 1,
            "description": "Keep last 30 images",
            "selection": {
                "tagStatus": "tagged",
                "tagPrefixList": ["v"],
                "countType": "imageCountMoreThan",
                "countNumber": 30
            },
            "action": {
                "type": "expire"
            }
        }
    ]
}
EOF
}
```

## Argument Reference

This resource supports the following arguments:

* `repository` - (Required) Name of the repository to apply the policy.
* `policy` - (Required) The policy document. This is a JSON formatted string. See more details about [Policy Parameters](http://docs.aws.amazon.com/AmazonECR/latest/userguide/LifecyclePolicies.html#lifecycle_policy_parameters) in the official AWS docs. Consider using the [`aws_ecr_lifecycle_policy_document` data_source](/docs/providers/aws/d/ecr_lifecycle_policy_document.html) to generate/manage the JSON document used for the `policy` argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `repository` - The name of the repository.
* `registry_id` - The registry ID where the repository was created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECR Lifecycle Policy using the name of the repository. For example:

```terraform
import {
  to = aws_ecr_lifecycle_policy.example
  id = "tf-example"
}
```

Using `terraform import`, import ECR Lifecycle Policy using the name of the repository. For example:

```console
% terraform import aws_ecr_lifecycle_policy.example tf-example
```
