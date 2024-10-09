---
subcategory: "CodeArtifact"
layout: "aws"
page_title: "AWS: aws_codeartifact_repository_permissions_policy"
description: |-
  Provides a CodeArtifact Repository Permissions Policy resource.
---

# Resource: aws_codeartifact_repository_permissions_policy

Provides a CodeArtifact Repostory Permissions Policy Resource.

## Example Usage

```terraform
resource "aws_kms_key" "example" {
  description = "domain key"
}

resource "aws_codeartifact_domain" "example" {
  domain         = "example"
  encryption_key = aws_kms_key.example.arn
}

resource "aws_codeartifact_repository" "example" {
  repository = "example"
  domain     = aws_codeartifact_domain.example.domain
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions   = ["codeartifact:ReadFromRepository"]
    resources = [aws_codeartifact_repository.example.arn]
  }
}
resource "aws_codeartifact_repository_permissions_policy" "example" {
  repository      = aws_codeartifact_repository.example.repository
  domain          = aws_codeartifact_domain.example.domain
  policy_document = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `repository` - (Required) The name of the repository to set the resource policy on.
* `domain` - (Required) The name of the domain on which to set the resource policy.
* `policy_document` - (Required) A JSON policy string to be set as the access control resource policy on the provided domain.
* `domain_owner` - (Optional) The account number of the AWS account that owns the domain.
* `policy_revision` - (Optional) The current revision of the resource policy to be set. This revision is used for optimistic locking, which prevents others from overwriting your changes to the domain's resource policy.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of the resource associated with the resource policy.
* `resource_arn` - The ARN of the resource associated with the resource policy.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeArtifact Repository Permissions Policies using the CodeArtifact Repository ARN. For example:

```terraform
import {
  to = aws_codeartifact_repository_permissions_policy.example
  id = "arn:aws:codeartifact:us-west-2:012345678912:repository/tf-acc-test-6968272603913957763/tf-acc-test-6968272603913957763"
}
```

Using `terraform import`, import CodeArtifact Repository Permissions Policies using the CodeArtifact Repository ARN. For example:

```console
% terraform import aws_codeartifact_repository_permissions_policy.example arn:aws:codeartifact:us-west-2:012345678912:repository/tf-acc-test-6968272603913957763/tf-acc-test-6968272603913957763
```
