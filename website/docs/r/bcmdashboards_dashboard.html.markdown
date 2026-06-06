---
subcategory: "BCM Dashboards"
layout: "aws"
page_title: "AWS: aws_bcmdashboards_dashboard"
description: |-
  Manages an AWS Billing and Cost Management Dashboards Dashboard.
---

# Resource: aws_bcmdashboards_dashboard

Manages an AWS Billing and Cost Management (BCM) Dashboards Dashboard.

## Example Usage

### Basic Usage

```terraform
resource "aws_bcmdashboards_dashboard" "example" {
  name        = "example-dashboard"
  description = "Managed by Terraform"

  widget {
    title  = "Monthly unblended cost"
    height = 4
    width  = 4

    configs {
      query_parameters {
        cost_and_usage {
          granularity = "MONTHLY"
          metrics     = ["UnblendedCost"]

          time_range {
            start_time {
              type  = "ABSOLUTE"
              value = "2025-01-01"
            }
            end_time {
              type  = "ABSOLUTE"
              value = "2025-07-31"
            }
          }

          group_by {
            key  = "SERVICE"
            type = "DIMENSION"
          }
        }
      }

      display_config {
        graph {
          metric      = "UnblendedCost"
          visual_type = "BAR"
        }
      }
    }
  }
}
```

### Filtering Cost and Usage Data

```terraform
resource "aws_bcmdashboards_dashboard" "example" {
  name = "filtered-dashboard"

  widget {
    title = "Production data transfer"

    configs {
      query_parameters {
        cost_and_usage {
          granularity = "MONTHLY"
          metrics     = ["UnblendedCost"]

          time_range {
            start_time {
              type  = "RELATIVE"
              value = "-P3M"
            }
            end_time {
              type  = "ABSOLUTE"
              value = "2025-07-31"
            }
          }

          filter {
            and {
              tags {
                key    = "Environment"
                values = ["production"]
              }
            }
            and {
              dimensions {
                key    = "USAGE_TYPE"
                values = ["DataTransfer-In-Bytes"]
              }
            }
          }
        }
      }

      display_config {
        table {}
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the dashboard. Must be unique within your account.

The following arguments are optional:

* `description` - (Optional) Description of the dashboard.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `widget` - (Optional) Configuration blocks for the dashboard widgets. Up to 20 widgets are supported. See [`widget`](#widget-block) below.

### `widget` Block

* `configs` - (Required) One or two configuration blocks defining the data query and display settings for the widget. See [`configs`](#configs-block) below.
* `description` - (Optional) Description of the widget.
* `height` - (Optional) Height of the widget in dashboard grid row spans.
* `horizontal_offset` - (Optional) Starting column position of the widget in the dashboard grid layout.
* `title` - (Required) Title of the widget.
* `width` - (Optional) Width of the widget in dashboard grid column spans.

### `configs` Block

* `display_config` - (Required) Configuration block defining how the retrieved data is visualized. See [`display_config`](#display_config-block) below.
* `query_parameters` - (Required) Configuration block defining what data the widget retrieves. Exactly one query type must be specified. See [`query_parameters`](#query_parameters-block) below.

### `query_parameters` Block

* `cost_and_usage` - (Optional) Cost and usage query. Exactly one query type must be specified. See [`cost_and_usage`](#cost_and_usage-block) below.
* `reservation_coverage` - (Optional) Reserved Instance coverage query. See [`reservation_coverage`](#reservation_coverage-block) below.
* `reservation_utilization` - (Optional) Reserved Instance utilization query. See [`reservation_utilization`](#reservation_utilization-block) below.
* `savings_plans_coverage` - (Optional) Savings Plans coverage query. See [`savings_plans_coverage`](#savings_plans_coverage-block) below.
* `savings_plans_utilization` - (Optional) Savings Plans utilization query. See [`savings_plans_utilization`](#savings_plans_utilization-block) below.

### `cost_and_usage` Block

* `granularity` - (Required) Granularity of the retrieved data. Valid values: `HOURLY`, `DAILY`, `MONTHLY`.
* `metrics` - (Required) Cost and usage metrics to retrieve, such as `UnblendedCost` or `UsageQuantity`.
* `filter` - (Optional) Configuration block defining the filter expression applied to the data. See [`filter`](#filter-block) below.
* `group_by` - (Optional) Configuration blocks specifying how to group the data. See [`group_by`](#group_by-block) below.
* `time_range` - (Required) Configuration block specifying the time period for the data. See [`time_range`](#time_range-block) below.

### `reservation_coverage` Block

* `filter` - (Optional) Configuration block defining the filter expression applied to the data. See [`filter`](#filter-block) below.
* `granularity` - (Optional) Granularity of the retrieved data. Valid values: `HOURLY`, `DAILY`, `MONTHLY`.
* `group_by` - (Optional) Configuration blocks specifying how to group the data. See [`group_by`](#group_by-block) below.
* `metrics` - (Optional) Coverage metrics to retrieve, such as `Hour`, `Unit`, or `Cost`.
* `time_range` - (Required) Configuration block specifying the time period for the data. See [`time_range`](#time_range-block) below.

### `reservation_utilization` Block

* `filter` - (Optional) Configuration block defining the filter expression applied to the data. See [`filter`](#filter-block) below.
* `granularity` - (Optional) Granularity of the retrieved data. Valid values: `HOURLY`, `DAILY`, `MONTHLY`.
* `group_by` - (Optional) Configuration blocks specifying how to group the data. See [`group_by`](#group_by-block) below.
* `time_range` - (Required) Configuration block specifying the time period for the data. See [`time_range`](#time_range-block) below.

### `savings_plans_coverage` Block

* `filter` - (Optional) Configuration block defining the filter expression applied to the data. See [`filter`](#filter-block) below.
* `granularity` - (Optional) Granularity of the retrieved data. Valid values: `HOURLY`, `DAILY`, `MONTHLY`.
* `group_by` - (Optional) Configuration blocks specifying how to group the data. See [`group_by`](#group_by-block) below.
* `metrics` - (Optional) Coverage metrics to retrieve, such as `SpendCoveredBySavingsPlans`.
* `time_range` - (Required) Configuration block specifying the time period for the data. See [`time_range`](#time_range-block) below.

### `savings_plans_utilization` Block

* `filter` - (Optional) Configuration block defining the filter expression applied to the data. See [`filter`](#filter-block) below.
* `granularity` - (Optional) Granularity of the retrieved data. Valid values: `HOURLY`, `DAILY`, `MONTHLY`.
* `time_range` - (Required) Configuration block specifying the time period for the data. See [`time_range`](#time_range-block) below.

### `time_range` Block

* `end_time` - (Required) Configuration block specifying the end of the time range. See [`end_time`](#end_time-block) below.
* `start_time` - (Required) Configuration block specifying the start of the time range. See [`start_time`](#start_time-block) below.

### `start_time` Block

* `type` - (Required) Type of the date/time value. Valid values: `ABSOLUTE` (a specific date, such as `2025-07-01`) or `RELATIVE` (an ISO 8601 duration, such as `-P3M`).
* `value` - (Required) Date/time value.

### `end_time` Block

* `type` - (Required) Type of the date/time value. Valid values: `ABSOLUTE` (a specific date, such as `2025-07-01`) or `RELATIVE` (an ISO 8601 duration, such as `-P3M`).
* `value` - (Required) Date/time value.

### `group_by` Block

* `key` - (Required) Key to group by, such as `SERVICE` or `REGION`.
* `type` - (Optional) Type of grouping. Valid values: `DIMENSION`, `TAG`, `COST_CATEGORY`.

### `filter` Block

The filter is a cost expression. The top-level block supports the logical operators `and`, `or`, and `not` (whose operands are leaf filters) in addition to leaf filters (`dimensions`, `tags`, `cost_categories`) applied directly.

* `and` - (Optional) Leaf filter expressions combined with `AND` logic. See [`and`](#and-block) below.
* `cost_categories` - (Optional) Configuration block to filter on a cost category. See [`cost_categories`](#cost_categories-block) below.
* `dimensions` - (Optional) Configuration block to filter on a dimension. See [`dimensions`](#dimensions-block) below.
* `not` - (Optional) Single leaf filter expression negated with `NOT` logic. See [`not`](#not-block) below.
* `or` - (Optional) Leaf filter expressions combined with `OR` logic. See [`or`](#or-block) below.
* `tags` - (Optional) Configuration block to filter on a cost allocation tag. See [`tags`](#tags-block) below.

### `and` Block

* `cost_categories` - (Optional) Configuration block to filter on a cost category. See [`cost_categories`](#cost_categories-block) below.
* `dimensions` - (Optional) Configuration block to filter on a dimension. See [`dimensions`](#dimensions-block) below.
* `tags` - (Optional) Configuration block to filter on a cost allocation tag. See [`tags`](#tags-block) below.

### `or` Block

* `cost_categories` - (Optional) Configuration block to filter on a cost category. See [`cost_categories`](#cost_categories-block) below.
* `dimensions` - (Optional) Configuration block to filter on a dimension. See [`dimensions`](#dimensions-block) below.
* `tags` - (Optional) Configuration block to filter on a cost allocation tag. See [`tags`](#tags-block) below.

### `not` Block

* `cost_categories` - (Optional) Configuration block to filter on a cost category. See [`cost_categories`](#cost_categories-block) below.
* `dimensions` - (Optional) Configuration block to filter on a dimension. See [`dimensions`](#dimensions-block) below.
* `tags` - (Optional) Configuration block to filter on a cost allocation tag. See [`tags`](#tags-block) below.

### `dimensions` Block

* `key` - (Required) Dimension to filter on, such as `SERVICE`, `USAGE_TYPE`, or `REGION`.
* `match_options` - (Optional) Match options, such as `EQUALS` or `CONTAINS`.
* `values` - (Required) Values to match for the dimension.

### `tags` Block

* `key` - (Optional) Tag key to filter on.
* `match_options` - (Optional) Match options, such as `EQUALS` or `CONTAINS`.
* `values` - (Optional) Tag values to match.

### `cost_categories` Block

* `key` - (Optional) Cost category key to filter on.
* `match_options` - (Optional) Match options, such as `EQUALS` or `CONTAINS`.
* `values` - (Optional) Cost category values to match.

### `display_config` Block

Exactly one of the following blocks must be specified:

* `graph` - (Optional) Configuration blocks for a graphical display, one per metric. See [`graph`](#graph-block) below.
* `table` - (Optional) Empty configuration block (`table {}`) selecting a tabular display.

### `graph` Block

* `metric` - (Required) Metric the visualization applies to.
* `visual_type` - (Required) Type of visualization. Valid values: `LINE`, `BAR`, `STACK`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the dashboard.
* `created_at` - Timestamp when the dashboard was created.
* `dashboard_type` - Type of the dashboard.
* `id` - (**Deprecated**) ARN of the dashboard. Use `arn` instead.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Timestamp when the dashboard was last updated.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import BCM Dashboards Dashboard using the `arn`. For example:

```terraform
import {
  to = aws_bcmdashboards_dashboard.example
  id = "arn:aws:bcm-dashboards::123456789012:dashboard/01234567-89ab-cdef-0123-456789abcdef"
}
```

Using `terraform import`, import BCM Dashboards Dashboard using the `arn`. For example:

```console
% terraform import aws_bcmdashboards_dashboard.example arn:aws:bcm-dashboards::123456789012:dashboard/01234567-89ab-cdef-0123-456789abcdef
```
