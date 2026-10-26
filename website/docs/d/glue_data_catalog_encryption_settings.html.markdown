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

This data source supports the following arguments:

* `catalog_id` - (Required) ID of the Data Catalog. This is typically the AWS account ID.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `data_catalog_encryption_settings` - Security configuration to set. see [`data_catalog_encryption_settings` Block](#data_catalog_encryption_settings-block).
* `id` - ID of the Data Catalog to set the security configuration for.

### `data_catalog_encryption_settings` Block

* `connection_password_encryption` - When connection password protection is enabled, the Data Catalog uses a customer-provided key to encrypt the password as part of CreateConnection or UpdateConnection and store it in the ENCRYPTED_PASSWORD field in the connection properties. You can enable catalog encryption or only password encryption. see [`connection_password_encryption` Block](#connection_password_encryption-block).
* `encryption_at_rest` - Encryption-at-rest configuration for the Data Catalog. see [`encryption_at_rest` Block](#encryption_at_rest-block).

### `connection_password_encryption` Block

* `aws_kms_key_id` - KMS key ARN that is used to encrypt the connection password.
* `return_connection_password_encrypted` - When set to `true`, passwords remain encrypted in the responses of GetConnection and GetConnections. This encryption takes effect independently of the catalog encryption.

### `encryption_at_rest` Block

* `catalog_encryption_mode` - Encryption-at-rest mode for encrypting Data Catalog data.
* `catalog_encryption_service_role` - ARN of the AWS IAM role used for accessing encrypted Data Catalog data.
* `sse_aws_kms_key_id` - ARN of the AWS KMS key to use for encryption at rest.
