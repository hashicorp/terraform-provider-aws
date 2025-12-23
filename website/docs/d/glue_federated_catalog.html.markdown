---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_federated_catalog"
description: |-
  Provides details about an AWS Glue Federated Catalog.
---
# Data Source: aws_glue_federated_catalog

Provides details about an AWS Glue Federated Catalog. Federated catalogs allow you to connect external data sources like Amazon S3 Tables to AWS Glue.

## Example Usage

### Basic Usage

```terraform
data "aws_glue_federated_catalog" "example" {
  name = "s3tablescatalog"
}
```

### With Specific Catalog ID

```terraform
data "aws_glue_federated_catalog" "example" {
  name       = "my-federated-catalog"
  catalog_id = "123456789012"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the federated catalog to retrieve.

The following arguments are optional:

* `catalog_id` - (Optional) ID of the catalog. If omitted, this defaults to the AWS Account ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Federated Catalog.
* `description` - Description of the federated catalog.
* `federated_catalog` - Configuration block for federated catalog parameters:
  * `identifier` - Unique identifier for the federated catalog.
  * `connection_name` - Name of the connection for the federated catalog.
* `catalog_properties` - Configuration block for catalog properties:
  * `data_lake_access_properties` - Configuration block for data lake access properties:
    * `catalog_type` - Type of catalog (e.g., "aws:redshift").
    * `data_lake_access` - Whether data lake access is enabled for the catalog.
    * `data_transfer_role` - ARN of the IAM role for data transfer operations.
    * `kms_key` - KMS key for encryption.
