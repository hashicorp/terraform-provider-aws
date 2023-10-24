---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_data_catalog"
description: |-
  Provides an Athena data catalog.
---

# Resource: aws_athena_data_catalog

Provides an Athena data catalog.

More information about Athena and Athena data catalogs can be found in the [Athena User Guide](https://docs.aws.amazon.com/athena/latest/ug/what-is.html).

-> **Tip:** for a more detailed explanation on the usage of `parameters`, see the [DataCatalog API documentation](https://docs.aws.amazon.com/athena/latest/APIReference/API_DataCatalog.html)

## Example Usage

```terraform
resource "aws_athena_data_catalog" "example" {
  name        = "athena-data-catalog"
  description = "Example Athena data catalog"
  type        = "LAMBDA"

  parameters = {
    "function" = "arn:aws:lambda:eu-central-1:123456789012:function:not-important-lambda-function"
  }

  tags = {
    Name = "example-athena-data-catalog"
  }
}
```

### Hive based Data Catalog

```terraform
resource "aws_athena_data_catalog" "example" {
  name        = "hive-data-catalog"
  description = "Hive based Data Catalog"
  type        = "HIVE"

  parameters = {
    "metadata-function" = "arn:aws:lambda:eu-central-1:123456789012:function:not-important-lambda-function"
  }
}
```

### Glue based Data Catalog

```terraform
resource "aws_athena_data_catalog" "example" {
  name        = "glue-data-catalog"
  description = "Glue based Data Catalog"
  type        = "GLUE"

  parameters = {
    "catalog-id" = "123456789012"
  }
}
```

### Lambda based Data Catalog

```terraform
resource "aws_athena_data_catalog" "example" {
  name        = "lambda-data-catalog"
  description = "Lambda based Data Catalog"
  type        = "LAMBDA"

  parameters = {
    "metadata-function" = "arn:aws:lambda:eu-central-1:123456789012:function:not-important-lambda-function-1"
    "record-function"   = "arn:aws:lambda:eu-central-1:123456789012:function:not-important-lambda-function-2"
  }
}
```

## Argument Reference

This resource supports the following arguments:

- `name` - (Required) Name of the data catalog. The catalog name must be unique for the AWS account and can use a maximum of 128 alphanumeric, underscore, at sign, or hyphen characters.
- `type` - (Required) Type of data catalog: `LAMBDA` for a federated catalog, `GLUE` for AWS Glue Catalog, or `HIVE` for an external hive metastore.
- `parameters` - (Required) Key value pairs that specifies the Lambda function or functions to use for the data catalog. The mapping used depends on the catalog type.
- `description` - (Required) Description of the data catalog.
- `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `id` - Name of the data catalog.
- `arn` - ARN of the data catalog.
- `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import data catalogs using their `name`. For example:

```terraform
import {
  to = aws_athena_data_catalog.example
  id = "example-data-catalog"
}
```

Using `terraform import`, import data catalogs using their `name`. For example:

```console
% terraform import aws_athena_data_catalog.example example-data-catalog
```
