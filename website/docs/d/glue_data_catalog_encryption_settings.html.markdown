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

* `catalog_id` - (Required) ID of the Data Catalog. This is typically the AWS account ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `data_catalog_encryption_settings` – The security configuration to set. see [Data Catalog Encryption Settings](#data_catalog_encryption_settings).
* `id` – The ID of the Data Catalog to set the security configuration for.

### data_catalog_encryption_settings

* `connection_password_encryption` - When connection password protection is enabled, the Data Catalog uses a customer-provided key to encrypt the password as part of CreateConnection or UpdateConnection and store it in the ENCRYPTED_PASSWORD field in the connection properties. You can enable catalog encryption or only password encryption. see [Connection Password Encryption](#connection_password_encryption).
* `encryption_at_rest` - Encryption-at-rest configuration for the Data Catalog. see [Encryption At Rest](#encryption_at_rest).

### connection_password_encryption

* `return_connection_password_encrypted` - When set to `true`, passwords remain encrypted in the responses of GetConnection and GetConnections. This encryption takes effect independently of the catalog encryption.
* `aws_kms_key_id` - KMS key ARN that is used to encrypt the connection password.

### encryption_at_rest

* `catalog_encryption_mode` - The encryption-at-rest mode for encrypting Data Catalog data.
* `catalog_encryption_service_role` - The ARN of the AWS IAM role used for accessing encrypted Data Catalog data.
* `sse_aws_kms_key_id` - ARN of the AWS KMS key to use for encryption at rest.
