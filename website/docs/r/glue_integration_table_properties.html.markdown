---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_integration_table_properties"
description: |-
  Manages AWS Glue Integration Table Properties for zero‑ETL, configuring per-table source/target options.
---

# Resource: aws_glue_integration_table_properties

Manages per-table override properties for Glue zero‑ETL integrations. These include filtering/partitioning (source/target) and target table options.

Reference: AWS API [CreateIntegrationTableProperties](https://docs.aws.amazon.com/glue/latest/webapi/API_CreateIntegrationTableProperties.html)

Note: As of the AWS documentation, `resource_arn` must be the Glue Catalog target table ARN.

## Example Usage

```terraform
resource "aws_glue_integration_table_properties" "example" {
  resource_arn = "arn:aws:glue:us-east-1:111122223333:table/example-db/example-table"
  table_name   = "example-table"

  source_table_config {
    fields              = ["col_a", "col_b"]
    filter_predicate    = "col_a > 100"
    primary_key         = ["pk"]
    record_update_field = "last_updated"
  }

  target_table_config {
    target_table_name = "example_table_flattened"
    unnest_spec       = "TOPLEVEL"

    partition_spec {
      field_name      = "event_time"
      function_spec   = "hour"
      conversion_spec = "epoch_milli"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

- `resource_arn` (Required) Glue Data Catalog target table ARN for which to create properties.
- `table_name` (Required) Name of the table to be replicated.
- `source_table_config` (Optional) Source table configuration block:
    - `fields` (Optional) List of fields for column-level filtering.
    - `filter_predicate` (Optional) Row-level filter expression.
    - `primary_key` (Optional) List of primary key fields (limited support per AWS docs).
    - `record_update_field` (Optional) Timestamp field used for incremental pulls.
- `target_table_config` (Optional) Target table configuration block:
    - `partition_spec` (Optional) List of partition specifications:
        - `field_name` (Required) Field used for partitioning.
        - `function_spec` (Optional) Partition function: `identity`, `year`, `month`, `day`, or `hour`.
        - `conversion_spec` (Optional) Timestamp format when using time-based functions: `epoch_sec`, `epoch_milli`, or `iso`.
    - `target_table_name` (Optional) Name of the target table override.
    - `unnest_spec` (Optional) How to flatten nested objects: `TOPLEVEL`, `FULL`, or `NOUNNEST`.

* `region` - (Optional) Region where this resource will be managed (https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the provider configuration (https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Import

Import Integration Table Properties using a composite ID in the form `resource_arn,table_name`. For example:

```terraform
import {
  to = aws_glue_integration_table_properties.example
  id = "arn:aws:glue:us-east-1:111122223333:table/example-db/example-table,example-table"
}
```

```console
% terraform import aws_glue_integration_table_properties.example arn:aws:glue:us-east-1:111122223333:table/example-db/example-table,example-table
```

## Attribute Reference

This resource exports no additional attributes.
