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

### Example Usage with KMS Key

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

This resource supports the following arguments:

* `key_id` - (Optional) AWS KMS customer master key (CMK) ARN.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `type` - (Required) Type of encryption. Set to `KMS` to use your own key for encryption. Set to `NONE` for default encryption.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Region name.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_xray_encryption_config.example
  identity = {
    region = "us-west-2"
  }
}

resource "aws_xray_encryption_config" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Optional

* `account_id` (String) Account ID where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import XRay Encryption Config using the region name. For example:

```terraform
import {
  to = aws_xray_encryption_config.example
  id = "us-west-2"
}
```

Using `terraform import`, import XRay Encryption Config using the region name. For example:

```console
% terraform import aws_xray_encryption_config.example us-west-2
```
