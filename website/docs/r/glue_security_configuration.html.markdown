---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_security_configuration"
description: |-
  Manages a Glue Security Configuration
---

# Resource: aws_glue_security_configuration

Manages a Glue Security Configuration.

## Example Usage

```terraform
resource "aws_glue_security_configuration" "example" {
  name = "example"

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      kms_key_arn        = data.aws_kms_key.example.arn
      s3_encryption_mode = "SSE-KMS"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `encryption_configuration` – (Required) Configuration block containing encryption configuration. Detailed below.
* `name` – (Required) Name of the security configuration.

### encryption_configuration Argument Reference

* `cloudwatch_encryption ` - (Required) A `cloudwatch_encryption ` block as described below, which contains encryption configuration for CloudWatch.
* `job_bookmarks_encryption ` - (Required) A `job_bookmarks_encryption ` block as described below, which contains encryption configuration for job bookmarks.
* `s3_encryption` - (Required) A `s3_encryption ` block as described below, which contains encryption configuration for S3 data.

#### cloudwatch_encryption Argument Reference

* `cloudwatch_encryption_mode` - (Optional) Encryption mode to use for CloudWatch data. Valid values: `DISABLED`, `SSE-KMS`. Default value: `DISABLED`.
* `kms_key_arn` - (Optional) Amazon Resource Name (ARN) of the KMS key to be used to encrypt the data.

#### job_bookmarks_encryption Argument Reference

* `job_bookmarks_encryption_mode` - (Optional) Encryption mode to use for job bookmarks data. Valid values: `CSE-KMS`, `DISABLED`. Default value: `DISABLED`.
* `kms_key_arn` - (Optional) Amazon Resource Name (ARN) of the KMS key to be used to encrypt the data.

#### s3_encryption Argument Reference

* `s3_encryption_mode` - (Optional) Encryption mode to use for S3 data. Valid values: `DISABLED`, `SSE-KMS`, `SSE-S3`. Default value: `DISABLED`.
* `kms_key_arn` - (Optional) Amazon Resource Name (ARN) of the KMS key to be used to encrypt the data.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Glue security configuration name

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Security Configurations using `name`. For example:

```terraform
import {
  to = aws_glue_security_configuration.example
  id = "example"
}
```

Using `terraform import`, import Glue Security Configurations using `name`. For example:

```console
% terraform import aws_glue_security_configuration.example example
```
