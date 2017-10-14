---
layout: "aws"
page_title: "AWS: aws_ecr_lifecycle_policy"
sidebar_current: "docs-aws-resource-ecr-lifecycle-policy"
description: |-
  Provides an ECR Lifecycle Policy.
---

# aws_ecr_ifecycle_policy

Provides an ECR lifecycle policy.

## Example Usage

```hcl
resource "aws_ecr_repository" "foo" {
  name = "bar"
}

resource "aws_ecr_lifecycle_policy" "foopolicy" {
  repository = "${aws_ecr_repository.foo.name}"

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

## Argument Reference

The following arguments are supported:

* `repository` - (Required) Name of the repository to apply the policy.
* `policy` - (Required) The policy document. This is a JSON formatted string.

## Attributes Reference

The following attributes are exported:

* `repository` - The name of the repository.
* `registry_id` - The registry ID where the repository was created.
