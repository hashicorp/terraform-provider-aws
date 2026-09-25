---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_template"
description: |-
  Manages a QuickSight Template.
---

# Resource: aws_quicksight_template

Resource for managing a QuickSight Template.

## Example Usage

### From Source Template

```terraform
resource "aws_quicksight_template" "example" {
  template_id         = "example-id"
  name                = "example-name"
  version_description = "version"
  source_entity {
    source_template {
      arn = aws_quicksight_template.source.arn
    }
  }
}
```

### With Definition

```terraform
resource "aws_quicksight_template" "example" {
  template_id         = "example-id"
  name                = "example-name"
  version_description = "version"
  definition {
    data_set_configuration {
      data_set_schema {
        column_schema_list {
          name      = "Column1"
          data_type = "STRING"
        }
        column_schema_list {
          name      = "Column2"
          data_type = "INTEGER"
        }
      }
      placeholder = "1"
    }
    sheets {
      title    = "Test"
      sheet_id = "Test1"
      visuals {
        bar_chart_visual {
          visual_id = "BarChart"
          chart_configuration {
            field_wells {
              bar_chart_aggregated_field_wells {
                category {
                  categorical_dimension_field {
                    field_id = "1"
                    column {
                      column_name         = "Column1"
                      data_set_identifier = "1"
                    }
                  }
                }
                values {
                  numerical_measure_field {
                    field_id = "2"
                    column {
                      column_name         = "Column2"
                      data_set_identifier = "1"
                    }
                    aggregation_function {
                      simple_numerical_aggregation = "SUM"
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
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Display name for the template.
* `template_id` - (Required, Forces new resource) Identifier for the template.
* `version_description` - (Required) Description of the current template version being created or updated.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `definition` - (Optional) Detailed template definition. Only one of `definition` or `source_entity` should be configured. See [below](#definition-block).
* `permissions` - (Optional) Set of resource permissions on the template. Maximum of 64 items. See [below](#permissions-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `source_entity` - (Optional) Entity used as the source when creating the template (an analysis or a template). Only one of `definition` or `source_entity` should be configured. See [below](#source_entity-block).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `permissions` Block

* `actions` - (Required) List of IAM actions to grant or revoke permissions on.
* `principal` - (Required) ARN of the principal. See the [ResourcePermission documentation](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ResourcePermission.html) for the applicable ARN values.

### `source_entity` Block

* `source_analysis` - (Optional) Source analysis, when the template is based on an analysis. Only one of `source_analysis` or `source_template` should be configured. See [below](#source_analysis-block).
* `source_template` - (Optional) Source template, when the template is based on another template. Only one of `source_analysis` or `source_template` should be configured. See [below](#source_template-block).

### `source_analysis` Block

* `arn` - (Required) ARN of the resource.
* `data_set_references` - (Required) Dataset references used as placeholders in the template. See [below](#data_set_references-block).

### `source_template` Block

* `arn` - (Required) ARN of the resource.

### `data_set_references` Block

* `data_set_arn` - (Required) Dataset ARN.
* `data_set_placeholder` - (Required) Dataset placeholder.

### `definition` Block

* `analysis_defaults` - (Optional) Configuration for default analysis settings. See [below](#analysis_defaults-block).
* `calculated_fields` - (Optional) Calculated field definitions for the template. See [below](#calculated_fields-block).
* `column_configurations` - (Optional) Column configurations that set default formatting for columns used throughout the template. See [below](#column_configurations-block).
* `data_set_configuration` - (Required) Dataset configurations that define the required columns for each dataset used within the template. See [below](#data_set_configuration-block).
* `filter_groups` - (Optional) Filter definitions for the template. See [below](#filter_groups-block).
* `parameters_declarations` - (Optional) Parameter declarations for the template. See [below](#parameters_declarations-block).
* `sheets` - (Optional) Sheet definitions for the template. See [below](#sheets-block).

### `action_operations` Block

* `filter_operation` - (Optional) See [below](#filter_operation-block).
* `navigation_operation` - (Optional) See [below](#navigation_operation-block).
* `set_parameters_operation` - (Optional) See [below](#set_parameters_operation-block).
* `url_operation` - (Optional) See [below](#url_operation-block).

### `actions` Block

* `action_operations` - (Required) See [below](#action_operations-block).
* `custom_action_id` - (Required)
* `name` - (Required)
* `status` - (Required)
* `trigger` - (Required)

### `actual_value` Block

* `icon` - (Optional) See [below](#icon-block).
* `text_color` - (Required) See [below](#text_color-block).

### `after` Block

* `status` - (Optional)

### `aggregation` Block

* `categorical_aggregation_function` - (Optional)
* `date_aggregation_function` - (Optional)
* `numerical_aggregation_function` - (Optional) See [below](#numerical_aggregation_function-block).

### `aggregation_function` Block

* `categorical_aggregation_function` - (Optional)
* `date_aggregation_function` - (Optional)
* `numerical_aggregation_function` - (Optional) See [below](#numerical_aggregation_function-block).
* `percentile_aggregation` - (Optional) See [below](#percentile_aggregation-block).
* `simple_numerical_aggregation` - (Optional)

### `aggregation_sort_configuration` Block

* `aggregation_function` - (Required) See [below](#aggregation_function-block).
* `column` - (Required) See [below](#column-block).
* `sort_direction` - (Required)

### `analysis_defaults` Block

* `default_new_sheet_configuration` - (Required) See [below](#default_new_sheet_configuration-block).

### `anchor_date_configuration` Block

* `anchor_option` - (Optional)
* `parameter_name` - (Optional)

### `apply_to` Block

* `column` - (Required) See [below](#column-block).
* `field_id` - (Required)

### `arc` Block

* `arc_angle` - (Optional)
* `arc_thickness` - (Optional)
* `foreground_color` - (Required) See [below](#foreground_color-block).

### `arc_axis` Block

* `range` - (Optional) See [below](#range-block).
* `reserve_range` - (Optional)

### `arc_options` Block

* `arc_thickness` - (Optional)

### `area_style_settings` Block

* `visibility` - (Optional)

### `axis_label_options` Block

* `apply_to` - (Optional) See [below](#apply_to-block).
* `custom_label` - (Optional)
* `font_configuration` - (Optional) See [below](#font_configuration-block).

### `axis_options` Block

* `axis_line_visibility` - (Optional)
* `axis_offset` - (Optional)
* `data_options` - (Optional) See [below](#data_options-block).
* `grid_line_visibility` - (Optional)
* `scrollbar_options` - (Optional) See [below](#scrollbar_options-block).
* `tick_label_options` - (Optional) See [below](#tick_label_options-block).

### `background_color` Block

* `gradient` - (Optional) See [below](#gradient-block).
* `solid` - (Optional) See [below](#solid-block).

### `background_style` Block

* `color` - (Optional)
* `visibility` - (Optional)

### `bar_chart_aggregated_field_wells` Block

* `category` - (Optional) See [below](#category-block).
* `colors` - (Optional) See [below](#colors-block).
* `small_multiples` - (Optional) See [below](#small_multiples-block).
* `values` - (Optional) See [below](#values-block).

### `bar_chart_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `bar_data_labels` Block

* `category_label_visibility` - (Optional)
* `data_label_types` - (Optional) See [below](#data_label_types-block).
* `label_color` - (Optional)
* `label_content` - (Optional)
* `label_font_configuration` - (Optional) See [below](#label_font_configuration-block).
* `measure_label_visibility` - (Optional)
* `overlap` - (Optional)
* `position` - (Optional)
* `visibility` - (Optional)

### `bar_values` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `base_series_settings` Block

* `area_style_settings` - (Optional) See [below](#area_style_settings-block).

### `bin_count` Block

* `value` - (Optional)

### `bin_options` Block

* `bin_count` - (Optional) See [below](#bin_count-block).
* `bin_width` - (Optional) See [below](#bin_width-block).
* `selected_bin_type` - (Optional)
* `start_value` - (Optional)

### `bin_width` Block

* `bin_count_limit` - (Optional)
* `value` - (Optional)

### `body_sections` Block

* `content` - (Required) See [below](#content-block).
* `page_break_configuration` - (Optional) See [below](#page_break_configuration-block).
* `section_id` - (Required)
* `style` - (Optional) See [below](#style-block).

### `border` Block

* `side_specific_border` - (Optional) See [below](#side_specific_border-block).
* `uniform_border` - (Required) See [below](#uniform_border-block).

### `border_style` Block

* `color` - (Optional)
* `visibility` - (Optional)

### `bottom` Block

* `color` - (Optional)
* `style` - (Optional)
* `thickness` - (Optional)

### `bounds` Block

* `east` - (Required)
* `north` - (Required)
* `south` - (Required)
* `west` - (Required)

### `box_plot_aggregated_field_wells` Block

* `group_by` - (Optional) See [below](#group_by-block).
* `values` - (Optional) See [below](#values-block).

### `box_plot_options` Block

* `all_data_points_visibility` - (Optional)
* `outlier_visibility` - (Optional)
* `style_options` - (Optional) See [below](#style_options-block).

### `box_plot_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `breakdown_items_limit` Block

* `items_limit` - (Optional)
* `other_categories` - (Required)

### `breakdowns` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `calculated_fields` Block

* `data_set_identifier` - (Required)
* `expression` - (Required)
* `name` - (Required)

### `calculated_measure_field` Block

* `expression` - (Required)
* `field_id` - (Required)

### `calculation` Block

* `percentile_aggregation` - (Optional) See [below](#percentile_aggregation-block).
* `simple_numerical_aggregation` - (Optional)

### `canvas_size_options` Block

* `paper_canvas_size_options` - (Optional) See [below](#paper_canvas_size_options-block).
* `screen_canvas_size_options` - (Optional) See [below](#screen_canvas_size_options-block).

### `cascading_control_configuration` Block

* `source_controls` - (Optional) See [below](#source_controls-block).

### `categorical_dimension_field` Block

* `column` - (Required) See [below](#column-block).
* `field_id` - (Required)
* `format_configuration` - (Optional) See [below](#format_configuration-block).
* `hierarchy_id` - (Optional)

### `categorical_measure_field` Block

* `aggregation_function` - (Optional)
* `column` - (Required) See [below](#column-block).
* `field_id` - (Required)
* `format_configuration` - (Optional) See [below](#format_configuration-block).

### `categories` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `category` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `category_axis` Block

* `axis_line_visibility` - (Optional)
* `axis_offset` - (Optional)
* `data_options` - (Optional) See [below](#data_options-block).
* `grid_line_visibility` - (Optional)
* `scrollbar_options` - (Optional) See [below](#scrollbar_options-block).
* `tick_label_options` - (Optional) See [below](#tick_label_options-block).

### `category_axis_display_options` Block

* `axis_line_visibility` - (Optional)
* `axis_offset` - (Optional)
* `data_options` - (Optional) See [below](#data_options-block).
* `grid_line_visibility` - (Optional)
* `scrollbar_options` - (Optional) See [below](#scrollbar_options-block).
* `tick_label_options` - (Optional) See [below](#tick_label_options-block).

### `category_axis_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

### `category_filter` Block

* `category_values` - (Required)
* `column` - (Required) See [below](#column-block).
* `configuration` - (Required) See [below](#configuration-block).
* `filter_id` - (Required)

### `category_items_limit` Block

* `items_limit` - (Optional)
* `other_categories` - (Required)

### `category_items_limit_configuration` Block

* `items_limit` - (Optional)
* `other_categories` - (Required)

### `category_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

### `category_sort` Block

* `column_sort` - (Optional) See [below](#column_sort-block).
* `field_sort` - (Optional) See [below](#field_sort-block).

### `cell` Block

* `field_id` - (Required)
* `scope` - (Optional) See [below](#scope-block).
* `text_format` - (Optional) See [below](#text_format-block).

### `cell_style` Block

* `background_color` - (Optional)
* `border` - (Optional) See [below](#border-block).
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `height` - (Optional)
* `horizontal_text_alignment` - (Optional)
* `text_wrap` - (Optional)
* `vertical_text_alignment` - (Optional)
* `visibility` - (Optional)

### `chart_configuration` Block

* `alternate_band_colors_visibility` - (Optional)
* `alternate_band_even_color` - (Optional)
* `alternate_band_odd_color` - (Optional)
* `bar_data_labels` - (Optional) See [below](#bar_data_labels-block).
* `bars_arrangement` - (Optional)
* `base_series_settings` - (Optional) See [below](#base_series_settings-block).
* `bin_options` - (Optional) See [below](#bin_options-block).
* `box_plot_options` - (Optional) See [below](#box_plot_options-block).
* `category_axis` - (Optional) See [below](#category_axis-block).
* `category_axis_display_options` - (Optional) See [below](#category_axis_display_options-block).
* `category_axis_label_options` - (Optional) See [below](#category_axis_label_options-block).
* `category_label_options` - (Optional) See [below](#category_label_options-block).
* `color_axis` - (Optional) See [below](#color_axis-block).
* `color_label_options` - (Optional) See [below](#color_label_options-block).
* `color_scale` - (Optional) See [below](#color_scale-block).
* `column_label_options` - (Optional) See [below](#column_label_options-block).
* `content_type` - (Optional)
* `content_url` - (Optional)
* `contribution_analysis_defaults` - (Optional) See [below](#contribution_analysis_defaults-block).
* `data_label_options` - (Optional) See [below](#data_label_options-block).
* `data_labels` - (Optional) See [below](#data_labels-block).
* `default_series_settings` - (Optional) See [below](#default_series_settings-block).
* `donut_options` - (Optional) See [below](#donut_options-block).
* `field_options` - (Optional) See [below](#field_options-block).
* `field_wells` - (Optional) See [below](#field_wells-block).
* `forecast_configurations` - (Optional) See [below](#forecast_configurations-block).
* `gauge_chart_options` - (Optional) See [below](#gauge_chart_options-block).
* `group_label_options` - (Optional) See [below](#group_label_options-block).
* `image_scaling` - (Optional)
* `kpi_options` - (Optional) See [below](#kpi_options-block).
* `legend` - (Optional) See [below](#legend-block).
* `line_data_labels` - (Optional) See [below](#line_data_labels-block).
* `map_style_options` - (Optional) See [below](#map_style_options-block).
* `orientation` - (Optional)
* `paginated_report_options` - (Optional) See [below](#paginated_report_options-block).
* `point_style_options` - (Optional) See [below](#point_style_options-block).
* `primary_y_axis_display_options` - (Optional) See [below](#primary_y_axis_display_options-block).
* `primary_y_axis_label_options` - (Optional) See [below](#primary_y_axis_label_options-block).
* `reference_lines` - (Optional) See [below](#reference_lines-block).
* `row_label_options` - (Optional) See [below](#row_label_options-block).
* `secondary_y_axis_display_options` - (Optional) See [below](#secondary_y_axis_display_options-block).
* `secondary_y_axis_label_options` - (Optional) See [below](#secondary_y_axis_label_options-block).
* `series` - (Optional) See [below](#series-block).
* `shape` - (Optional)
* `size_label_options` - (Optional) See [below](#size_label_options-block).
* `small_multiples_options` - (Optional) See [below](#small_multiples_options-block).
* `sort_configuration` - (Optional) See [below](#sort_configuration-block).
* `start_angle` - (Optional)
* `table_inline_visualizations` - (Optional) See [below](#table_inline_visualizations-block).
* `table_options` - (Optional) See [below](#table_options-block).
* `tooltip` - (Optional) See [below](#tooltip-block).
* `total_options` - (Optional) See [below](#total_options-block).
* `type` - (Optional)
* `value_axis` - (Optional) See [below](#value_axis-block).
* `value_label_options` - (Optional) See [below](#value_label_options-block).
* `visual_palette` - (Optional) See [below](#visual_palette-block).
* `waterfall_chart_options` - (Optional) See [below](#waterfall_chart_options-block).
* `window_options` - (Optional) See [below](#window_options-block).
* `word_cloud_options` - (Optional) See [below](#word_cloud_options-block).
* `x_axis_display_options` - (Optional) See [below](#x_axis_display_options-block).
* `x_axis_label_options` - (Optional) See [below](#x_axis_label_options-block).
* `y_axis_display_options` - (Optional) See [below](#y_axis_display_options-block).
* `y_axis_label_options` - (Optional) See [below](#y_axis_label_options-block).

### `cluster_marker` Block

* `simple_cluster_marker` - (Optional) See [below](#simple_cluster_marker-block).

### `cluster_marker_configuration` Block

* `cluster_marker` - (Optional) See [below](#cluster_marker-block).

### `color` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).
* `stops` - (Optional) See [below](#stops-block).

### `color_axis` Block

* `axis_line_visibility` - (Optional)
* `axis_offset` - (Optional)
* `data_options` - (Optional) See [below](#data_options-block).
* `grid_line_visibility` - (Optional)
* `scrollbar_options` - (Optional) See [below](#scrollbar_options-block).
* `tick_label_options` - (Optional) See [below](#tick_label_options-block).

### `color_items_limit` Block

* `items_limit` - (Optional)
* `other_categories` - (Required)

### `color_items_limit_configuration` Block

* `items_limit` - (Optional)
* `other_categories` - (Required)

### `color_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

### `color_map` Block

* `color` - (Required)
* `element` - (Required) See [below](#element-block).
* `time_granularity` - (Optional)

### `color_scale` Block

* `color_fill_type` - (Required)
* `colors` - (Required) See [below](#colors-block).
* `null_value_color` - (Optional) See [below](#null_value_color-block).

### `color_sort` Block

* `column_sort` - (Optional) See [below](#column_sort-block).
* `field_sort` - (Optional) See [below](#field_sort-block).

### `colors` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `color` - (Optional)
* `data_value` - (Optional)
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `column` Block

* `aggregation_function` - (Optional) See [below](#aggregation_function-block).
* `column_name` - (Required)
* `data_set_identifier` - (Required)
* `direction` - (Required)
* `sort_by` - (Required) See [below](#sort_by-block).

### `column_configurations` Block

* `column` - (Required) See [below](#column-block).
* `format_configuration` - (Optional) See [below](#format_configuration-block).
* `role` - (Optional)

### `column_group_column_schema_list` Block

* `name` - (Optional)

### `column_group_schema_list` Block

* `column_group_column_schema_list` - (Optional) See [below](#column_group_column_schema_list-block).
* `name` - (Optional)

### `column_header_style` Block

* `background_color` - (Optional)
* `border` - (Optional) See [below](#border-block).
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `height` - (Optional)
* `horizontal_text_alignment` - (Optional)
* `text_wrap` - (Optional)
* `vertical_text_alignment` - (Optional)
* `visibility` - (Optional)

### `column_hierarchies` Block

* `date_time_hierarchy` - (Optional) See [below](#date_time_hierarchy-block).
* `explicit_hierarchy` - (Optional) See [below](#explicit_hierarchy-block).
* `predefined_hierarchy` - (Optional) See [below](#predefined_hierarchy-block).

### `column_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

### `column_schema_list` Block

* `data_type` - (Optional)
* `geographic_role` - (Optional)
* `name` - (Optional)

### `column_sort` Block

* `aggregation_function` - (Optional) See [below](#aggregation_function-block).
* `direction` - (Required)
* `sort_by` - (Required) See [below](#sort_by-block).

### `column_subtotal_options` Block

* `custom_label` - (Optional)
* `field_level` - (Optional)
* `field_level_options` - (Optional) See [below](#field_level_options-block).
* `metric_header_cell_style` - (Optional) See [below](#metric_header_cell_style-block).
* `total_cell_style` - (Optional) See [below](#total_cell_style-block).
* `totals_visibility` - (Optional)
* `value_cell_style` - (Optional) See [below](#value_cell_style-block).

### `column_to_match` Block

* `column_name` - (Required)
* `data_set_identifier` - (Required)

### `column_tooltip_item` Block

* `aggregation` - (Optional) See [below](#aggregation-block).
* `column` - (Required) See [below](#column-block).
* `label` - (Optional)
* `visibility` - (Optional)

### `column_total_options` Block

* `custom_label` - (Optional)
* `metric_header_cell_style` - (Optional) See [below](#metric_header_cell_style-block).
* `placement` - (Optional)
* `scroll_status` - (Optional)
* `total_cell_style` - (Optional) See [below](#total_cell_style-block).
* `totals_visibility` - (Optional)
* `value_cell_style` - (Optional) See [below](#value_cell_style-block).

### `columns` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `column_name` - (Required)
* `data_set_identifier` - (Required)
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `combo_chart_aggregated_field_wells` Block

* `bar_values` - (Optional) See [below](#bar_values-block).
* `category` - (Optional) See [below](#category-block).
* `colors` - (Optional) See [below](#colors-block).
* `line_values` - (Optional) See [below](#line_values-block).

### `combo_chart_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `comparison` Block

* `comparison_format` - (Optional) See [below](#comparison_format-block).
* `comparison_method` - (Optional)

### `comparison_format` Block

* `number_display_format_configuration` - (Optional) See [below](#number_display_format_configuration-block).
* `percentage_display_format_configuration` - (Optional) See [below](#percentage_display_format_configuration-block).

### `comparison_value` Block

* `icon` - (Optional) See [below](#icon-block).
* `text_color` - (Required) See [below](#text_color-block).

### `computation` Block

* `forecast` - (Optional) See [below](#forecast-block).
* `growth_rate` - (Optional) See [below](#growth_rate-block).
* `maximum_minimum` - (Optional) See [below](#maximum_minimum-block).
* `metric_comparison` - (Optional) See [below](#metric_comparison-block).
* `period_over_period` - (Optional) See [below](#period_over_period-block).
* `period_to_date` - (Optional) See [below](#period_to_date-block).
* `top_bottom_movers` - (Optional) See [below](#top_bottom_movers-block).
* `top_bottom_ranked` - (Optional) See [below](#top_bottom_ranked-block).
* `total_aggregation` - (Optional) See [below](#total_aggregation-block).
* `unique_values` - (Optional) See [below](#unique_values-block).

### `conditional_formatting` Block

* `conditional_formatting_options` - (Required) See [below](#conditional_formatting_options-block).

### `conditional_formatting_options` Block

* `actual_value` - (Optional) See [below](#actual_value-block).
* `arc` - (Optional) See [below](#arc-block).
* `cell` - (Optional) See [below](#cell-block).
* `comparison_value` - (Optional) See [below](#comparison_value-block).
* `primary_value` - (Optional) See [below](#primary_value-block).
* `progress_bar` - (Optional) See [below](#progress_bar-block).
* `row` - (Optional) See [below](#row-block).
* `shape` - (Required) See [below](#shape-block).

### `configuration` Block

* `custom_filter_configuration` - (Optional) See [below](#custom_filter_configuration-block).
* `custom_filter_list_configuration` - (Optional) See [below](#custom_filter_list_configuration-block).
* `filter_list_configuration` - (Optional) See [below](#filter_list_configuration-block).
* `free_form_layout` - (Optional) See [below](#free_form_layout-block).
* `grid_layout` - (Optional) See [below](#grid_layout-block).
* `section_based_layout` - (Optional) See [below](#section_based_layout-block).

### `configuration_overrides` Block

* `visibility` - (Optional)

### `content` Block

* `custom_icon_content` - (Optional) See [below](#custom_icon_content-block).
* `custom_text_content` - (Optional) See [below](#custom_text_content-block).
* `layout` - (Optional) See [below](#layout-block).

### `contribution_analysis_defaults` Block

* `contributor_dimensions` - (Required) See [below](#contributor_dimensions-block).
* `measure_field_id` - (Required)

### `contributor_dimensions` Block

* `column_name` - (Required)
* `data_set_identifier` - (Required)

### `currency_display_format_configuration` Block

* `decimal_places_configuration` - (Optional) See [below](#decimal_places_configuration-block).
* `negative_value_configuration` - (Optional) See [below](#negative_value_configuration-block).
* `null_value_format_configuration` - (Optional) See [below](#null_value_format_configuration-block).
* `number_scale` - (Optional)
* `prefix` - (Optional)
* `separator_configuration` - (Optional) See [below](#separator_configuration-block).
* `suffix` - (Optional)
* `symbol` - (Optional)

### `custom_condition` Block

* `color` - (Optional)
* `display_configuration` - (Optional) See [below](#display_configuration-block).
* `expression` - (Required)
* `icon_options` - (Required) See [below](#icon_options-block).

### `custom_content_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `data_set_identifier` - (Required)
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `custom_filter_configuration` Block

* `category_value` - (Optional)
* `match_operator` - (Required)
* `null_option` - (Required)
* `parameter_name` - (Optional)
* `select_all_options` - (Optional)

### `custom_filter_list_configuration` Block

* `category_values` - (Optional)
* `match_operator` - (Required)
* `null_option` - (Required)
* `select_all_options` - (Optional)

### `custom_icon_content` Block

* `icon` - (Optional)

### `custom_label_configuration` Block

* `custom_label` - (Required)

### `custom_narrative` Block

* `narrative` - (Required)

### `custom_text_content` Block

* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `value` - (Optional)

### `custom_values` Block

* `date_time_values` - (Optional)
* `decimal_values` - (Optional)
* `integer_values` - (Optional)
* `string_values` - (Optional)

### `custom_values_configuration` Block

* `custom_values` - (Required) See [below](#custom_values-block).
* `include_null_value` - (Optional)

### `data_bars` Block

* `field_id` - (Required)
* `negative_color` - (Optional)
* `positive_color` - (Optional)

### `data_configuration` Block

* `axis_binding` - (Optional)
* `dynamic_configuration` - (Optional) See [below](#dynamic_configuration-block).
* `static_configuration` - (Optional) See [below](#static_configuration-block).

### `data_field_series_item` Block

* `axis_binding` - (Required)
* `field_id` - (Required)
* `field_value` - (Optional)
* `settings` - (Optional) See [below](#settings-block).

### `data_label_options` Block

* `category_label_visibility` - (Optional)
* `label_color` - (Optional)
* `label_font_configuration` - (Optional) See [below](#label_font_configuration-block).
* `measure_data_label_style` - (Optional)
* `measure_label_visibility` - (Optional)
* `position` - (Optional)
* `visibility` - (Optional)

### `data_label_types` Block

* `data_path_label_type` - (Optional) See [below](#data_path_label_type-block).
* `field_label_type` - (Optional) See [below](#field_label_type-block).
* `maximum_label_type` - (Optional) See [below](#maximum_label_type-block).
* `minimum_label_type` - (Optional) See [below](#minimum_label_type-block).
* `range_ends_label_type` - (Optional) See [below](#range_ends_label_type-block).

### `data_labels` Block

* `category_label_visibility` - (Optional)
* `data_label_types` - (Optional) See [below](#data_label_types-block).
* `label_color` - (Optional)
* `label_content` - (Optional)
* `label_font_configuration` - (Optional) See [below](#label_font_configuration-block).
* `measure_label_visibility` - (Optional)
* `overlap` - (Optional)
* `position` - (Optional)
* `visibility` - (Optional)

### `data_options` Block

* `date_axis_options` - (Optional) See [below](#date_axis_options-block).
* `numeric_axis_options` - (Optional) See [below](#numeric_axis_options-block).

### `data_path` Block

* `direction` - (Required)
* `sort_paths` - (Required) See [below](#sort_paths-block).

### `data_path_label_type` Block

* `field_id` - (Optional)
* `field_value` - (Optional)
* `visibility` - (Optional)

### `data_path_list` Block

* `field_id` - (Required)
* `field_value` - (Required)

### `data_path_options` Block

* `data_path_list` - (Required) See [below](#data_path_list-block).
* `width` - (Optional)

### `data_set_configuration` Block

* `column_group_schema_list` - (Optional) See [below](#column_group_schema_list-block).
* `data_set_schema` - (Optional) See [below](#data_set_schema-block).
* `placeholder` - (Optional)

### `data_set_schema` Block

* `column_schema_list` - (Optional) See [below](#column_schema_list-block).

### `date_axis_options` Block

* `missing_date_visibility` - (Optional)

### `date_dimension_field` Block

* `column` - (Required) See [below](#column-block).
* `date_granularity` - (Optional)
* `field_id` - (Required)
* `format_configuration` - (Optional) See [below](#format_configuration-block).
* `hierarchy_id` - (Optional)

### `date_measure_field` Block

* `aggregation_function` - (Optional)
* `column` - (Required) See [below](#column-block).
* `field_id` - (Required)
* `format_configuration` - (Optional) See [below](#format_configuration-block).

### `date_time_format_configuration` Block

* `date_time_format` - (Optional)
* `null_value_format_configuration` - (Optional) See [below](#null_value_format_configuration-block).
* `numeric_format_configuration` - (Optional) See [below](#numeric_format_configuration-block).

### `date_time_hierarchy` Block

* `drill_down_filters` - (Optional) See [below](#drill_down_filters-block).
* `hierarchy_id` - (Required)

### `date_time_parameter_declaration` Block

* `default_values` - (Optional) See [below](#default_values-block).
* `name` - (Required)
* `time_granularity` - (Optional)
* `values_when_unset` - (Optional) See [below](#values_when_unset-block).

### `date_time_picker` Block

* `display_options` - (Optional) See [below](#display_options-block).
* `filter_control_id` - (Required)
* `parameter_control_id` - (Required)
* `source_filter_id` - (Required)
* `source_parameter_name` - (Required)
* `title` - (Required)
* `type` - (Optional)

### `decimal_parameter_declaration` Block

* `default_values` - (Optional) See [below](#default_values-block).
* `name` - (Required)
* `parameter_value_type` - (Required)
* `values_when_unset` - (Optional) See [below](#values_when_unset-block).

### `decimal_places_configuration` Block

* `decimal_places` - (Required)

### `default_new_sheet_configuration` Block

* `interactive_layout_configuration` - (Optional) See [below](#interactive_layout_configuration-block).
* `paginated_layout_configuration` - (Optional) See [below](#paginated_layout_configuration-block).
* `sheet_content_type` - (Optional)

### `default_series_settings` Block

* `axis_binding` - (Optional)
* `line_style_settings` - (Optional) See [below](#line_style_settings-block).
* `marker_style_settings` - (Optional) See [below](#marker_style_settings-block).

### `default_value_column` Block

* `column_name` - (Required)
* `data_set_identifier` - (Required)

### `default_values` Block

* `dynamic_value` - (Optional) See [below](#dynamic_value-block).
* `rolling_date` - (Optional) See [below](#rolling_date-block).
* `static_values` - (Optional)

### `destination` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `destination_items_limit` Block

* `items_limit` - (Optional)
* `other_categories` - (Required)

### `display_configuration` Block

* `icon_display_option` - (Optional)

### `display_options` Block

* `date_time_format` - (Optional)
* `placeholder_options` - (Optional) See [below](#placeholder_options-block).
* `search_options` - (Optional) See [below](#search_options-block).
* `select_all_options` - (Optional) See [below](#select_all_options-block).
* `title_options` - (Optional) See [below](#title_options-block).

### `donut_center_options` Block

* `label_visibility` - (Optional)

### `donut_options` Block

* `arc_options` - (Optional) See [below](#arc_options-block).
* `donut_center_options` - (Optional) See [below](#donut_center_options-block).

### `drill_down_filters` Block

* `category_filter` - (Optional) See [below](#category_filter-block).
* `numeric_equality_filter` - (Optional) See [below](#numeric_equality_filter-block).
* `time_range_filter` - (Optional) See [below](#time_range_filter-block).

### `dropdown` Block

* `cascading_control_configuration` - (Optional) See [below](#cascading_control_configuration-block).
* `display_options` - (Optional) See [below](#display_options-block).
* `filter_control_id` - (Required)
* `parameter_control_id` - (Required)
* `selectable_values` - (Optional) See [below](#selectable_values-block).
* `source_filter_id` - (Required)
* `source_parameter_name` - (Required)
* `title` - (Required)
* `type` - (Optional)

### `dynamic_configuration` Block

* `calculation` - (Required) See [below](#calculation-block).
* `column` - (Required) See [below](#column-block).
* `measure_aggregation_function` - (Required) See [below](#measure_aggregation_function-block).

### `dynamic_value` Block

* `default_value_column` - (Required) See [below](#default_value_column-block).
* `group_name_column` - (Optional) See [below](#group_name_column-block).
* `user_name_column` - (Optional) See [below](#user_name_column-block).

### `element` Block

* `field_id` - (Required)
* `field_value` - (Required)

### `elements` Block

* `background_style` - (Optional) See [below](#background_style-block).
* `border_style` - (Optional) See [below](#border_style-block).
* `column_index` - (Optional)
* `column_span` - (Required)
* `element_id` - (Required)
* `element_type` - (Required)
* `height` - (Required)
* `loading_animation` - (Optional) See [below](#loading_animation-block).
* `rendering_rules` - (Optional) See [below](#rendering_rules-block).
* `row_index` - (Optional)
* `row_span` - (Required)
* `selected_border_style` - (Optional) See [below](#selected_border_style-block).
* `visibility` - (Optional)
* `width` - (Required)
* `x_axis_location` - (Required)
* `y_axis_location` - (Required)

### `empty_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `data_set_identifier` - (Required)
* `visual_id` - (Required)

### `exclude_period_configuration` Block

* `amount` - (Required)
* `granularity` - (Required)
* `status` - (Optional)

### `explicit_hierarchy` Block

* `columns` - (Required) See [below](#columns-block).
* `drill_down_filters` - (Optional) See [below](#drill_down_filters-block).
* `hierarchy_id` - (Required)

### `field` Block

* `direction` - (Required)
* `field_id` - (Required)

### `field_base_tooltip` Block

* `aggregation_visibility` - (Optional)
* `tooltip_fields` - (Optional) See [below](#tooltip_fields-block).
* `tooltip_title_type` - (Optional)

### `field_label_type` Block

* `field_id` - (Optional)
* `visibility` - (Optional)

### `field_level_options` Block

* `field_id` - (Optional)

### `field_options` Block

* `data_path_options` - (Optional) See [below](#data_path_options-block).
* `order` - (Optional)
* `selected_field_options` - (Optional) See [below](#selected_field_options-block).

### `field_series_item` Block

* `axis_binding` - (Required)
* `field_id` - (Required)
* `settings` - (Optional) See [below](#settings-block).

### `field_sort` Block

* `direction` - (Required)
* `field_id` - (Required)

### `field_sort_options` Block

* `field_id` - (Required)
* `sort_by` - (Required) See [below](#sort_by-block).

### `field_tooltip_item` Block

* `field_id` - (Required)
* `label` - (Optional)
* `visibility` - (Optional)

### `field_wells` Block

* `bar_chart_aggregated_field_wells` - (Optional) See [below](#bar_chart_aggregated_field_wells-block).
* `box_plot_aggregated_field_wells` - (Optional) See [below](#box_plot_aggregated_field_wells-block).
* `combo_chart_aggregated_field_wells` - (Optional) See [below](#combo_chart_aggregated_field_wells-block).
* `filled_map_aggregated_field_wells` - (Optional) See [below](#filled_map_aggregated_field_wells-block).
* `funnel_chart_aggregated_field_wells` - (Optional) See [below](#funnel_chart_aggregated_field_wells-block).
* `geospatial_map_aggregated_field_wells` - (Optional) See [below](#geospatial_map_aggregated_field_wells-block).
* `heat_map_aggregated_field_wells` - (Optional) See [below](#heat_map_aggregated_field_wells-block).
* `histogram_aggregated_field_wells` - (Optional) See [below](#histogram_aggregated_field_wells-block).
* `line_chart_aggregated_field_wells` - (Optional) See [below](#line_chart_aggregated_field_wells-block).
* `pie_chart_aggregated_field_wells` - (Optional) See [below](#pie_chart_aggregated_field_wells-block).
* `pivot_table_aggregated_field_wells` - (Optional) See [below](#pivot_table_aggregated_field_wells-block).
* `radar_chart_aggregated_field_wells` - (Optional) See [below](#radar_chart_aggregated_field_wells-block).
* `sankey_diagram_aggregated_field_wells` - (Optional) See [below](#sankey_diagram_aggregated_field_wells-block).
* `scatter_plot_categorically_aggregated_field_wells` - (Optional) See [below](#scatter_plot_categorically_aggregated_field_wells-block).
* `scatter_plot_unaggregated_field_wells` - (Optional) See [below](#scatter_plot_unaggregated_field_wells-block).
* `table_aggregated_field_wells` - (Optional) See [below](#table_aggregated_field_wells-block).
* `table_unaggregated_field_wells` - (Optional) See [below](#table_unaggregated_field_wells-block).
* `target_values` - (Optional) See [below](#target_values-block).
* `tree_map_aggregated_field_wells` - (Optional) See [below](#tree_map_aggregated_field_wells-block).
* `trend_groups` - (Optional) See [below](#trend_groups-block).
* `values` - (Optional) See [below](#values-block).
* `waterfall_chart_aggregated_field_wells` - (Optional) See [below](#waterfall_chart_aggregated_field_wells-block).
* `word_cloud_aggregated_field_wells` - (Optional) See [below](#word_cloud_aggregated_field_wells-block).

### `filled_map_aggregated_field_wells` Block

* `geospatial` - (Optional) See [below](#geospatial-block).
* `values` - (Optional) See [below](#values-block).

### `filled_map_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `conditional_formatting` - (Optional) See [below](#conditional_formatting-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `filter_controls` Block

* `date_time_picker` - (Optional) See [below](#date_time_picker-block).
* `dropdown` - (Optional) See [below](#dropdown-block).
* `list` - (Optional) See [below](#list-block).
* `relative_date_time` - (Optional) See [below](#relative_date_time-block).
* `slider` - (Optional) See [below](#slider-block).
* `text_area` - (Optional) See [below](#text_area-block).
* `text_field` - (Optional) See [below](#text_field-block).

### `filter_groups` Block

* `cross_dataset` - (Required)
* `filter_group_id` - (Required)
* `filters` - (Required) See [below](#filters-block).
* `scope_configuration` - (Required) See [below](#scope_configuration-block).
* `status` - (Optional)

### `filter_list_configuration` Block

* `category_values` - (Optional)
* `match_operator` - (Required)
* `select_all_options` - (Optional)

### `filter_operation` Block

* `selected_fields_configuration` - (Required) See [below](#selected_fields_configuration-block).
* `target_visuals_configuration` - (Required) See [below](#target_visuals_configuration-block).

### `filters` Block

* `category_filter` - (Optional) See [below](#category_filter-block).
* `numeric_equality_filter` - (Optional) See [below](#numeric_equality_filter-block).
* `numeric_range_filter` - (Optional) See [below](#numeric_range_filter-block).
* `relative_dates_filter` - (Optional) See [below](#relative_dates_filter-block).
* `time_equality_filter` - (Optional) See [below](#time_equality_filter-block).
* `time_range_filter` - (Optional) See [below](#time_range_filter-block).
* `top_bottom_filter` - (Optional) See [below](#top_bottom_filter-block).

### `font_configuration` Block

* `font_color` - (Optional)
* `font_decoration` - (Optional)
* `font_size` - (Optional) See [below](#font_size-block).
* `font_style` - (Optional)
* `font_weight` - (Optional) See [below](#font_weight-block).

### `font_size` Block

* `relative` - (Optional)

### `font_weight` Block

* `name` - (Optional)

### `footer_sections` Block

* `layout` - (Optional) See [below](#layout-block).
* `section_id` - (Required)
* `style` - (Optional) See [below](#style-block).

### `forecast` Block

* `computation_id` - (Required)
* `custom_seasonality_value` - (Optional)
* `lower_boundary` - (Optional)
* `name` - (Optional)
* `periods_backward` - (Optional)
* `periods_forward` - (Optional)
* `prediction_interval` - (Optional)
* `seasonality` - (Required)
* `time` - (Optional) See [below](#time-block).
* `upper_boundary` - (Optional)
* `value` - (Optional) See [below](#value-block).

### `forecast_configurations` Block

* `forecast_properties` - (Optional) See [below](#forecast_properties-block).
* `scenario` - (Optional) See [below](#scenario-block).

### `forecast_properties` Block

* `lower_boundary` - (Optional)
* `periods_backward` - (Optional)
* `periods_forward` - (Optional)
* `prediction_interval` - (Optional)
* `seasonality` - (Optional)
* `upper_boundary` - (Optional)

### `foreground_color` Block

* `gradient` - (Optional) See [below](#gradient-block).
* `solid` - (Optional) See [below](#solid-block).

### `format` Block

* `background_color` - (Required) See [below](#background_color-block).

### `format_configuration` Block

* `currency_display_format_configuration` - (Optional) See [below](#currency_display_format_configuration-block).
* `date_time_format` - (Optional)
* `date_time_format_configuration` - (Optional) See [below](#date_time_format_configuration-block).
* `null_value_format_configuration` - (Optional) See [below](#null_value_format_configuration-block).
* `number_display_format_configuration` - (Optional) See [below](#number_display_format_configuration-block).
* `number_format_configuration` - (Optional) See [below](#number_format_configuration-block).
* `numeric_format_configuration` - (Optional) See [below](#numeric_format_configuration-block).
* `percentage_display_format_configuration` - (Optional) See [below](#percentage_display_format_configuration-block).
* `string_format_configuration` - (Optional) See [below](#string_format_configuration-block).

### `format_text` Block

* `plain_text` - (Optional)
* `rich_text` - (Optional)

### `free_form` Block

* `canvas_size_options` - (Required) See [below](#canvas_size_options-block).

### `free_form_layout` Block

* `canvas_size_options` - (Optional) See [below](#canvas_size_options-block).
* `elements` - (Required) See [below](#elements-block).

### `from_value` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `funnel_chart_aggregated_field_wells` Block

* `category` - (Optional) See [below](#category-block).
* `values` - (Optional) See [below](#values-block).

### `funnel_chart_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `gauge_chart_options` Block

* `arc` - (Optional) See [below](#arc-block).
* `arc_axis` - (Optional) See [below](#arc_axis-block).
* `comparison` - (Optional) See [below](#comparison-block).
* `primary_value_display_type` - (Optional)
* `primary_value_font_configuration` - (Optional) See [below](#primary_value_font_configuration-block).

### `gauge_chart_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `conditional_formatting` - (Optional) See [below](#conditional_formatting-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `geospatial` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `geospatial_map_aggregated_field_wells` Block

* `colors` - (Optional) See [below](#colors-block).
* `geospatial` - (Optional) See [below](#geospatial-block).
* `values` - (Optional) See [below](#values-block).

### `geospatial_map_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `gradient` Block

* `color` - (Required) See [below](#color-block).
* `expression` - (Required)

### `grid` Block

* `canvas_size_options` - (Required) See [below](#canvas_size_options-block).

### `grid_layout` Block

* `canvas_size_options` - (Optional) See [below](#canvas_size_options-block).
* `elements` - (Required) See [below](#elements-block).

### `group_by` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `group_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

### `group_name_column` Block

* `column_name` - (Required)
* `data_set_identifier` - (Required)

### `groups` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `growth_rate` Block

* `computation_id` - (Required)
* `name` - (Optional)
* `period_size` - (Optional)
* `time` - (Optional) See [below](#time-block).
* `value` - (Optional) See [below](#value-block).

### `header_sections` Block

* `layout` - (Optional) See [below](#layout-block).
* `section_id` - (Required)
* `style` - (Optional) See [below](#style-block).

### `header_style` Block

* `background_color` - (Optional)
* `border` - (Optional) See [below](#border-block).
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `height` - (Optional)
* `horizontal_text_alignment` - (Optional)
* `text_wrap` - (Optional)
* `vertical_text_alignment` - (Optional)
* `visibility` - (Optional)

### `heat_map_aggregated_field_wells` Block

* `columns` - (Optional) See [below](#columns-block).
* `rows` - (Optional) See [below](#rows-block).
* `values` - (Optional) See [below](#values-block).

### `heat_map_column_items_limit_configuration` Block

* `items_limit` - (Optional)
* `other_categories` - (Required)

### `heat_map_column_sort` Block

* `column_sort` - (Optional) See [below](#column_sort-block).
* `field_sort` - (Optional) See [below](#field_sort-block).

### `heat_map_row_items_limit_configuration` Block

* `items_limit` - (Optional)
* `other_categories` - (Required)

### `heat_map_row_sort` Block

* `column_sort` - (Optional) See [below](#column_sort-block).
* `field_sort` - (Optional) See [below](#field_sort-block).

### `heat_map_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `histogram_aggregated_field_wells` Block

* `values` - (Optional) See [below](#values-block).

### `histogram_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `icon` Block

* `custom_condition` - (Optional) See [below](#custom_condition-block).
* `icon_set` - (Optional) See [below](#icon_set-block).

### `icon_options` Block

* `icon` - (Optional)
* `unicode_icon` - (Optional)

### `icon_set` Block

* `expression` - (Required)
* `icon_set_type` - (Optional)

### `image_configuration` Block

* `sizing_options` - (Optional) See [below](#sizing_options-block).

### `inner_horizontal` Block

* `color` - (Optional)
* `style` - (Optional)
* `thickness` - (Optional)

### `inner_vertical` Block

* `color` - (Optional)
* `style` - (Optional)
* `thickness` - (Optional)

### `insight_configuration` Block

* `computation` - (Optional) See [below](#computation-block).
* `custom_narrative` - (Optional) See [below](#custom_narrative-block).

### `insight_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `data_set_identifier` - (Required)
* `insight_configuration` - (Optional) See [below](#insight_configuration-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `integer_parameter_declaration` Block

* `default_values` - (Optional) See [below](#default_values-block).
* `name` - (Required)
* `parameter_value_type` - (Required)
* `values_when_unset` - (Optional) See [below](#values_when_unset-block).

### `interactive_layout_configuration` Block

* `free_form` - (Optional) See [below](#free_form-block).
* `grid` - (Optional) See [below](#grid-block).

### `kpi_options` Block

* `comparison` - (Optional) See [below](#comparison-block).
* `primary_value_display_type` - (Optional)
* `primary_value_font_configuration` - (Optional) See [below](#primary_value_font_configuration-block).
* `progress_bar` - (Optional) See [below](#progress_bar-block).
* `secondary_value` - (Optional) See [below](#secondary_value-block).
* `secondary_value_font_configuration` - (Optional) See [below](#secondary_value_font_configuration-block).
* `sparkline` - (Optional) See [below](#sparkline-block).
* `trend_arrows` - (Optional) See [below](#trend_arrows-block).
* `visual_layout_options` - (Optional) See [below](#visual_layout_options-block).

### `kpi_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `conditional_formatting` - (Optional) See [below](#conditional_formatting-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `label_configuration` Block

* `custom_label_configuration` - (Optional) See [below](#custom_label_configuration-block).
* `font_color` - (Optional)
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `horizontal_position` - (Optional)
* `value_label_configuration` - (Optional) See [below](#value_label_configuration-block).
* `vertical_position` - (Optional)

### `label_font_configuration` Block

* `font_color` - (Optional)
* `font_decoration` - (Optional)
* `font_size` - (Optional) See [below](#font_size-block).
* `font_style` - (Optional)
* `font_weight` - (Optional) See [below](#font_weight-block).

### `label_options` Block

* `custom_label` - (Optional)
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `visibility` - (Optional)

### `layout` Block

* `free_form_layout` - (Required) See [below](#free_form_layout-block).

### `layouts` Block

* `configuration` - (Required) See [below](#configuration-block).

### `left` Block

* `color` - (Optional)
* `style` - (Optional)
* `thickness` - (Optional)

### `legend` Block

* `height` - (Optional)
* `position` - (Optional)
* `title` - (Optional) See [below](#title-block).
* `visibility` - (Optional)
* `width` - (Optional)

### `line_chart_aggregated_field_wells` Block

* `category` - (Optional) See [below](#category-block).
* `colors` - (Optional) See [below](#colors-block).
* `small_multiples` - (Optional) See [below](#small_multiples-block).
* `values` - (Optional) See [below](#values-block).

### `line_chart_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `line_data_labels` Block

* `category_label_visibility` - (Optional)
* `data_label_types` - (Optional) See [below](#data_label_types-block).
* `label_color` - (Optional)
* `label_content` - (Optional)
* `label_font_configuration` - (Optional) See [below](#label_font_configuration-block).
* `measure_label_visibility` - (Optional)
* `overlap` - (Optional)
* `position` - (Optional)
* `visibility` - (Optional)

### `line_style_settings` Block

* `line_interpolation` - (Optional)
* `line_style` - (Optional)
* `line_visibility` - (Optional)
* `line_width` - (Optional)

### `line_values` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `linear` Block

* `step_count` - (Optional)
* `step_size` - (Optional)

### `link_configuration` Block

* `content` - (Optional) See [below](#content-block).
* `target` - (Optional)

### `link_to_data_set_column` Block

* `column_name` - (Required)
* `data_set_identifier` - (Required)

### `list` Block

* `cascading_control_configuration` - (Optional) See [below](#cascading_control_configuration-block).
* `display_options` - (Optional) See [below](#display_options-block).
* `filter_control_id` - (Required)
* `parameter_control_id` - (Required)
* `selectable_values` - (Optional) See [below](#selectable_values-block).
* `source_filter_id` - (Required)
* `source_parameter_name` - (Required)
* `title` - (Required)
* `type` - (Optional)

### `loading_animation` Block

* `visibility` - (Optional)

### `local_navigation_configuration` Block

* `target_sheet_id` - (Required)

### `logarithmic` Block

* `base` - (Optional)

### `map_style_options` Block

* `base_map_style` - (Optional)

### `marker_style_settings` Block

* `marker_color` - (Optional)
* `marker_shape` - (Optional)
* `marker_size` - (Optional)
* `marker_visibility` - (Optional)

### `maximum_label_type` Block

* `visibility` - (Optional)

### `maximum_minimum` Block

* `computation_id` - (Required)
* `name` - (Optional)
* `time` - (Optional) See [below](#time-block).
* `type` - (Required)
* `value` - (Optional) See [below](#value-block).

### `measure_aggregation_function` Block

* `categorical_aggregation_function` - (Optional)
* `date_aggregation_function` - (Optional)
* `numerical_aggregation_function` - (Optional) See [below](#numerical_aggregation_function-block).

### `metric_comparison` Block

* `computation_id` - (Required)
* `from_value` - (Optional) See [below](#from_value-block).
* `name` - (Optional)
* `target_value` - (Optional) See [below](#target_value-block).
* `time` - (Optional) See [below](#time-block).

### `metric_header_cell_style` Block

* `background_color` - (Optional)
* `border` - (Optional) See [below](#border-block).
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `height` - (Optional)
* `horizontal_text_alignment` - (Optional)
* `text_wrap` - (Optional)
* `vertical_text_alignment` - (Optional)
* `visibility` - (Optional)

### `min_max` Block

* `maximum` - (Optional)
* `minimum` - (Optional)

### `minimum_label_type` Block

* `visibility` - (Optional)

### `missing_data_configuration` Block

* `treatment_option` - (Optional)

### `navigation_operation` Block

* `local_navigation_configuration` - (Optional) See [below](#local_navigation_configuration-block).

### `negative_value_configuration` Block

* `display_mode` - (Required)

### `null_value_color` Block

* `color` - (Optional)
* `data_value` - (Optional)

### `null_value_format_configuration` Block

* `null_string` - (Required)

### `number_display_format_configuration` Block

* `decimal_places_configuration` - (Optional) See [below](#decimal_places_configuration-block).
* `negative_value_configuration` - (Optional) See [below](#negative_value_configuration-block).
* `null_value_format_configuration` - (Optional) See [below](#null_value_format_configuration-block).
* `number_scale` - (Optional)
* `prefix` - (Optional)
* `separator_configuration` - (Optional) See [below](#separator_configuration-block).
* `suffix` - (Optional)

### `number_format_configuration` Block

* `numeric_format_configuration` - (Optional) See [below](#numeric_format_configuration-block).

### `numeric_axis_options` Block

* `range` - (Optional) See [below](#range-block).
* `scale` - (Optional) See [below](#scale-block).

### `numeric_equality_filter` Block

* `aggregation_function` - (Optional) See [below](#aggregation_function-block).
* `column` - (Required) See [below](#column-block).
* `filter_id` - (Required)
* `match_operator` - (Required)
* `null_option` - (Required)
* `parameter_name` - (Optional)
* `select_all_options` - (Optional)
* `value` - (Optional)

### `numeric_format_configuration` Block

* `currency_display_format_configuration` - (Optional) See [below](#currency_display_format_configuration-block).
* `number_display_format_configuration` - (Optional) See [below](#number_display_format_configuration-block).
* `percentage_display_format_configuration` - (Optional) See [below](#percentage_display_format_configuration-block).

### `numeric_range_filter` Block

* `aggregation_function` - (Optional) See [below](#aggregation_function-block).
* `column` - (Required) See [below](#column-block).
* `filter_id` - (Required)
* `include_maximum` - (Optional)
* `include_minimum` - (Optional)
* `null_option` - (Required)
* `range_maximum` - (Optional) See [below](#range_maximum-block).
* `range_minimum` - (Optional) See [below](#range_minimum-block).
* `select_all_options` - (Optional)

### `numerical_aggregation_function` Block

* `percentile_aggregation` - (Optional) See [below](#percentile_aggregation-block).
* `simple_numerical_aggregation` - (Optional)

### `numerical_dimension_field` Block

* `column` - (Required) See [below](#column-block).
* `field_id` - (Required)
* `format_configuration` - (Optional) See [below](#format_configuration-block).
* `hierarchy_id` - (Optional)

### `numerical_measure_field` Block

* `aggregation_function` - (Optional) See [below](#aggregation_function-block).
* `column` - (Required) See [below](#column-block).
* `field_id` - (Required)
* `format_configuration` - (Optional) See [below](#format_configuration-block).

### `padding` Block

* `bottom` - (Optional)
* `left` - (Optional)
* `right` - (Optional)
* `top` - (Optional)

### `page_break_configuration` Block

* `after` - (Optional) See [below](#after-block).

### `paginated_layout_configuration` Block

* `section_based` - (Optional) See [below](#section_based-block).

### `paginated_report_options` Block

* `overflow_column_header_visibility` - (Optional)
* `vertical_overflow_visibility` - (Optional)

### `pagination_configuration` Block

* `page_number` - (Required)
* `page_size` - (Required)

### `panel_configuration` Block

* `background_color` - (Optional)
* `background_visibility` - (Optional)
* `border_color` - (Optional)
* `border_style` - (Optional)
* `border_thickness` - (Optional)
* `border_visibility` - (Optional)
* `gutter_spacing` - (Optional)
* `gutter_visibility` - (Optional)
* `title` - (Optional) See [below](#title-block).

### `paper_canvas_size_options` Block

* `paper_margin` - (Optional) See [below](#paper_margin-block).
* `paper_orientation` - (Optional)
* `paper_size` - (Optional)

### `paper_margin` Block

* `bottom` - (Optional)
* `left` - (Optional)
* `right` - (Optional)
* `top` - (Optional)

### `parameter_controls` Block

* `date_time_picker` - (Optional) See [below](#date_time_picker-block).
* `dropdown` - (Optional) See [below](#dropdown-block).
* `list` - (Optional) See [below](#list-block).
* `slider` - (Optional) See [below](#slider-block).
* `text_area` - (Optional) See [below](#text_area-block).
* `text_field` - (Optional) See [below](#text_field-block).

### `parameter_value_configurations` Block

* `destination_parameter_name` - (Required)
* `value` - (Required) See [below](#value-block).

### `parameters_declarations` Block

* `date_time_parameter_declaration` - (Optional) See [below](#date_time_parameter_declaration-block).
* `decimal_parameter_declaration` - (Optional) See [below](#decimal_parameter_declaration-block).
* `integer_parameter_declaration` - (Optional) See [below](#integer_parameter_declaration-block).
* `string_parameter_declaration` - (Optional) See [below](#string_parameter_declaration-block).

### `percent_range` Block

* `from` - (Optional)
* `to` - (Optional)

### `percentage_display_format_configuration` Block

* `decimal_places_configuration` - (Optional) See [below](#decimal_places_configuration-block).
* `negative_value_configuration` - (Optional) See [below](#negative_value_configuration-block).
* `null_value_format_configuration` - (Optional) See [below](#null_value_format_configuration-block).
* `prefix` - (Optional)
* `separator_configuration` - (Optional) See [below](#separator_configuration-block).
* `suffix` - (Optional)

### `percentile_aggregation` Block

* `percentile_value` - (Optional)

### `period_over_period` Block

* `computation_id` - (Required)
* `name` - (Optional)
* `time` - (Optional) See [below](#time-block).
* `value` - (Optional) See [below](#value-block).

### `period_to_date` Block

* `computation_id` - (Required)
* `name` - (Optional)
* `period_time_granularity` - (Required)
* `time` - (Optional) See [below](#time-block).
* `value` - (Optional) See [below](#value-block).

### `pie_chart_aggregated_field_wells` Block

* `category` - (Optional) See [below](#category-block).
* `small_multiples` - (Optional) See [below](#small_multiples-block).
* `values` - (Optional) See [below](#values-block).

### `pie_chart_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `pivot_table_aggregated_field_wells` Block

* `columns` - (Optional) See [below](#columns-block).
* `rows` - (Optional) See [below](#rows-block).
* `values` - (Optional) See [below](#values-block).

### `pivot_table_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `conditional_formatting` - (Optional) See [below](#conditional_formatting-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `placeholder_options` Block

* `visibility` - (Optional)

### `point_style_options` Block

* `cluster_marker_configuration` - (Optional) See [below](#cluster_marker_configuration-block).
* `selected_point_style` - (Optional)

### `predefined_hierarchy` Block

* `columns` - (Required) See [below](#columns-block).
* `drill_down_filters` - (Optional) See [below](#drill_down_filters-block).
* `hierarchy_id` - (Required)

### `primary_value` Block

* `icon` - (Optional) See [below](#icon-block).
* `text_color` - (Required) See [below](#text_color-block).

### `primary_value_font_configuration` Block

* `font_color` - (Optional)
* `font_decoration` - (Optional)
* `font_size` - (Optional) See [below](#font_size-block).
* `font_style` - (Optional)
* `font_weight` - (Optional) See [below](#font_weight-block).

### `primary_y_axis_display_options` Block

* `axis_line_visibility` - (Optional)
* `axis_offset` - (Optional)
* `axis_options` - (Optional) See [below](#axis_options-block).
* `data_options` - (Optional) See [below](#data_options-block).
* `grid_line_visibility` - (Optional)
* `missing_data_configuration` - (Optional) See [below](#missing_data_configuration-block).
* `scrollbar_options` - (Optional) See [below](#scrollbar_options-block).
* `tick_label_options` - (Optional) See [below](#tick_label_options-block).

### `primary_y_axis_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

### `progress_bar` Block

* `foreground_color` - (Required) See [below](#foreground_color-block).
* `visibility` - (Optional)

### `radar_chart_aggregated_field_wells` Block

* `category` - (Optional) See [below](#category-block).
* `color` - (Optional) See [below](#color-block).
* `values` - (Optional) See [below](#values-block).

### `radar_chart_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `range` Block

* `data_driven` - (Optional)
* `max` - (Optional)
* `min` - (Optional)
* `min_max` - (Optional) See [below](#min_max-block).

### `range_ends_label_type` Block

* `visibility` - (Optional)

### `range_maximum` Block

* `parameter` - (Optional)
* `static_value` - (Optional)

### `range_maximum_value` Block

* `parameter` - (Optional)
* `rolling_date` - (Optional) See [below](#rolling_date-block).
* `static_value` - (Optional)

### `range_minimum` Block

* `parameter` - (Optional)
* `static_value` - (Optional)

### `range_minimum_value` Block

* `parameter` - (Optional)
* `rolling_date` - (Optional) See [below](#rolling_date-block).
* `static_value` - (Optional)

### `reference_lines` Block

* `data_configuration` - (Required) See [below](#data_configuration-block).
* `label_configuration` - (Optional) See [below](#label_configuration-block).
* `status` - (Optional)
* `style_configuration` - (Optional) See [below](#style_configuration-block).

### `relative_date_time` Block

* `display_options` - (Optional) See [below](#display_options-block).
* `filter_control_id` - (Required)
* `source_filter_id` - (Required)
* `title` - (Required)

### `relative_dates_filter` Block

* `anchor_date_configuration` - (Required) See [below](#anchor_date_configuration-block).
* `column` - (Required) See [below](#column-block).
* `exclude_period_configuration` - (Optional) See [below](#exclude_period_configuration-block).
* `filter_id` - (Required)
* `minimum_granularity` - (Required)
* `null_option` - (Required)
* `parameter_name` - (Optional)
* `relative_date_type` - (Required)
* `relative_date_value` - (Optional)
* `time_granularity` - (Required)

### `rendering_rules` Block

* `configuration_overrides` - (Required) See [below](#configuration_overrides-block).
* `expression` - (Required)

### `right` Block

* `color` - (Optional)
* `style` - (Optional)
* `thickness` - (Optional)

### `rolling_date` Block

* `data_set_identifier` - (Optional)
* `expression` - (Required)

### `row` Block

* `background_color` - (Required) See [below](#background_color-block).
* `text_color` - (Required) See [below](#text_color-block).

### `row_alternate_color_options` Block

* `row_alternate_colors` - (Optional)
* `status` - (Optional)

### `row_field_names_style` Block

* `background_color` - (Optional)
* `border` - (Optional) See [below](#border-block).
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `height` - (Optional)
* `horizontal_text_alignment` - (Optional)
* `text_wrap` - (Optional)
* `vertical_text_alignment` - (Optional)
* `visibility` - (Optional)

### `row_header_style` Block

* `background_color` - (Optional)
* `border` - (Optional) See [below](#border-block).
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `height` - (Optional)
* `horizontal_text_alignment` - (Optional)
* `text_wrap` - (Optional)
* `vertical_text_alignment` - (Optional)
* `visibility` - (Optional)

### `row_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

### `row_sort` Block

* `column_sort` - (Optional) See [below](#column_sort-block).
* `field_sort` - (Optional) See [below](#field_sort-block).

### `row_subtotal_options` Block

* `custom_label` - (Optional)
* `field_level` - (Optional)
* `field_level_options` - (Optional) See [below](#field_level_options-block).
* `metric_header_cell_style` - (Optional) See [below](#metric_header_cell_style-block).
* `total_cell_style` - (Optional) See [below](#total_cell_style-block).
* `totals_visibility` - (Optional)
* `value_cell_style` - (Optional) See [below](#value_cell_style-block).

### `row_total_options` Block

* `custom_label` - (Optional)
* `metric_header_cell_style` - (Optional) See [below](#metric_header_cell_style-block).
* `placement` - (Optional)
* `scroll_status` - (Optional)
* `total_cell_style` - (Optional) See [below](#total_cell_style-block).
* `totals_visibility` - (Optional)
* `value_cell_style` - (Optional) See [below](#value_cell_style-block).

### `rows` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `same_sheet_target_visual_configuration` Block

* `target_visual_option` - (Optional)
* `target_visuals` - (Optional)

### `sankey_diagram_aggregated_field_wells` Block

* `destination` - (Optional) See [below](#destination-block).
* `source` - (Optional) See [below](#source-block).
* `weight` - (Optional) See [below](#weight-block).

### `sankey_diagram_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `scale` Block

* `linear` - (Optional) See [below](#linear-block).
* `logarithmic` - (Optional) See [below](#logarithmic-block).

### `scatter_plot_categorically_aggregated_field_wells` Block

* `category` - (Optional) See [below](#category-block).
* `size` - (Optional) See [below](#size-block).
* `x_axis` - (Optional) See [below](#x_axis-block).
* `y_axis` - (Optional) See [below](#y_axis-block).

### `scatter_plot_unaggregated_field_wells` Block

* `size` - (Optional) See [below](#size-block).
* `x_axis` - (Optional) See [below](#x_axis-block).
* `y_axis` - (Optional) See [below](#y_axis-block).

### `scatter_plot_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `scenario` Block

* `what_if_point_scenario` - (Optional) See [below](#what_if_point_scenario-block).
* `what_if_range_scenario` - (Optional) See [below](#what_if_range_scenario-block).

### `scope` Block

* `role` - (Optional)

### `scope_configuration` Block

* `selected_sheets` - (Optional) See [below](#selected_sheets-block).

### `screen_canvas_size_options` Block

* `optimized_view_port_width` - (Required)
* `resize_option` - (Required)

### `scrollbar_options` Block

* `visibility` - (Optional)
* `visible_range` - (Optional) See [below](#visible_range-block).

### `search_options` Block

* `visibility` - (Optional)

### `secondary_value` Block

* `visibility` - (Optional)

### `secondary_value_font_configuration` Block

* `font_color` - (Optional)
* `font_decoration` - (Optional)
* `font_size` - (Optional) See [below](#font_size-block).
* `font_style` - (Optional)
* `font_weight` - (Optional) See [below](#font_weight-block).

### `secondary_y_axis_display_options` Block

* `axis_line_visibility` - (Optional)
* `axis_offset` - (Optional)
* `axis_options` - (Optional) See [below](#axis_options-block).
* `data_options` - (Optional) See [below](#data_options-block).
* `grid_line_visibility` - (Optional)
* `missing_data_configuration` - (Optional) See [below](#missing_data_configuration-block).
* `scrollbar_options` - (Optional) See [below](#scrollbar_options-block).
* `tick_label_options` - (Optional) See [below](#tick_label_options-block).

### `secondary_y_axis_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

### `section_based` Block

* `canvas_size_options` - (Required) See [below](#canvas_size_options-block).

### `section_based_layout` Block

* `body_sections` - (Required) See [below](#body_sections-block).
* `canvas_size_options` - (Optional) See [below](#canvas_size_options-block).
* `footer_sections` - (Required) See [below](#footer_sections-block).
* `header_sections` - (Required) See [below](#header_sections-block).

### `select_all_options` Block

* `visibility` - (Optional)

### `selectable_values` Block

* `link_to_data_set_column` - (Optional) See [below](#link_to_data_set_column-block).
* `values` - (Optional)

### `selected_border_style` Block

* `color` - (Optional)
* `visibility` - (Optional)

### `selected_field_options` Block

* `custom_label` - (Optional)
* `field_id` - (Required)
* `url_styling` - (Optional) See [below](#url_styling-block).
* `visibility` - (Optional)
* `width` - (Optional)

### `selected_fields_configuration` Block

* `selected_field_option` - (Optional)
* `selected_fields` - (Optional)

### `selected_sheets` Block

* `sheet_visual_scoping_configurations` - (Optional) See [below](#sheet_visual_scoping_configurations-block).

### `separator_configuration` Block

* `decimal_separator` - (Optional)
* `thousands_separator` - (Optional) See [below](#thousands_separator-block).

### `series` Block

* `data_field_series_item` - (Optional) See [below](#data_field_series_item-block).
* `field_series_item` - (Optional) See [below](#field_series_item-block).

### `set_parameters_operation` Block

* `parameter_value_configurations` - (Required) See [below](#parameter_value_configurations-block).

### `settings` Block

* `line_style_settings` - (Optional) See [below](#line_style_settings-block).
* `marker_style_settings` - (Optional) See [below](#marker_style_settings-block).

### `shape` Block

* `field_id` - (Required)
* `format` - (Optional) See [below](#format-block).

### `sheet_control_layouts` Block

* `configuration` - (Required) See [below](#configuration-block).

### `sheet_visual_scoping_configurations` Block

* `scope` - (Required)
* `sheet_id` - (Required)
* `visual_ids` - (Optional)

### `sheets` Block

* `content_type` - (Optional)
* `description` - (Optional)
* `filter_controls` - (Optional) See [below](#filter_controls-block).
* `layouts` - (Optional) See [below](#layouts-block).
* `name` - (Optional)
* `parameter_controls` - (Optional) See [below](#parameter_controls-block).
* `sheet_control_layouts` - (Optional) See [below](#sheet_control_layouts-block).
* `sheet_id` - (Required)
* `text_boxes` - (Optional) See [below](#text_boxes-block).
* `title` - (Optional)
* `visuals` - (Optional) See [below](#visuals-block).

### `side_specific_border` Block

* `bottom` - (Required) See [below](#bottom-block).
* `inner_horizontal` - (Required) See [below](#inner_horizontal-block).
* `inner_vertical` - (Required) See [below](#inner_vertical-block).
* `left` - (Required) See [below](#left-block).
* `right` - (Required) See [below](#right-block).
* `top` - (Required) See [below](#top-block).

### `simple_cluster_marker` Block

* `color` - (Optional)

### `size` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `size_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

### `sizes` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `sizing_options` Block

* `table_cell_image_scaling_configuration` - (Optional)

### `slider` Block

* `display_options` - (Optional) See [below](#display_options-block).
* `filter_control_id` - (Required)
* `maximum_value` - (Required)
* `minimum_value` - (Required)
* `parameter_control_id` - (Required)
* `source_filter_id` - (Required)
* `source_parameter_name` - (Required)
* `step_size` - (Required)
* `title` - (Required)
* `type` - (Optional)

### `small_multiples` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `small_multiples_limit_configuration` Block

* `items_limit` - (Optional)
* `other_categories` - (Required)

### `small_multiples_options` Block

* `max_visible_columns` - (Optional)
* `max_visible_rows` - (Optional)
* `panel_configuration` - (Optional) See [below](#panel_configuration-block).

### `small_multiples_sort` Block

* `column_sort` - (Optional) See [below](#column_sort-block).
* `field_sort` - (Optional) See [below](#field_sort-block).

### `solid` Block

* `color` - (Optional)
* `expression` - (Required)

### `sort_by` Block

* `column` - (Optional) See [below](#column-block).
* `column_name` - (Required)
* `data_path` - (Optional) See [below](#data_path-block).
* `data_set_identifier` - (Required)
* `field` - (Optional) See [below](#field-block).

### `sort_configuration` Block

* `breakdown_items_limit` - (Optional) See [below](#breakdown_items_limit-block).
* `category_items_limit` - (Optional) See [below](#category_items_limit-block).
* `category_items_limit_configuration` - (Optional) See [below](#category_items_limit_configuration-block).
* `category_sort` - (Optional) See [below](#category_sort-block).
* `color_items_limit` - (Optional) See [below](#color_items_limit-block).
* `color_items_limit_configuration` - (Optional) See [below](#color_items_limit_configuration-block).
* `color_sort` - (Optional) See [below](#color_sort-block).
* `destination_items_limit` - (Optional) See [below](#destination_items_limit-block).
* `field_sort_options` - (Optional) See [below](#field_sort_options-block).
* `heat_map_column_items_limit_configuration` - (Optional) See [below](#heat_map_column_items_limit_configuration-block).
* `heat_map_column_sort` - (Optional) See [below](#heat_map_column_sort-block).
* `heat_map_row_items_limit_configuration` - (Optional) See [below](#heat_map_row_items_limit_configuration-block).
* `heat_map_row_sort` - (Optional) See [below](#heat_map_row_sort-block).
* `pagination_configuration` - (Optional) See [below](#pagination_configuration-block).
* `row_sort` - (Optional) See [below](#row_sort-block).
* `small_multiples_limit_configuration` - (Optional) See [below](#small_multiples_limit_configuration-block).
* `small_multiples_sort` - (Optional) See [below](#small_multiples_sort-block).
* `source_items_limit` - (Optional) See [below](#source_items_limit-block).
* `tree_map_group_items_limit_configuration` - (Optional) See [below](#tree_map_group_items_limit_configuration-block).
* `tree_map_sort` - (Optional) See [below](#tree_map_sort-block).
* `trend_group_sort` - (Optional) See [below](#trend_group_sort-block).
* `weight_sort` - (Optional) See [below](#weight_sort-block).

### `sort_paths` Block

* `field_id` - (Required)
* `field_value` - (Required)

### `source` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `source_controls` Block

* `column_to_match` - (Required) See [below](#column_to_match-block).
* `source_sheet_control_id` - (Optional)

### `source_items_limit` Block

* `items_limit` - (Optional)
* `other_categories` - (Required)

### `sparkline` Block

* `color` - (Optional)
* `tooltip_visibility` - (Optional)
* `type` - (Required)
* `visibility` - (Optional)

### `standard_layout` Block

* `type` - (Required)

### `static_configuration` Block

* `value` - (Required)

### `stops` Block

* `color` - (Optional)
* `data_value` - (Optional)
* `gradient_offset` - (Required)

### `string_format_configuration` Block

* `null_value_format_configuration` - (Optional) See [below](#null_value_format_configuration-block).
* `numeric_format_configuration` - (Optional) See [below](#numeric_format_configuration-block).

### `string_parameter_declaration` Block

* `default_values` - (Optional) See [below](#default_values-block).
* `name` - (Required)
* `parameter_value_type` - (Required)
* `values_when_unset` - (Optional) See [below](#values_when_unset-block).

### `style` Block

* `height` - (Optional)
* `padding` - (Optional) See [below](#padding-block).

### `style_configuration` Block

* `color` - (Optional)
* `pattern` - (Optional)

### `style_options` Block

* `fill_style` - (Optional)

### `subtitle` Block

* `format_text` - (Optional) See [below](#format_text-block).
* `visibility` - (Optional)

### `table_aggregated_field_wells` Block

* `group_by` - (Optional) See [below](#group_by-block).
* `values` - (Optional) See [below](#values-block).

### `table_inline_visualizations` Block

* `data_bars` - (Optional) See [below](#data_bars-block).

### `table_options` Block

* `cell_style` - (Optional) See [below](#cell_style-block).
* `collapsed_row_dimensions_visibility` - (Optional)
* `column_header_style` - (Optional) See [below](#column_header_style-block).
* `column_names_visibility` - (Optional)
* `header_style` - (Optional) See [below](#header_style-block).
* `metric_placement` - (Optional)
* `orientation` - (Optional)
* `row_alternate_color_options` - (Optional) See [below](#row_alternate_color_options-block).
* `row_field_names_style` - (Optional) See [below](#row_field_names_style-block).
* `row_header_style` - (Optional) See [below](#row_header_style-block).
* `single_metric_visibility` - (Optional)
* `toggle_buttons_visibility` - (Optional)

### `table_unaggregated_field_wells` Block

* `values` - (Optional) See [below](#values-block).

### `table_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `conditional_formatting` - (Optional) See [below](#conditional_formatting-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `target_value` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `target_values` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `target_visuals_configuration` Block

* `same_sheet_target_visual_configuration` - (Optional) See [below](#same_sheet_target_visual_configuration-block).

### `text_area` Block

* `delimiter` - (Optional)
* `display_options` - (Optional) See [below](#display_options-block).
* `filter_control_id` - (Required)
* `parameter_control_id` - (Required)
* `source_filter_id` - (Required)
* `source_parameter_name` - (Required)
* `title` - (Required)

### `text_boxes` Block

* `content` - (Optional)
* `sheet_text_box_id` - (Required)

### `text_color` Block

* `gradient` - (Optional) See [below](#gradient-block).
* `solid` - (Optional) See [below](#solid-block).

### `text_field` Block

* `display_options` - (Optional) See [below](#display_options-block).
* `filter_control_id` - (Required)
* `parameter_control_id` - (Required)
* `source_filter_id` - (Required)
* `source_parameter_name` - (Required)
* `title` - (Required)

### `text_format` Block

* `background_color` - (Required) See [below](#background_color-block).
* `icon` - (Optional) See [below](#icon-block).
* `text_color` - (Required) See [below](#text_color-block).

### `thousands_separator` Block

* `symbol` - (Optional)
* `visibility` - (Optional)

### `tick_label_options` Block

* `label_options` - (Optional) See [below](#label_options-block).
* `rotation_angle` - (Optional)

### `time` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `time_equality_filter` Block

* `column` - (Required) See [below](#column-block).
* `filter_id` - (Required)
* `parameter_name` - (Optional)
* `time_granularity` - (Required)
* `value` - (Optional)

### `time_range_filter` Block

* `column` - (Required) See [below](#column-block).
* `exclude_period_configuration` - (Optional) See [below](#exclude_period_configuration-block).
* `filter_id` - (Required)
* `include_maximum` - (Optional)
* `include_minimum` - (Optional)
* `null_option` - (Required)
* `range_maximum` - (Required)
* `range_maximum_value` - (Optional) See [below](#range_maximum_value-block).
* `range_minimum` - (Required)
* `range_minimum_value` - (Optional) See [below](#range_minimum_value-block).
* `time_granularity` - (Required)

### `title` Block

* `custom_label` - (Optional)
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `format_text` - (Optional) See [below](#format_text-block).
* `horizontal_text_alignment` - (Optional)
* `visibility` - (Optional)

### `title_options` Block

* `custom_label` - (Optional)
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `visibility` - (Optional)

### `tooltip` Block

* `field_base_tooltip` - (Optional) See [below](#field_base_tooltip-block).
* `selected_tooltip_type` - (Optional)
* `tooltip_visibility` - (Optional)

### `tooltip_fields` Block

* `column_tooltip_item` - (Optional) See [below](#column_tooltip_item-block).
* `field_tooltip_item` - (Optional) See [below](#field_tooltip_item-block).

### `top` Block

* `color` - (Optional)
* `style` - (Optional)
* `thickness` - (Optional)

### `top_bottom_filter` Block

* `aggregation_sort_configuration` - (Required) See [below](#aggregation_sort_configuration-block).
* `column` - (Required) See [below](#column-block).
* `filter_id` - (Required)
* `limit` - (Optional)
* `parameter_name` - (Optional)
* `time_granularity` - (Required)

### `top_bottom_movers` Block

* `category` - (Optional) See [below](#category-block).
* `computation_id` - (Required)
* `mover_size` - (Optional)
* `name` - (Optional)
* `sort_order` - (Required)
* `time` - (Optional) See [below](#time-block).
* `type` - (Required)
* `value` - (Optional) See [below](#value-block).

### `top_bottom_ranked` Block

* `category` - (Optional) See [below](#category-block).
* `computation_id` - (Required)
* `name` - (Optional)
* `result_size` - (Optional)
* `type` - (Required)
* `value` - (Optional) See [below](#value-block).

### `total_aggregation` Block

* `computation_id` - (Required)
* `name` - (Optional)
* `value` - (Optional) See [below](#value-block).

### `total_cell_style` Block

* `background_color` - (Optional)
* `border` - (Optional) See [below](#border-block).
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `height` - (Optional)
* `horizontal_text_alignment` - (Optional)
* `text_wrap` - (Optional)
* `vertical_text_alignment` - (Optional)
* `visibility` - (Optional)

### `total_options` Block

* `column_subtotal_options` - (Optional) See [below](#column_subtotal_options-block).
* `column_total_options` - (Optional) See [below](#column_total_options-block).
* `custom_label` - (Optional)
* `placement` - (Optional)
* `row_subtotal_options` - (Optional) See [below](#row_subtotal_options-block).
* `row_total_options` - (Optional) See [below](#row_total_options-block).
* `scroll_status` - (Optional)
* `total_cell_style` - (Optional) See [below](#total_cell_style-block).
* `totals_visibility` - (Optional)

### `tree_map_aggregated_field_wells` Block

* `colors` - (Optional) See [below](#colors-block).
* `groups` - (Optional) See [below](#groups-block).
* `sizes` - (Optional) See [below](#sizes-block).

### `tree_map_group_items_limit_configuration` Block

* `items_limit` - (Optional)
* `other_categories` - (Required)

### `tree_map_sort` Block

* `column_sort` - (Optional) See [below](#column_sort-block).
* `field_sort` - (Optional) See [below](#field_sort-block).

### `tree_map_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `trend_arrows` Block

* `visibility` - (Optional)

### `trend_group_sort` Block

* `column_sort` - (Optional) See [below](#column_sort-block).
* `field_sort` - (Optional) See [below](#field_sort-block).

### `trend_groups` Block

* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).

### `uniform_border` Block

* `color` - (Optional)
* `style` - (Optional)
* `thickness` - (Optional)

### `unique_values` Block

* `category` - (Optional) See [below](#category-block).
* `computation_id` - (Required)
* `name` - (Optional)

### `url_operation` Block

* `url_target` - (Required)
* `url_template` - (Required)

### `url_styling` Block

* `image_configuration` - (Optional) See [below](#image_configuration-block).
* `link_configuration` - (Optional) See [below](#link_configuration-block).

### `user_name_column` Block

* `column_name` - (Required)
* `data_set_identifier` - (Required)

### `value` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `custom_values_configuration` - (Optional) See [below](#custom_values_configuration-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).
* `select_all_value_options` - (Optional)
* `source_field` - (Optional)
* `source_parameter_name` - (Optional)

### `value_axis` Block

* `axis_line_visibility` - (Optional)
* `axis_offset` - (Optional)
* `data_options` - (Optional) See [below](#data_options-block).
* `grid_line_visibility` - (Optional)
* `scrollbar_options` - (Optional) See [below](#scrollbar_options-block).
* `tick_label_options` - (Optional) See [below](#tick_label_options-block).

### `value_cell_style` Block

* `background_color` - (Optional)
* `border` - (Optional) See [below](#border-block).
* `font_configuration` - (Optional) See [below](#font_configuration-block).
* `height` - (Optional)
* `horizontal_text_alignment` - (Optional)
* `text_wrap` - (Optional)
* `vertical_text_alignment` - (Optional)
* `visibility` - (Optional)

### `value_label_configuration` Block

* `format_configuration` - (Optional) See [below](#format_configuration-block).
* `relative_position` - (Optional)

### `value_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

### `values` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `column` - (Required) See [below](#column-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `field_id` - (Required)
* `format_configuration` - (Optional) See [below](#format_configuration-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `values_when_unset` Block

* `custom_value` - (Optional)
* `value_when_unset_option` - (Optional)

### `visible_range` Block

* `percent_range` - (Optional) See [below](#percent_range-block).

### `visual_layout_options` Block

* `standard_layout` - (Optional) See [below](#standard_layout-block).

### `visual_palette` Block

* `chart_color` - (Optional)
* `color_map` - (Optional) See [below](#color_map-block).

### `visuals` Block

* `bar_chart_visual` - (Optional) See [below](#bar_chart_visual-block).
* `box_plot_visual` - (Optional) See [below](#box_plot_visual-block).
* `combo_chart_visual` - (Optional) See [below](#combo_chart_visual-block).
* `custom_content_visual` - (Optional) See [below](#custom_content_visual-block).
* `empty_visual` - (Optional) See [below](#empty_visual-block).
* `filled_map_visual` - (Optional) See [below](#filled_map_visual-block).
* `funnel_chart_visual` - (Optional) See [below](#funnel_chart_visual-block).
* `gauge_chart_visual` - (Optional) See [below](#gauge_chart_visual-block).
* `geospatial_map_visual` - (Optional) See [below](#geospatial_map_visual-block).
* `heat_map_visual` - (Optional) See [below](#heat_map_visual-block).
* `histogram_visual` - (Optional) See [below](#histogram_visual-block).
* `insight_visual` - (Optional) See [below](#insight_visual-block).
* `kpi_visual` - (Optional) See [below](#kpi_visual-block).
* `line_chart_visual` - (Optional) See [below](#line_chart_visual-block).
* `pie_chart_visual` - (Optional) See [below](#pie_chart_visual-block).
* `pivot_table_visual` - (Optional) See [below](#pivot_table_visual-block).
* `radar_chart_visual` - (Optional) See [below](#radar_chart_visual-block).
* `sankey_diagram_visual` - (Optional) See [below](#sankey_diagram_visual-block).
* `scatter_plot_visual` - (Optional) See [below](#scatter_plot_visual-block).
* `table_visual` - (Optional) See [below](#table_visual-block).
* `tree_map_visual` - (Optional) See [below](#tree_map_visual-block).
* `waterfall_visual` - (Optional) See [below](#waterfall_visual-block).
* `word_cloud_visual` - (Optional) See [below](#word_cloud_visual-block).

### `waterfall_chart_aggregated_field_wells` Block

* `breakdowns` - (Optional) See [below](#breakdowns-block).
* `categories` - (Optional) See [below](#categories-block).
* `values` - (Optional) See [below](#values-block).

### `waterfall_chart_options` Block

* `total_bar_label` - (Optional)

### `waterfall_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `weight` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `weight_sort` Block

* `column_sort` - (Optional) See [below](#column_sort-block).
* `field_sort` - (Optional) See [below](#field_sort-block).

### `what_if_point_scenario` Block

* `date` - (Required)
* `value` - (Required)

### `what_if_range_scenario` Block

* `end_date` - (Required)
* `start_date` - (Required)
* `value` - (Required)

### `window_options` Block

* `bounds` - (Optional) See [below](#bounds-block).
* `map_zoom_mode` - (Optional)

### `word_cloud_aggregated_field_wells` Block

* `group_by` - (Optional) See [below](#group_by-block).
* `size` - (Optional) See [below](#size-block).

### `word_cloud_options` Block

* `cloud_layout` - (Optional)
* `maximum_string_length` - (Optional)
* `word_casing` - (Optional)
* `word_orientation` - (Optional)
* `word_padding` - (Optional)
* `word_scaling` - (Optional)

### `word_cloud_visual` Block

* `actions` - (Optional) See [below](#actions-block).
* `chart_configuration` - (Optional) See [below](#chart_configuration-block).
* `column_hierarchies` - (Optional) See [below](#column_hierarchies-block).
* `subtitle` - (Optional) See [below](#subtitle-block).
* `title` - (Optional) See [below](#title-block).
* `visual_id` - (Required)

### `x_axis` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `x_axis_display_options` Block

* `axis_line_visibility` - (Optional)
* `axis_offset` - (Optional)
* `data_options` - (Optional) See [below](#data_options-block).
* `grid_line_visibility` - (Optional)
* `scrollbar_options` - (Optional) See [below](#scrollbar_options-block).
* `tick_label_options` - (Optional) See [below](#tick_label_options-block).

### `x_axis_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

### `y_axis` Block

* `calculated_measure_field` - (Optional) See [below](#calculated_measure_field-block).
* `categorical_dimension_field` - (Optional) See [below](#categorical_dimension_field-block).
* `categorical_measure_field` - (Optional) See [below](#categorical_measure_field-block).
* `date_dimension_field` - (Optional) See [below](#date_dimension_field-block).
* `date_measure_field` - (Optional) See [below](#date_measure_field-block).
* `numerical_dimension_field` - (Optional) See [below](#numerical_dimension_field-block).
* `numerical_measure_field` - (Optional) See [below](#numerical_measure_field-block).

### `y_axis_display_options` Block

* `axis_line_visibility` - (Optional)
* `axis_offset` - (Optional)
* `data_options` - (Optional) See [below](#data_options-block).
* `grid_line_visibility` - (Optional)
* `scrollbar_options` - (Optional) See [below](#scrollbar_options-block).
* `tick_label_options` - (Optional) See [below](#tick_label_options-block).

### `y_axis_label_options` Block

* `axis_label_options` - (Optional) See [below](#axis_label_options-block).
* `sort_icon_visibility` - (Optional)
* `visibility` - (Optional)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the template.
* `created_time` - Time that the template was created.
* `id` - Comma-delimited string joining AWS account ID and template ID.
* `last_updated_time` - Time that the template was last updated.
* `source_entity_arn` - ARN of an analysis or template used to create this template.
* `status` - Template creation status.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `version_number` - Version number of the template version.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a QuickSight Template using the AWS account ID and template ID separated by a comma (`,`). For example:

```terraform
import {
  to = aws_quicksight_template.example
  id = "123456789012,example-id"
}
```

Using `terraform import`, import a QuickSight Template using the AWS account ID and template ID separated by a comma (`,`). For example:

```console
% terraform import aws_quicksight_template.example 123456789012,example-id
```
