---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_federated_catalog"
description: |-
  Manages an AWS Glue Federated Catalog.
---

# Resource: aws_glue_federated_catalog

Manages an AWS Glue Federated Catalog. Federated catalogs allow you to connect external data sources like Amazon S3 Tables to AWS Glue.

More information about AWS Glue and federated catalogs can be found in the [AWS Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/federated-catalogs.html).

## Example Usage

### S3 Tables Federated Catalog

```terraform
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

# IAM role for Lake Formation data access
resource "aws_iam_role" "example" {
  name = "glue-s3tables-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lakeformation.amazonaws.com"
        }
        Action = [
          "sts:AssumeRole",
          "sts:SetSourceIdentity",
          "sts:SetContext"
        ]
      }
    ]
  })
}

# IAM policy for S3 Tables permissions
resource "aws_iam_role_policy" "example" {
  name = "glue-s3tables-policy"
  role = aws_iam_role.example.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3tables:ListTableBuckets",
          "s3tables:CreateTableBucket",
          "s3tables:GetTableBucket",
          "s3tables:CreateNamespace",
          "s3tables:GetNamespace",
          "s3tables:ListNamespaces",
          "s3tables:DeleteNamespace",
          "s3tables:CreateTable",
          "s3tables:DeleteTable",
          "s3tables:GetTable",
          "s3tables:ListTables",
          "s3tables:GetTableData",
          "s3tables:PutTableData"
        ]
        Resource = [
          "arn:${data.aws_partition.current.partition}:s3tables:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:bucket/*"
        ]
      }
    ]
  })
}

# Register the S3 Tables location with Lake Formation
resource "aws_lakeformation_resource" "example" {
  arn      = "arn:${data.aws_partition.current.partition}:s3tables:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:bucket/*"
  role_arn = aws_iam_role.example.arn
}

resource "aws_glue_federated_catalog" "example" {
  name        = "s3tablescatalog"
  description = "S3 Tables federated catalog for analytics"

  federated_catalog {
    identifier      = "arn:${data.aws_partition.current.partition}:s3tables:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:bucket/*"
    connection_name = "aws:s3tables"
  }

  depends_on = [aws_lakeformation_resource.example]
}
```

### Redshift Data Lake Catalog

```terraform
resource "aws_iam_role" "redshift_example" {
  name = "glue-redshift-catalog-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lakeformation.amazonaws.com"
        }
        Action = [
          "sts:AssumeRole",
          "sts:SetSourceIdentity",
          "sts:SetContext"
        ]
      }
    ]
  })
}

resource "aws_glue_federated_catalog" "redshift_example" {
  name        = "redshift-catalog"
  description = "Redshift federated catalog for data lake access"

  catalog_properties {
    data_lake_access_properties {
      catalog_type       = "aws:redshift"
      data_lake_access   = true
      data_transfer_role = aws_iam_role.redshift_example.arn
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the federated catalog.

**Note:** At least one of `federated_catalog` or `catalog_properties` must be specified.

The following arguments are optional:

* `catalog_id` - (Optional) ID of the catalog. If omitted, this defaults to the AWS Account ID.
* `description` - (Optional) Description of the federated catalog.
* `federated_catalog` - (Optional) Configuration block for federated catalog parameters. See [federated_catalog](#federated_catalog) below.
* `catalog_properties` - (Optional) Configuration block for catalog properties. See [catalog_properties](#catalog_properties) below.

### federated_catalog

* `identifier` - (Optional) Unique identifier for the federated catalog.
* `connection_name` - (Optional) Name of the connection for the federated catalog.

### catalog_properties

* `data_lake_access_properties` - (Optional) Configuration block for data lake access properties. See [data_lake_access_properties](#data_lake_access_properties) below.

### data_lake_access_properties

* `catalog_type` - (Optional) Type of catalog. Currently only `aws:redshift` is supported.
* `data_lake_access` - (Optional) Whether to enable data lake access for the catalog.
* `data_transfer_role` - (Optional) ARN of the IAM role for data transfer operations.
* `kms_key` - (Optional) KMS key for encryption.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Federated Catalog.
* `id` - Federated catalog identifier.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Federated Catalog using the `catalog_id:name`. For example:

```terraform
import {
  to = aws_glue_federated_catalog.example
  id = "123456789012:s3tablescatalog"
}
```

Using `terraform import`, import Glue Federated Catalog using the `catalog_id:name`. For example:

```console
% terraform import aws_glue_federated_catalog.example 123456789012:s3tablescatalog
```
