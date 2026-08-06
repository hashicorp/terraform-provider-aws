---
subcategory: "Clean Rooms"
layout: "aws"
page_title: "AWS: aws_cleanrooms_configured_table_analysis_rule"
description: |-
  Provides a Clean Rooms Configured Table Analysis Rule.
---

# Resource: aws_cleanrooms_configured_table_analysis_rule

Creates a new analysis rule for a configured table. Currently, only one analysis rule can be created for a given configured table.

## Example Usage

### Configured Table Analysis Rule type LIST

```terraform
resource "aws_cleanrooms_configured_table_analysis_rule" "example" {
  name                        = "example"
  analysis_rule_type          = "LIST"
  configured_table_identifier = "example_configured_table_id"

  analysis_rule_policy {
    v1 {
      list {
        join_columns = [ "my_column_1" ]
        allowed_join_operators = [ "AND" ]
        list_columns = [ "my_column_3" ]
      }
    }
  }
}
```

### Configured Table Analysis Rule type AGGREGATION

```terraform
resource "aws_cleanrooms_configured_table_analysis_rule" "example" {
  name                        = "example"
  analysis_rule_type          = "AGGREGATION"
  configured_table_identifier = "example_configured_table_id"

  analysis_rule_policy {
    v1 {
      aggregation {
        aggregate_columns {
          column_names = [ "my_column_1" ]
          function     = "SUM"
        }
        join_aggregate_columns = [ "my_column_2" ]
        join_required          = "QUERY_RUNNER"
        allowed_join_operators = [ "OR" ]
        dimension_columns      = [ "my_column_3" ]
        scalar_functions       = [ "TRUNC" ]
        output_constraints {
          column_name = "my_column_4"
          minimum     = "2"
          type        = "COUNT_DISTINCT"
        }
      }
    }
  }
}
```

### Configured Table Analysis Rule type CUSTOM

```terraform
resource "aws_cleanrooms_configured_table_analysis_rule" "example" {
  name                        = "example"
  analysis_rule_type          = "CUSTOM"
  configured_table_identifier = "example_configured_table_id"

  analysis_rule_policy {
    v1 {
      custom {
        allowed_custom_analyses    = [ "ANY_QUERY" ]
        allowed_analyses_providers = [ "123456789012" ]
        differential_privacy {
          columns {
            name = "my_column_1"
          }
        }
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) - The name of the analysis rule.
* `analysis_rule_type` - (Required - Forces new resource) - The type of analysis rule. Possible values are LIST, AGGREGATION and CUSTOM.
* `configured_table_identifier` - (Required - Forces new resource) - The identifier for the configured table to create the analysis rule for. Currently accepts the configured table ID.
* `analysis_rule_policy` - (Required) - The entire created configured table analysis rule object.
* `v1` - (Required) - Controls on the query specifications that can be run on a configured table.
* `list` - (Optional - Forces new resource) - Analysis rule type that enables only list queries on a configured table.
    - `join_columns` - (Optional - Required if analysis_rule_type is LIST) - Columns that can be used to join a configured table with the table of the member who can query and other members’ configured tables.
    - `allowed_join_operators` - (Optional - Required if analysis_rule_type is LIST) - The logical operators (if any) that are to be used in an INNER JOIN match condition. Default is AND.
    - `list_columns` - (Optional - Required if analysis_rule_type is LIST) - Columns that can be listed in the output.
* `aggregation` - (Optional - Forces new resource) - Analysis rule type that enables only aggregation queries on a configured table.
    - `aggregate_columns` - (Optional - Required if analysis_rule_type is AGGREGATION) - The columns that query runners are allowed to use in aggregation queries. Multiple aggregate columns are allowed.
        - `column_names` - (Optional) - Column names in configured table of aggregate columns.
        - `function` - (Optional) - Aggregation function that can be applied to aggregate column in query.
    - `join_columns` - (Optional - Required if analysis_rule_type is AGGREGATION) - Columns in configured table that can be used in join statements and/or as aggregate columns. They can never be outputted directly.
    - `join_required` - (Optional - Required if analysis_rule_type is AGGREGATION) - Control that requires member who runs query to do a join with their configured table and/or other configured table in query.
    - `allowed_join_operators` - (Optional - Required if analysis_rule_type is AGGREGATION) - Which logical operators (if any) are to be used in an INNER JOIN match condition. Default is AND.
    - `dimension_columns` - (Optional - Required if analysis_rule_type is AGGREGATION) - The columns that query runners are allowed to select, group by, or filter by.
    - `scalar_functions` - (Optional - Required if analysis_rule_type is AGGREGATION) - Set of scalar functions that are allowed to be used on dimension columns and the output of aggregation of metrics.
    - `output_constraints` - (Optional - Required if analysis_rule_type is AGGREGATION) - Columns that must meet a specific threshold value (after an aggregation function is applied to it) for each output row to be returned. Multiple output constraints are allowed.
        - `column_name` - (Optional) - Column in aggregation constraint for which there must be a minimum number of distinct values in an output row for it to be in the query output.
        - `minimum` - (Optional) - The minimum number of distinct values that an output row must be an aggregation of. Minimum threshold of distinct values for a specified column that must exist in an output row for it to be in the query output.
        - `type` - (Optional) - The type of aggregation the constraint allows. The only valid value is currently COUNT_DISTINCT.
* `custom` - (Optional - Forces new resource) - A type of analysis rule that enables the table owner to approve custom SQL queries on their configured tables. It supports differential privacy.
    - `allowed_custom_analyses` - (Optional - Required if analysis_rule_type is CUSTOM) - The ARN of the analysis templates that are allowed by the custom analysis rule.
    - `allowed_analyses_providers` - (Optional) - The IDs of the Amazon Web Services accounts that are allowed to query by the custom analysis rule. Required when allowed_custom_analyses is ANY_QUERY.
    - `differential_privacy` - (Optional) - The differential privacy configuration.
        - `columns` - (Optional) -  The name of the column (such as user_id) that contains the unique identifier of your users whose privacy you want to protect. If you want to turn on diﬀerential privacy for two or more tables in a collaboration, you must conﬁgure the same column as the user identiﬁer column in both analysis rules.
            - `name` - (Optional) - The name of the column, such as user_id, that contains the unique identifier of your users, whose privacy you want to protect. If you want to turn on differential privacy for two or more tables in a collaboration, you must configure the same column as the user identifier column in both analysis rules.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `2m`)
- `update` - (Default `2m`)
- `delete` - (Default `2m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_cleanrooms_configured_table_analysis_rule` using the `id`. For example:

```terraform
import {
  to = aws_cleanrooms_configured_table_analysis_rule.example
  id = "1234abcd-12ab-34cd-56ef-1234567890ab"
}
```

Using `terraform import`, import `aws_cleanrooms_configured_table_analysis_rule` using the `id`. For example:

```console
% terraform import aws_cleanrooms_configured_table_analysis_rule.example 1234abcd-12ab-34cd-56ef-1234567890ab
```