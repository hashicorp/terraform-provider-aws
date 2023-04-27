---
subcategory: "X-Ray"
layout: "aws"
page_title: "AWS: aws_xray_encryption_config"
description: |-
    Creates and manages an AWS XRay Encryption Config.
---

# Resource: aws_xray_encryption_config

Creates and manages an AWS XRay Encryption Config.

~> **NOTE:** Removing this resource from Terraform has no effect to the encryption configuration within X-Ray.

## Example Usage

```terraform
resource "aws_xray_encryption_config" "example" {
  type = "NONE"
}
```

## Example Usage with KMS Key

```terraform
data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "example" {
  statement {
    sid    = "Enable IAM User Permissions"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }

    actions   = ["kms:*"]
    resources = ["*"]
  }
}
resource "aws_kms_key" "example" {
  description             = "Some Key"
  deletion_window_in_days = 7
  policy                  = data.aws_iam_policy_document.example.json
}

resource "aws_xray_encryption_config" "example" {
  type   = "KMS"
  key_id = aws_kms_key.example.arn
}
```

## Argument Reference

* `type` - (Required) The type of encryption. Set to `KMS` to use your own key for encryption. Set to `NONE` for default encryption.
* `key_id` - (Optional) An AWS KMS customer master key (CMK) ARN.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Region name.

## Import

XRay Encryption Config can be imported using the region name, e.g.,

```
$ terraform import aws_xray_encryption_config.example us-west-2
```
