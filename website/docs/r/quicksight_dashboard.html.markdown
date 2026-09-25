---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_dashboard"
description: |-
  Manages a QuickSight Dashboard.
---

# Resource: aws_quicksight_dashboard

Resource for managing a QuickSight Dashboard.

## Example Usage

### From Source Template

```terraform
resource "aws_quicksight_dashboard" "example" {
  dashboard_id        = "example-id"
  name                = "example-name"
  version_description = "version"
  source_entity {
    source_template {
      arn = aws_quicksight_template.source.arn
      data_set_references {
        data_set_arn         = aws_quicksight_data_set.dataset.arn
        data_set_placeholder = "1"
      }
    }
  }
}
```

### With Definition

```terraform
resource "aws_quicksight_dashboard" "example" {
  dashboard_id        = "example-id"
  name                = "example-name"
  version_description = "version"
  definition {
    data_set_identifiers_declarations {
      data_set_arn = aws_quicksight_data_set.dataset.arn
      identifier   = "1"
    }
    sheets {
      title    = "Example"
      sheet_id = "Example1"
      visuals {
        line_chart_visual {
          visual_id = "LineChart"
          title {
            format_text {
              plain_text = "Line Chart Example"
            }
          }
          chart_configuration {
            field_wells {
              line_chart_aggregated_field_wells {
                category {
                  categorical_dimension_field {
                    field_id = "1"
                    column {
                      data_set_identifier = "1"
                      column_name         = "Column1"
                    }
                  }
                }
                values {
                  categorical_measure_field {
                    field_id = "2"
                    column {
                      data_set_identifier = "1"
                      column_name         = "Column1"
                    }
                    aggregation_function = "COUNT"
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `dashboard_id` - (Required, Forces new resource) Identifier for the dashboard.
* `name` - (Required) Display name for the dashboard.
* `version_description` - (Required) Description of the current dashboard version being created/updated.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `dashboard_publish_options` - (Optional) Options for publishing the dashboard. See [`dashboard_publish_options`](#dashboard_publish_options-block).
* `definition` - (Optional) Detailed dashboard definition. Only one of `definition` or `source_entity` should be configured. See [`definition`](#definition-block).
* `parameters` - (Optional) Parameters for the creation of the dashboard, which you want to use to override the default settings. A dashboard can have any type of parameters, and some parameters might accept multiple values. See [`parameters`](#parameters-block).
* `permissions` - (Optional) Set of resource permissions on the dashboard. Maximum of 64 items. See [`permissions`](#permissions-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `source_entity` - (Optional) Entity that you are using as a source when you create the dashboard (template). Only one of `definition` or `source_entity` should be configured. See [`source_entity`](#source_entity-block).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `theme_arn` - (Optional) ARN of the theme that is being used for this dashboard. The theme ARN must exist in the same AWS account where you create the dashboard.

### `permissions` Block

* `actions` - (Required) List of IAM actions to grant or revoke permissions on.
* `principal` - (Required) ARN of the principal. See the [ResourcePermission documentation](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ResourcePermission.html) for the applicable ARN values.

### `source_entity` Block

* `source_template` - (Optional) Source template. See [`source_template`](#source_template-block).

### `source_template` Block

* `arn` - (Required) ARN of the resource.
* `data_set_references` - (Required) List of dataset references. See [`data_set_references`](#data_set_references-block).

### `data_set_references` Block

* `data_set_arn` - (Required) Dataset ARN.
* `data_set_placeholder` - (Required) Dataset placeholder.

### `dashboard_publish_options` Block

* `ad_hoc_filtering_option` - (Optional) Ad hoc (one-time) filtering option. See [`ad_hoc_filtering_option`](#ad_hoc_filtering_option-block).
* `data_point_drill_up_down_option` - (Optional) Drill-down options of data points in a dashboard. See [`data_point_drill_up_down_option`](#data_point_drill_up_down_option-block).
* `data_point_menu_label_option` - (Optional) Data point menu label options of a dashboard. See [`data_point_menu_label_option`](#data_point_menu_label_option-block).
* `data_point_tooltip_option` - (Optional) Data point tool tip options of a dashboard. See [`data_point_tooltip_option`](#data_point_tooltip_option-block).
* `export_to_csv_option` - (Optional) Export to .csv option. See [`export_to_csv_option`](#export_to_csv_option-block).
* `export_with_hidden_fields_option` - (Optional) Whether hidden fields are exported with a dashboard. See [`export_with_hidden_fields_option`](#export_with_hidden_fields_option-block).
* `sheet_controls_option` - (Optional) Sheet controls option. See [`sheet_controls_option`](#sheet_controls_option-block).
* `sheet_layout_element_maximization_option` - (Optional) Sheet layout maximization options of a dashboard. See [`sheet_layout_element_maximization_option`](#sheet_layout_element_maximization_option-block).
* `visual_axis_sort_option` - (Optional) Axis sort options of a dashboard. See [`visual_axis_sort_option`](#visual_axis_sort_option-block).
* `visual_menu_option` - (Optional) Menu options of a visual in a dashboard. See [`visual_menu_option`](#visual_menu_option-block).

### `ad_hoc_filtering_option` Block

* `availability_status` - (Optional) Availability status. Possibles values: ENABLED, DISABLED.

### `data_point_drill_up_down_option` Block

* `availability_status` - (Optional) Availability status. Possibles values: ENABLED, DISABLED.

### `data_point_menu_label_option` Block

* `availability_status` - (Optional) Availability status. Possibles values: ENABLED, DISABLED.

### `data_point_tooltip_option` Block

* `availability_status` - (Optional) Availability status. Possibles values: ENABLED, DISABLED.

### `export_to_csv_option` Block

* `availability_status` - (Optional) Availability status. Possibles values: ENABLED, DISABLED.

### `export_with_hidden_fields_option` Block

* `availability_status` - (Optional) Availability status. Possibles values: ENABLED, DISABLED.

### `sheet_controls_option` Block

* `visibility_state` - (Optional) Visibility state. Possibles values: EXPANDED, COLLAPSED.

### `sheet_layout_element_maximization_option` Block

* `availability_status` - (Optional) Availability status. Possibles values: ENABLED, DISABLED.

### `visual_axis_sort_option` Block

* `availability_status` - (Optional) Availability status. Possibles values: ENABLED, DISABLED.

### `visual_menu_option` Block

* `availability_status` - (Optional) Availability status. Possibles values: ENABLED, DISABLED.

### `parameters` Block

* `date_time_parameters` - (Optional) List of parameters that have a data type of date-time. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeParameter.html).
* `decimal_parameters` - (Optional) List of parameters that have a data type of decimal. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalParameter.html).
* `integer_parameters` - (Optional) List of parameters that have a data type of integer. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_IntegerParameter.html).
* `string_parameters` - (Optional) List of parameters that have a data type of string. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_StringParameter.html).

### `definition` Block

* `analysis_defaults` - (Optional) Configuration for default analysis settings. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AnalysisDefaults.html).
* `calculated_fields` - (Optional) List of calculated field definitions for the dashboard. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CalculatedField.html).
* `column_configurations` - (Optional) List of dashboard-level column configurations. Column configurations are used to set default formatting for a column that's used throughout a dashboard. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnConfiguration.html).
* `data_set_identifiers_declarations` - (Required) List dataset identifier declarations. With this mapping,you can use dataset identifiers instead of dataset ARNs throughout the dashboard's sub-structures. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetIdentifierDeclaration.html).
* `filter_groups` - (Optional) List of filter definitions for a dashboard. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterGroup.html). For more information, see [Filtering Data](https://docs.aws.amazon.com/quicksight/latest/user/filtering-visual-data.html) in Amazon QuickSight User Guide.
* `parameter_declarations` - (Optional) List of parameter declarations for a dashboard. Parameters are named variables that can transfer a value for use by an action or an object. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterDeclaration.html). For more information, see [Parameters in Amazon QuickSight](https://docs.aws.amazon.com/quicksight/latest/user/parameters-in-quicksight.html) in the Amazon QuickSight User Guide.
* `sheets` - (Optional) List of sheet definitions for a dashboard. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetDefinition.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the dashboard.
* `created_time` - Time that the dashboard was created.
* `id` - Comma-delimited string joining AWS account ID and dashboard ID.
* `last_published_time` - Time that the dashboard was last published.
* `last_updated_time` - Time that the dashboard was last updated.
* `source_entity_arn` - ARN of a template that was used to create this dashboard.
* `status` - Dashboard creation status.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `version_number` - Version number of the dashboard version.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a QuickSight Dashboard using the AWS account ID and dashboard ID separated by a comma (`,`). For example:

```terraform
import {
  to = aws_quicksight_dashboard.example
  id = "123456789012,example-id"
}
```

Using `terraform import`, import a QuickSight Dashboard using the AWS account ID and dashboard ID separated by a comma (`,`). For example:

```console
% terraform import aws_quicksight_dashboard.example 123456789012,example-id
```
