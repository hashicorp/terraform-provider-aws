---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_integration_table_properties"
description: |-
  Manages an AWS Glue Integration Table Properties.
---

# Resource: aws_glue_integration_table_properties

Manages an AWS Glue Integration Table Properties configuration for Zero-ETL integrations.

## Example Usage

### Basic Usage

```terraform
resource "aws_glue_catalog_database" "example" {
  name = "example"
}

resource "aws_glue_integration_table_properties" "example" {
  resource_arn = aws_glue_catalog_database.example.arn
  table_name   = "example_table"

  target_table_config {
    unnest_spec       = "FULL"
    target_table_name = "example_target_table"

    partition_spec {
      field_name      = "created_at"
      function_spec   = "month"
      conversion_spec = "iso"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required) ARN of the Glue resource (database or table) for the integration table properties.
* `table_name` - (Required) Name of the table for the integration table properties.
* `source_table_config` - (Optional) Configuration block for source table properties. See [source_table_config](#source_table_config) below.
* `target_table_config` - (Optional) Configuration block for target table properties. See [target_table_config](#target_table_config) below.

### source_table_config

* `fields` - (Optional) List of fields to include from the source table.
* `filter_predicate` - (Optional) Filter predicate to apply to the source table data.
* `primary_key` - (Optional) List of columns that constitute the primary key.
* `record_update_field` - (Optional) Field that indicates when a record was last updated.

### target_table_config

* `target_table_name` - (Optional) Name of the target table in the integration.
* `unnest_spec` - (Optional) Specification for unnesting nested data structures.
* `partition_spec` - (Optional) Configuration block for partitioning specifications. See [partition_spec](#partition_spec) below.

### partition_spec

* `conversion_spec` - (Optional) Conversion specification for the partition field.
* `field_name` - (Optional) Name of the field used for partitioning.
* `function_spec` - (Optional) Function specification for partition processing.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Composite ID of the integration table properties in the format `resource_arn,table_name`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_glue_integration_table_properties.example
  identity = {
    resource_arn = "arn:aws:glue:us-east-1:123456789012:database/example"
    table_name   = "example_table"
  }
}

resource "aws_glue_integration_table_properties" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `resource_arn` - ARN of the Glue resource (database or table) for the integration table properties.
* `table_name` - Name of the table for the integration table properties.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Integration Table Properties using the `resource_arn` and `table_name` separated by a comma. For example:

```terraform
import {
  to = aws_glue_integration_table_properties.example
  id = "arn:aws:glue:us-east-1:123456789012:database/example,example_table"
}
```

Using `terraform import`, import Glue Integration Table Properties using the `resource_arn` and `table_name` separated by a comma. For example:

```console
% terraform import aws_glue_integration_table_properties.example "arn:aws:glue:us-east-1:123456789012:database/example,example_table"
```

  to = aws_glue_integration_table_properties.example
  id = "integration_table_properties-id-12345678"
}

```

Using `terraform import`, import Glue Integration Table Properties using the `example_id_arg`. For example:

```console
% terraform import aws_glue_integration_table_properties.example integration_table_properties-id-12345678
```
