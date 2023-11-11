---
subcategory: "Elemental MediaStore"
layout: "aws"
page_title: "AWS: aws_media_store_container_policy"
description: |-
  Provides a MediaStore Container Policy.
---

# Resource: aws_media_store_container_policy

Provides a MediaStore Container Policy.

## Example Usage

```terraform
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_media_store_container" "example" {
  name = "example"
}

data "aws_iam_policy_document" "example" {
  statement {
    sid    = "MediaStoreFullAccess"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }

    actions   = ["mediastore:*"]
    resources = ["arn:aws:mediastore:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:container/${aws_media_store_container.example.name}/*"]

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["true"]
    }
  }
}

resource "aws_media_store_container_policy" "example" {
  container_name = aws_media_store_container.example.name
  policy         = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `container_name` - (Required) The name of the container.
* `policy` - (Required) The contents of the policy. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MediaStore Container Policy using the MediaStore Container Name. For example:

```terraform
import {
  to = aws_media_store_container_policy.example
  id = "example"
}
```

Using `terraform import`, import MediaStore Container Policy using the MediaStore Container Name. For example:

```console
% terraform import aws_media_store_container_policy.example example
```
