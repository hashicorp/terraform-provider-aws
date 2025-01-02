---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_data_catalog_encryption_settings"
description: |-
  Provides a Glue Data Catalog Encryption Settings resource.
---

# Resource: aws_glue_data_catalog_encryption_settings

Provides a Glue Data Catalog Encryption Settings resource.

## Example Usage

```terraform
resource "aws_glue_data_catalog_encryption_settings" "example" {
  data_catalog_encryption_settings {
    connection_password_encryption {
      aws_kms_key_id                       = aws_kms_key.test.arn
      return_connection_password_encrypted = true
    }

    encryption_at_rest {
      catalog_encryption_mode         = "SSE-KMS"
      catalog_encryption_service_role = aws_iam.role.test.arn
      sse_aws_kms_key_id              = aws_kms_key.test.arn
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `data_catalog_encryption_settings` – (Required) The security configuration to set. see [Data Catalog Encryption Settings](#data_catalog_encryption_settings).
* `catalog_id` – (Optional) The ID of the Data Catalog to set the security configuration for. If none is provided, the AWS account ID is used by default.

### data_catalog_encryption_settings

* `connection_password_encryption` - (Required) When connection password protection is enabled, the Data Catalog uses a customer-provided key to encrypt the password as part of CreateConnection or UpdateConnection and store it in the ENCRYPTED_PASSWORD field in the connection properties. You can enable catalog encryption or only password encryption. see [Connection Password Encryption](#connection_password_encryption).
* `encryption_at_rest` - (Required) Specifies the encryption-at-rest configuration for the Data Catalog. see [Encryption At Rest](#encryption_at_rest).

### connection_password_encryption

* `return_connection_password_encrypted` - (Required) When set to `true`, passwords remain encrypted in the responses of GetConnection and GetConnections. This encryption takes effect independently of the catalog encryption.
* `aws_kms_key_id` - (Optional) A KMS key ARN that is used to encrypt the connection password. If connection password protection is enabled, the caller of CreateConnection and UpdateConnection needs at least `kms:Encrypt` permission on the specified AWS KMS key, to encrypt passwords before storing them in the Data Catalog.

### encryption_at_rest

* `catalog_encryption_mode` - (Required) The encryption-at-rest mode for encrypting Data Catalog data. Valid values: `DISABLED`, `SSE-KMS`, `SSE-KMS-WITH-SERVICE-ROLE`.
* `catalog_encryption_service_role` - (Optional) The ARN of the AWS IAM role used for accessing encrypted Data Catalog data.
* `sse_aws_kms_key_id` - (Optional) The ARN of the AWS KMS key to use for encryption at rest.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the Data Catalog to set the security configuration for.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Data Catalog Encryption Settings using `CATALOG-ID` (AWS account ID if not custom). For example:

```terraform
import {
  to = aws_glue_data_catalog_encryption_settings.example
  id = "123456789012"
}
```

Using `terraform import`, import Glue Data Catalog Encryption Settings using `CATALOG-ID` (AWS account ID if not custom). For example:

```console
% terraform import aws_glue_data_catalog_encryption_settings.example 123456789012
```
