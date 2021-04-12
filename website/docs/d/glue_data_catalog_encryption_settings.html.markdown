---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_data_catalog_encryption_settings"
description: |-
  Get information on AWS Glue Data Catalog Encryption Settings
---

# Data Source: aws_glue_data_catalog_encryption_settings

This data source can be used to fetch information about AWS Glue Data Catalog Encryption Settings.

## Example Usage

```terraform
data "aws_glue_data_catalog_encryption_settings" "example" {
  id = "123456789123"
}
```

## Argument Reference

* `id` - (Required) The ID of the Data Catalog. This is typically the AWS account ID.

## Attributes Reference

* `connection_password_encrypted` - A boolean value which specifies whether connection passwords are encrypted.

* `connection_password_kms_key_arn` - The ARN of the KMS key that encrypts connection passwords.

* `encryption_mode` - The encryption mode of the Glue Data Catalog.

* `encryption_kms_key_arn` - The ARN of the KMS key that encrypts the Glue Data Catalog.
