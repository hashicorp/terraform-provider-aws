---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_analysis"
description: |-
  Manages a QuickSight Analysis.
---

# Resource: aws_quicksight_analysis

Resource for managing a QuickSight Analysis.

## Example Usage

### From Source Template

```terraform
resource "aws_quicksight_analysis" "example" {
  analysis_id = "example-id"
  name        = "example-name"
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
resource "aws_quicksight_analysis" "example" {
  analysis_id = "example-id"
  name        = "example-name"
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

* `analysis_id` - (Required, Forces new resource) Identifier for the analysis.
* `name` - (Required) Display name for the analysis.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `definition` - (Optional) Analysis definition. Only one of `definition` or `source_entity` should be configured. See [`definition` Block](#definition-block) below.
* `parameters` - (Optional) Parameters for the creation of the analysis, which you want to use to override the default settings. See [`parameters` Block](#parameters-block) below.
* `permissions` - (Optional) Set of resource permissions on the analysis. Maximum of 64 items. See [`permissions` Block](#permissions-block) below.
* `recovery_window_in_days` - (Optional) Number of days that Amazon QuickSight waits before it deletes the analysis. Use `0` to force deletion without recovery. Minimum value of `7`. Maximum value of `30`. Defaults to `30`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `source_entity` - (Optional) Entity that you are using as a source when you create the analysis (template). Only one of `definition` or `source_entity` should be configured. See [`source_entity` Block](#source_entity-block) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `theme_arn` - (Optional) ARN of the theme that is being used for this analysis. The theme ARN must exist in the same AWS account where you create the analysis.

### `action_operations` Block

* `filter_operation` - (Optional) Filter operation. See [`filter_operation` Block](#filter_operation-block) below.
* `navigation_operation` - (Optional) Navigation operation. See [`navigation_operation` Block](#navigation_operation-block) below.
* `set_parameters_operation` - (Optional) Set parameters operation. See [`set_parameters_operation` Block](#set_parameters_operation-block) below.
* `url_operation` - (Optional) URL operation. See [`url_operation` Block](#url_operation-block) below.

### `actions` Block

* `action_operations` - (Required) Action operations. See [`action_operations` Block](#action_operations-block) below.
* `custom_action_id` - (Required) Custom action ID.
* `name` - (Required) Name.
* `status` - (Required) Status.
* `trigger` - (Required) Trigger.

### `actual_value` Block

* `icon` - (Optional) Icon. See [`icon` Block](#icon-block) below.
* `text_color` - (Required) Text color. See [`text_color` Block](#text_color-block) below.

### `after` Block

* `status` - (Optional) Status.

### `aggregation` Block

* `categorical_aggregation_function` - (Optional) Categorical aggregation function.
* `date_aggregation_function` - (Optional) Date aggregation function.
* `numerical_aggregation_function` - (Optional) Numerical aggregation function. See [`numerical_aggregation_function` Block](#numerical_aggregation_function-block) below.

### `aggregation_function` Block

* `categorical_aggregation_function` - (Optional) Categorical aggregation function.
* `date_aggregation_function` - (Optional) Date aggregation function.
* `numerical_aggregation_function` - (Optional) Numerical aggregation function. See [`numerical_aggregation_function` Block](#numerical_aggregation_function-block) below.
* `percentile_aggregation` - (Optional) Percentile aggregation. See [`percentile_aggregation` Block](#percentile_aggregation-block) below.
* `simple_numerical_aggregation` - (Optional) Simple numerical aggregation.

### `aggregation_sort_configuration` Block

* `aggregation_function` - (Required) Aggregation function. See [`aggregation_function` Block](#aggregation_function-block) below.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `sort_direction` - (Required) Sort direction.

### `analysis_defaults` Block

* `default_new_sheet_configuration` - (Required) Default new sheet configuration. See [`default_new_sheet_configuration` Block](#default_new_sheet_configuration-block) below.

### `anchor_date_configuration` Block

* `anchor_option` - (Optional) Anchor option.
* `parameter_name` - (Optional) Parameter name.

### `apply_to` Block

* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `field_id` - (Required) Field ID.

### `arc` Block

* `arc_angle` - (Optional) Arc angle.
* `arc_thickness` - (Optional) Arc thickness.
* `foreground_color` - (Required) Foreground color. See [`foreground_color` Block](#foreground_color-block) below.

### `arc_axis` Block

* `range` - (Optional) Range. See [`range` Block](#range-block) below.
* `reserve_range` - (Optional) Reserve range.

### `arc_options` Block

* `arc_thickness` - (Optional) Arc thickness.

### `area_style_settings` Block

* `visibility` - (Optional) Visibility.

### `axis_label_options` Block

* `apply_to` - (Optional) Apply to. See [`apply_to` Block](#apply_to-block) below.
* `custom_label` - (Optional) Custom label.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.

### `axis_options` Block

* `axis_line_visibility` - (Optional) Axis line visibility.
* `axis_offset` - (Optional) Axis offset.
* `data_options` - (Optional) Data options. See [`data_options` Block](#data_options-block) below.
* `grid_line_visibility` - (Optional) Grid line visibility.
* `scrollbar_options` - (Optional) Scrollbar options. See [`scrollbar_options` Block](#scrollbar_options-block) below.
* `tick_label_options` - (Optional) Tick label options. See [`tick_label_options` Block](#tick_label_options-block) below.

### `background_color` Block

* `gradient` - (Optional) Gradient. See [`gradient` Block](#gradient-block) below.
* `solid` - (Optional) Solid. See [`solid` Block](#solid-block) below.

### `background_style` Block

* `color` - (Optional) Color.
* `visibility` - (Optional) Visibility.

### `bar_chart_aggregated_field_wells` Block

* `category` - (Optional) Category. See [`category` Block](#category-block) below.
* `colors` - (Optional) Colors. See [`colors` Block](#colors-block) below.
* `small_multiples` - (Optional) Small multiples. See [`small_multiples` Block](#small_multiples-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `bar_chart_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `bar_data_labels` Block

* `category_label_visibility` - (Optional) Category label visibility.
* `data_label_types` - (Optional) Data label types. See [`data_label_types` Block](#data_label_types-block) below.
* `label_color` - (Optional) Label color.
* `label_content` - (Optional) Label content.
* `label_font_configuration` - (Optional) Label font configuration. See [`label_font_configuration` Block](#label_font_configuration-block) below.
* `measure_label_visibility` - (Optional) Measure label visibility.
* `overlap` - (Optional) Overlap.
* `position` - (Optional) Position.
* `visibility` - (Optional) Visibility.

### `bar_values` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `base_series_settings` Block

* `area_style_settings` - (Optional) Area style settings. See [`area_style_settings` Block](#area_style_settings-block) below.

### `bin_count` Block

* `value` - (Optional) Value.

### `bin_options` Block

* `bin_count` - (Optional) Bin count. See [`bin_count` Block](#bin_count-block) below.
* `bin_width` - (Optional) Bin width. See [`bin_width` Block](#bin_width-block) below.
* `selected_bin_type` - (Optional) Selected bin type.
* `start_value` - (Optional) Start value.

### `bin_width` Block

* `bin_count_limit` - (Optional) Bin count limit.
* `value` - (Optional) Value.

### `body_sections` Block

* `content` - (Required) Content. See [`content` Block](#content-block) below.
* `page_break_configuration` - (Optional) Page break configuration. See [`page_break_configuration` Block](#page_break_configuration-block) below.
* `section_id` - (Required) Section ID.
* `style` - (Optional) Style. See [`style` Block](#style-block) below.

### `border` Block

* `side_specific_border` - (Optional) Side specific border. See [`side_specific_border` Block](#side_specific_border-block) below.
* `uniform_border` - (Required) Uniform border. See [`uniform_border` Block](#uniform_border-block) below.

### `border_style` Block

* `color` - (Optional) Color.
* `visibility` - (Optional) Visibility.

### `bottom` Block

* `color` - (Optional) Color.
* `style` - (Optional) Style.
* `thickness` - (Optional) Thickness.

### `bounds` Block

* `east` - (Required) East.
* `north` - (Required) North.
* `south` - (Required) South.
* `west` - (Required) West.

### `box_plot_aggregated_field_wells` Block

* `group_by` - (Optional) Group by. See [`group_by` Block](#group_by-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `box_plot_options` Block

* `all_data_points_visibility` - (Optional) All data points visibility.
* `outlier_visibility` - (Optional) Outlier visibility.
* `style_options` - (Optional) Style options. See [`style_options` Block](#style_options-block) below.

### `box_plot_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `breakdown_items_limit` Block

* `items_limit` - (Optional) Items limit.
* `other_categories` - (Required) Other categories.

### `breakdowns` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `calculated_fields` Block

* `data_set_identifier` - (Required) Data set identifier.
* `expression` - (Required) Expression.
* `name` - (Required) Name.

### `calculated_measure_field` Block

* `expression` - (Required) Expression.
* `field_id` - (Required) Field ID.

### `calculation` Block

* `percentile_aggregation` - (Optional) Percentile aggregation. See [`percentile_aggregation` Block](#percentile_aggregation-block) below.
* `simple_numerical_aggregation` - (Optional) Simple numerical aggregation.

### `canvas_size_options` Block

* `paper_canvas_size_options` - (Optional) Paper canvas size options. See [`paper_canvas_size_options` Block](#paper_canvas_size_options-block) below.
* `screen_canvas_size_options` - (Optional) Screen canvas size options. See [`screen_canvas_size_options` Block](#screen_canvas_size_options-block) below.

### `cascading_control_configuration` Block

* `source_controls` - (Optional) Source controls. See [`source_controls` Block](#source_controls-block) below.

### `categorical_dimension_field` Block

* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `field_id` - (Required) Field ID.
* `format_configuration` - (Optional) Format configuration. See [`format_configuration` Block](#format_configuration-block) below.
* `hierarchy_id` - (Optional) Hierarchy ID.

### `categorical_measure_field` Block

* `aggregation_function` - (Optional) Aggregation function.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `field_id` - (Required) Field ID.
* `format_configuration` - (Optional) Format configuration. See [`format_configuration` Block](#format_configuration-block) below.

### `categories` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `category` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `category_axis` Block

* `axis_line_visibility` - (Optional) Axis line visibility.
* `axis_offset` - (Optional) Axis offset.
* `data_options` - (Optional) Data options. See [`data_options` Block](#data_options-block) below.
* `grid_line_visibility` - (Optional) Grid line visibility.
* `scrollbar_options` - (Optional) Scrollbar options. See [`scrollbar_options` Block](#scrollbar_options-block) below.
* `tick_label_options` - (Optional) Tick label options. See [`tick_label_options` Block](#tick_label_options-block) below.

### `category_axis_display_options` Block

* `axis_line_visibility` - (Optional) Axis line visibility.
* `axis_offset` - (Optional) Axis offset.
* `data_options` - (Optional) Data options. See [`data_options` Block](#data_options-block) below.
* `grid_line_visibility` - (Optional) Grid line visibility.
* `scrollbar_options` - (Optional) Scrollbar options. See [`scrollbar_options` Block](#scrollbar_options-block) below.
* `tick_label_options` - (Optional) Tick label options. See [`tick_label_options` Block](#tick_label_options-block) below.

### `category_axis_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

### `category_filter` Block

* `category_values` - (Required) Category values.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `configuration` - (Required) Configuration. See [`configuration` Block](#configuration-block) below.
* `filter_id` - (Required) Filter ID.

### `category_items_limit` Block

* `items_limit` - (Optional) Items limit.
* `other_categories` - (Required) Other categories.

### `category_items_limit_configuration` Block

* `items_limit` - (Optional) Items limit.
* `other_categories` - (Required) Other categories.

### `category_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

### `category_sort` Block

* `column_sort` - (Optional) Column sort. See [`column_sort` Block](#column_sort-block) below.
* `field_sort` - (Optional) Field sort. See [`field_sort` Block](#field_sort-block) below.

### `cell` Block

* `field_id` - (Required) Field ID.
* `scope` - (Optional) Scope. See [`scope` Block](#scope-block) below.
* `text_format` - (Optional) Text format. See [`text_format` Block](#text_format-block) below.

### `cell_style` Block

* `background_color` - (Optional) Background color.
* `border` - (Optional) Border. See [`border` Block](#border-block) below.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `height` - (Optional) Height.
* `horizontal_text_alignment` - (Optional) Horizontal text alignment.
* `text_wrap` - (Optional) Text wrap.
* `vertical_text_alignment` - (Optional) Vertical text alignment.
* `visibility` - (Optional) Visibility.

### `chart_configuration` Block

* `alternate_band_colors_visibility` - (Optional) Alternate band colors visibility.
* `alternate_band_even_color` - (Optional) Alternate band even color.
* `alternate_band_odd_color` - (Optional) Alternate band odd color.
* `bar_data_labels` - (Optional) Bar data labels. See [`bar_data_labels` Block](#bar_data_labels-block) below.
* `bars_arrangement` - (Optional) Bars arrangement.
* `base_series_settings` - (Optional) Base series settings. See [`base_series_settings` Block](#base_series_settings-block) below.
* `bin_options` - (Optional) Bin options. See [`bin_options` Block](#bin_options-block) below.
* `box_plot_options` - (Optional) Box plot options. See [`box_plot_options` Block](#box_plot_options-block) below.
* `category_axis` - (Optional) Category axis. See [`category_axis` Block](#category_axis-block) below.
* `category_axis_display_options` - (Optional) Category axis display options. See [`category_axis_display_options` Block](#category_axis_display_options-block) below.
* `category_axis_label_options` - (Optional) Category axis label options. See [`category_axis_label_options` Block](#category_axis_label_options-block) below.
* `category_label_options` - (Optional) Category label options. See [`category_label_options` Block](#category_label_options-block) below.
* `color_axis` - (Optional) Color axis. See [`color_axis` Block](#color_axis-block) below.
* `color_label_options` - (Optional) Color label options. See [`color_label_options` Block](#color_label_options-block) below.
* `color_scale` - (Optional) Color scale. See [`color_scale` Block](#color_scale-block) below.
* `column_label_options` - (Optional) Column label options. See [`column_label_options` Block](#column_label_options-block) below.
* `content_type` - (Optional) Content type.
* `content_url` - (Optional) Content URL.
* `contribution_analysis_defaults` - (Optional) Contribution analysis defaults. See [`contribution_analysis_defaults` Block](#contribution_analysis_defaults-block) below.
* `data_label_options` - (Optional) Data label options. See [`data_label_options` Block](#data_label_options-block) below.
* `data_labels` - (Optional) Data labels. See [`data_labels` Block](#data_labels-block) below.
* `default_series_settings` - (Optional) Default series settings. See [`default_series_settings` Block](#default_series_settings-block) below.
* `donut_options` - (Optional) Donut options. See [`donut_options` Block](#donut_options-block) below.
* `field_options` - (Optional) Field options. See [`field_options` Block](#field_options-block) below.
* `field_wells` - (Optional) Field wells. See [`field_wells` Block](#field_wells-block) below.
* `forecast_configurations` - (Optional) Forecast configurations. See [`forecast_configurations` Block](#forecast_configurations-block) below.
* `gauge_chart_options` - (Optional) Gauge chart options. See [`gauge_chart_options` Block](#gauge_chart_options-block) below.
* `group_label_options` - (Optional) Group label options. See [`group_label_options` Block](#group_label_options-block) below.
* `image_scaling` - (Optional) Image scaling.
* `kpi_options` - (Optional) Kpi options. See [`kpi_options` Block](#kpi_options-block) below.
* `legend` - (Optional) Legend. See [`legend` Block](#legend-block) below.
* `line_data_labels` - (Optional) Line data labels. See [`line_data_labels` Block](#line_data_labels-block) below.
* `map_style_options` - (Optional) Map style options. See [`map_style_options` Block](#map_style_options-block) below.
* `orientation` - (Optional) Orientation.
* `paginated_report_options` - (Optional) Paginated report options. See [`paginated_report_options` Block](#paginated_report_options-block) below.
* `point_style_options` - (Optional) Point style options. See [`point_style_options` Block](#point_style_options-block) below.
* `primary_y_axis_display_options` - (Optional) Primary y axis display options. See [`primary_y_axis_display_options` Block](#primary_y_axis_display_options-block) below.
* `primary_y_axis_label_options` - (Optional) Primary y axis label options. See [`primary_y_axis_label_options` Block](#primary_y_axis_label_options-block) below.
* `reference_lines` - (Optional) Reference lines. See [`reference_lines` Block](#reference_lines-block) below.
* `row_label_options` - (Optional) Row label options. See [`row_label_options` Block](#row_label_options-block) below.
* `secondary_y_axis_display_options` - (Optional) Secondary y axis display options. See [`secondary_y_axis_display_options` Block](#secondary_y_axis_display_options-block) below.
* `secondary_y_axis_label_options` - (Optional) Secondary y axis label options. See [`secondary_y_axis_label_options` Block](#secondary_y_axis_label_options-block) below.
* `series` - (Optional) Series. See [`series` Block](#series-block) below.
* `shape` - (Optional) Shape.
* `size_label_options` - (Optional) Size label options. See [`size_label_options` Block](#size_label_options-block) below.
* `small_multiples_options` - (Optional) Small multiples options. See [`small_multiples_options` Block](#small_multiples_options-block) below.
* `sort_configuration` - (Optional) Sort configuration. See [`sort_configuration` Block](#sort_configuration-block) below.
* `start_angle` - (Optional) Start angle.
* `table_inline_visualizations` - (Optional) Table inline visualizations. See [`table_inline_visualizations` Block](#table_inline_visualizations-block) below.
* `table_options` - (Optional) Table options. See [`table_options` Block](#table_options-block) below.
* `tooltip` - (Optional) Tooltip. See [`tooltip` Block](#tooltip-block) below.
* `total_options` - (Optional) Total options. See [`total_options` Block](#total_options-block) below.
* `type` - (Optional) Type.
* `value_axis` - (Optional) Value axis. See [`value_axis` Block](#value_axis-block) below.
* `value_label_options` - (Optional) Value label options. See [`value_label_options` Block](#value_label_options-block) below.
* `visual_palette` - (Optional) Visual palette. See [`visual_palette` Block](#visual_palette-block) below.
* `waterfall_chart_options` - (Optional) Waterfall chart options. See [`waterfall_chart_options` Block](#waterfall_chart_options-block) below.
* `window_options` - (Optional) Window options. See [`window_options` Block](#window_options-block) below.
* `word_cloud_options` - (Optional) Word cloud options. See [`word_cloud_options` Block](#word_cloud_options-block) below.
* `x_axis_display_options` - (Optional) X axis display options. See [`x_axis_display_options` Block](#x_axis_display_options-block) below.
* `x_axis_label_options` - (Optional) X axis label options. See [`x_axis_label_options` Block](#x_axis_label_options-block) below.
* `y_axis_display_options` - (Optional) Y axis display options. See [`y_axis_display_options` Block](#y_axis_display_options-block) below.
* `y_axis_label_options` - (Optional) Y axis label options. See [`y_axis_label_options` Block](#y_axis_label_options-block) below.

### `cluster_marker` Block

* `simple_cluster_marker` - (Optional) Simple cluster marker. See [`simple_cluster_marker` Block](#simple_cluster_marker-block) below.

### `cluster_marker_configuration` Block

* `cluster_marker` - (Optional) Cluster marker. See [`cluster_marker` Block](#cluster_marker-block) below.

### `color` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.
* `stops` - (Optional) Stops. See [`stops` Block](#stops-block) below.

### `color_axis` Block

* `axis_line_visibility` - (Optional) Axis line visibility.
* `axis_offset` - (Optional) Axis offset.
* `data_options` - (Optional) Data options. See [`data_options` Block](#data_options-block) below.
* `grid_line_visibility` - (Optional) Grid line visibility.
* `scrollbar_options` - (Optional) Scrollbar options. See [`scrollbar_options` Block](#scrollbar_options-block) below.
* `tick_label_options` - (Optional) Tick label options. See [`tick_label_options` Block](#tick_label_options-block) below.

### `color_items_limit` Block

* `items_limit` - (Optional) Items limit.
* `other_categories` - (Required) Other categories.

### `color_items_limit_configuration` Block

* `items_limit` - (Optional) Items limit.
* `other_categories` - (Required) Other categories.

### `color_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

### `color_map` Block

* `color` - (Required) Color.
* `element` - (Required) Element. See [`element` Block](#element-block) below.
* `time_granularity` - (Optional) Time granularity.

### `color_scale` Block

* `color_fill_type` - (Required) Color fill type.
* `colors` - (Required) Colors. See [`colors` Block](#colors-block) below.
* `null_value_color` - (Optional) Null value color. See [`null_value_color` Block](#null_value_color-block) below.

### `color_sort` Block

* `column_sort` - (Optional) Column sort. See [`column_sort` Block](#column_sort-block) below.
* `field_sort` - (Optional) Field sort. See [`field_sort` Block](#field_sort-block) below.

### `colors` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `color` - (Optional) Color.
* `data_value` - (Optional) Data value.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `column` Block

* `aggregation_function` - (Optional) Aggregation function. See [`aggregation_function` Block](#aggregation_function-block) below.
* `column_name` - (Required) Column name.
* `data_set_identifier` - (Required) Data set identifier.
* `direction` - (Required) Direction.
* `sort_by` - (Required) Sort by. See [`sort_by` Block](#sort_by-block) below.

### `column_configurations` Block

* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `format_configuration` - (Optional) Format configuration. See [`format_configuration` Block](#format_configuration-block) below.
* `role` - (Optional) Role.

### `column_header_style` Block

* `background_color` - (Optional) Background color.
* `border` - (Optional) Border. See [`border` Block](#border-block) below.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `height` - (Optional) Height.
* `horizontal_text_alignment` - (Optional) Horizontal text alignment.
* `text_wrap` - (Optional) Text wrap.
* `vertical_text_alignment` - (Optional) Vertical text alignment.
* `visibility` - (Optional) Visibility.

### `column_hierarchies` Block

* `date_time_hierarchy` - (Optional) Date time hierarchy. See [`date_time_hierarchy` Block](#date_time_hierarchy-block) below.
* `explicit_hierarchy` - (Optional) Explicit hierarchy. See [`explicit_hierarchy` Block](#explicit_hierarchy-block) below.
* `predefined_hierarchy` - (Optional) Predefined hierarchy. See [`predefined_hierarchy` Block](#predefined_hierarchy-block) below.

### `column_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

### `column_sort` Block

* `aggregation_function` - (Optional) Aggregation function. See [`aggregation_function` Block](#aggregation_function-block) below.
* `direction` - (Required) Direction.
* `sort_by` - (Required) Sort by. See [`sort_by` Block](#sort_by-block) below.

### `column_subtotal_options` Block

* `custom_label` - (Optional) Custom label.
* `field_level` - (Optional) Field level.
* `field_level_options` - (Optional) Field level options. See [`field_level_options` Block](#field_level_options-block) below.
* `metric_header_cell_style` - (Optional) Metric header cell style. See [`metric_header_cell_style` Block](#metric_header_cell_style-block) below.
* `total_cell_style` - (Optional) Total cell style. See [`total_cell_style` Block](#total_cell_style-block) below.
* `totals_visibility` - (Optional) Totals visibility.
* `value_cell_style` - (Optional) Value cell style. See [`value_cell_style` Block](#value_cell_style-block) below.

### `column_to_match` Block

* `column_name` - (Required) Column name.
* `data_set_identifier` - (Required) Data set identifier.

### `column_tooltip_item` Block

* `aggregation` - (Optional) Aggregation. See [`aggregation` Block](#aggregation-block) below.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `label` - (Optional) Label.
* `visibility` - (Optional) Visibility.

### `column_total_options` Block

* `custom_label` - (Optional) Custom label.
* `metric_header_cell_style` - (Optional) Metric header cell style. See [`metric_header_cell_style` Block](#metric_header_cell_style-block) below.
* `placement` - (Optional) Placement.
* `scroll_status` - (Optional) Scroll status.
* `total_cell_style` - (Optional) Total cell style. See [`total_cell_style` Block](#total_cell_style-block) below.
* `totals_visibility` - (Optional) Totals visibility.
* `value_cell_style` - (Optional) Value cell style. See [`value_cell_style` Block](#value_cell_style-block) below.

### `columns` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `column_name` - (Required) Column name.
* `data_set_identifier` - (Required) Data set identifier.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `combo_chart_aggregated_field_wells` Block

* `bar_values` - (Optional) Bar values. See [`bar_values` Block](#bar_values-block) below.
* `category` - (Optional) Category. See [`category` Block](#category-block) below.
* `colors` - (Optional) Colors. See [`colors` Block](#colors-block) below.
* `line_values` - (Optional) Line values. See [`line_values` Block](#line_values-block) below.

### `combo_chart_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `comparison` Block

* `comparison_format` - (Optional) Comparison format. See [`comparison_format` Block](#comparison_format-block) below.
* `comparison_method` - (Optional) Comparison method.

### `comparison_format` Block

* `number_display_format_configuration` - (Optional) Number display format configuration. See [`number_display_format_configuration` Block](#number_display_format_configuration-block) below.
* `percentage_display_format_configuration` - (Optional) Percentage display format configuration. See [`percentage_display_format_configuration` Block](#percentage_display_format_configuration-block) below.

### `comparison_value` Block

* `icon` - (Optional) Icon. See [`icon` Block](#icon-block) below.
* `text_color` - (Required) Text color. See [`text_color` Block](#text_color-block) below.

### `computation` Block

* `forecast` - (Optional) Forecast. See [`forecast` Block](#forecast-block) below.
* `growth_rate` - (Optional) Growth rate. See [`growth_rate` Block](#growth_rate-block) below.
* `maximum_minimum` - (Optional) Maximum minimum. See [`maximum_minimum` Block](#maximum_minimum-block) below.
* `metric_comparison` - (Optional) Metric comparison. See [`metric_comparison` Block](#metric_comparison-block) below.
* `period_over_period` - (Optional) Period over period. See [`period_over_period` Block](#period_over_period-block) below.
* `period_to_date` - (Optional) Period to date. See [`period_to_date` Block](#period_to_date-block) below.
* `top_bottom_movers` - (Optional) Top bottom movers. See [`top_bottom_movers` Block](#top_bottom_movers-block) below.
* `top_bottom_ranked` - (Optional) Top bottom ranked. See [`top_bottom_ranked` Block](#top_bottom_ranked-block) below.
* `total_aggregation` - (Optional) Total aggregation. See [`total_aggregation` Block](#total_aggregation-block) below.
* `unique_values` - (Optional) Unique values. See [`unique_values` Block](#unique_values-block) below.

### `conditional_formatting` Block

* `conditional_formatting_options` - (Optional) Conditional formatting options. See [`conditional_formatting_options` Block](#conditional_formatting_options-block) below.

### `conditional_formatting_options` Block

* `actual_value` - (Optional) Actual value. See [`actual_value` Block](#actual_value-block) below.
* `arc` - (Optional) Arc. See [`arc` Block](#arc-block) below.
* `cell` - (Optional) Cell. See [`cell` Block](#cell-block) below.
* `comparison_value` - (Optional) Comparison value. See [`comparison_value` Block](#comparison_value-block) below.
* `primary_value` - (Optional) Primary value. See [`primary_value` Block](#primary_value-block) below.
* `progress_bar` - (Optional) Progress bar. See [`progress_bar` Block](#progress_bar-block) below.
* `row` - (Optional) Row. See [`row` Block](#row-block) below.
* `shape` - (Required) Shape. See [`shape` Block](#shape-block) below.

### `configuration` Block

* `custom_filter_configuration` - (Optional) Custom filter configuration. See [`custom_filter_configuration` Block](#custom_filter_configuration-block) below.
* `custom_filter_list_configuration` - (Optional) Custom filter list configuration. See [`custom_filter_list_configuration` Block](#custom_filter_list_configuration-block) below.
* `filter_list_configuration` - (Optional) Filter list configuration. See [`filter_list_configuration` Block](#filter_list_configuration-block) below.
* `free_form_layout` - (Optional) Free form layout. See [`free_form_layout` Block](#free_form_layout-block) below.
* `grid_layout` - (Optional) Grid layout. See [`grid_layout` Block](#grid_layout-block) below.
* `section_based_layout` - (Optional) Section based layout. See [`section_based_layout` Block](#section_based_layout-block) below.

### `configuration_overrides` Block

* `visibility` - (Optional) Visibility.

### `content` Block

* `custom_icon_content` - (Optional) Custom icon content. See [`custom_icon_content` Block](#custom_icon_content-block) below.
* `custom_text_content` - (Optional) Custom text content. See [`custom_text_content` Block](#custom_text_content-block) below.
* `layout` - (Optional) Layout. See [`layout` Block](#layout-block) below.

### `contribution_analysis_defaults` Block

* `contributor_dimensions` - (Required) Contributor dimensions. See [`contributor_dimensions` Block](#contributor_dimensions-block) below.
* `measure_field_id` - (Required) Measure field ID.

### `contributor_dimensions` Block

* `column_name` - (Required) Column name.
* `data_set_identifier` - (Required) Data set identifier.

### `currency_display_format_configuration` Block

* `decimal_places_configuration` - (Optional) Decimal places configuration. See [`decimal_places_configuration` Block](#decimal_places_configuration-block) below.
* `negative_value_configuration` - (Optional) Negative value configuration. See [`negative_value_configuration` Block](#negative_value_configuration-block) below.
* `null_value_format_configuration` - (Optional) Null value format configuration. See [`null_value_format_configuration` Block](#null_value_format_configuration-block) below.
* `number_scale` - (Optional) Number scale.
* `prefix` - (Optional) Prefix.
* `separator_configuration` - (Optional) Separator configuration. See [`separator_configuration` Block](#separator_configuration-block) below.
* `suffix` - (Optional) Suffix.
* `symbol` - (Optional) Symbol.

### `custom_condition` Block

* `color` - (Optional) Color.
* `display_configuration` - (Optional) Display configuration. See [`display_configuration` Block](#display_configuration-block) below.
* `expression` - (Required) Expression.
* `icon_options` - (Required) Icon options. See [`icon_options` Block](#icon_options-block) below.

### `custom_content_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `data_set_identifier` - (Required) Data set identifier.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `custom_filter_configuration` Block

* `category_value` - (Optional) Category value.
* `match_operator` - (Required) Match operator.
* `null_option` - (Required) Null option.
* `parameter_name` - (Optional) Parameter name.
* `select_all_options` - (Optional) Select all options.

### `custom_filter_list_configuration` Block

* `category_values` - (Optional) Category values.
* `match_operator` - (Required) Match operator.
* `null_option` - (Required) Null option.
* `select_all_options` - (Optional) Select all options.

### `custom_icon_content` Block

* `icon` - (Optional) Icon.

### `custom_label_configuration` Block

* `custom_label` - (Required) Custom label.

### `custom_narrative` Block

* `narrative` - (Required) Narrative.

### `custom_text_content` Block

* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `value` - (Optional) Value.

### `custom_values` Block

* `date_time_values` - (Optional) Date time values.
* `decimal_values` - (Optional) Decimal values.
* `integer_values` - (Optional) Integer values.
* `string_values` - (Optional) String values.

### `custom_values_configuration` Block

* `custom_values` - (Required) Custom values. See [`custom_values` Block](#custom_values-block) below.
* `include_null_value` - (Optional) Include null value.

### `data_bars` Block

* `field_id` - (Required) Field ID.
* `negative_color` - (Optional) Negative color.
* `positive_color` - (Optional) Positive color.

### `data_configuration` Block

* `axis_binding` - (Optional) Axis binding.
* `dynamic_configuration` - (Optional) Dynamic configuration. See [`dynamic_configuration` Block](#dynamic_configuration-block) below.
* `static_configuration` - (Optional) Static configuration. See [`static_configuration` Block](#static_configuration-block) below.

### `data_driven` Block

### `data_field_series_item` Block

* `axis_binding` - (Required) Axis binding.
* `field_id` - (Required) Field ID.
* `field_value` - (Optional) Field value.
* `settings` - (Optional) Settings. See [`settings` Block](#settings-block) below.

### `data_label_options` Block

* `category_label_visibility` - (Optional) Category label visibility.
* `label_color` - (Optional) Label color.
* `label_font_configuration` - (Optional) Label font configuration. See [`label_font_configuration` Block](#label_font_configuration-block) below.
* `measure_data_label_style` - (Optional) Measure data label style.
* `measure_label_visibility` - (Optional) Measure label visibility.
* `position` - (Optional) Position.
* `visibility` - (Optional) Visibility.

### `data_label_types` Block

* `data_path_label_type` - (Optional) Data path label type. See [`data_path_label_type` Block](#data_path_label_type-block) below.
* `field_label_type` - (Optional) Field label type. See [`field_label_type` Block](#field_label_type-block) below.
* `maximum_label_type` - (Optional) Maximum label type. See [`maximum_label_type` Block](#maximum_label_type-block) below.
* `minimum_label_type` - (Optional) Minimum label type. See [`minimum_label_type` Block](#minimum_label_type-block) below.
* `range_ends_label_type` - (Optional) Range ends label type. See [`range_ends_label_type` Block](#range_ends_label_type-block) below.

### `data_labels` Block

* `category_label_visibility` - (Optional) Category label visibility.
* `data_label_types` - (Optional) Data label types. See [`data_label_types` Block](#data_label_types-block) below.
* `label_color` - (Optional) Label color.
* `label_content` - (Optional) Label content.
* `label_font_configuration` - (Optional) Label font configuration. See [`label_font_configuration` Block](#label_font_configuration-block) below.
* `measure_label_visibility` - (Optional) Measure label visibility.
* `overlap` - (Optional) Overlap.
* `position` - (Optional) Position.
* `visibility` - (Optional) Visibility.

### `data_options` Block

* `date_axis_options` - (Optional) Date axis options. See [`date_axis_options` Block](#date_axis_options-block) below.
* `numeric_axis_options` - (Optional) Numeric axis options. See [`numeric_axis_options` Block](#numeric_axis_options-block) below.

### `data_path` Block

* `direction` - (Required) Direction.
* `sort_paths` - (Required) Sort paths. See [`sort_paths` Block](#sort_paths-block) below.

### `data_path_label_type` Block

* `field_id` - (Optional) Field ID.
* `field_value` - (Optional) Field value.
* `visibility` - (Optional) Visibility.

### `data_path_list` Block

* `field_id` - (Required) Field ID.
* `field_value` - (Required) Field value.

### `data_path_options` Block

* `data_path_list` - (Required) Data path list. See [`data_path_list` Block](#data_path_list-block) below.
* `width` - (Optional) Width.

### `data_set_identifiers_declarations` Block

* `data_set_arn` - (Optional) Data set ARN.
* `identifier` - (Optional) Identifier.

### `data_set_references` Block

* `data_set_arn` - (Required) Data set ARN.
* `data_set_placeholder` - (Required) Data set placeholder.

### `date_axis_options` Block

* `missing_date_visibility` - (Optional) Missing date visibility.

### `date_dimension_field` Block

* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `date_granularity` - (Optional) Date granularity.
* `field_id` - (Required) Field ID.
* `format_configuration` - (Optional) Format configuration. See [`format_configuration` Block](#format_configuration-block) below.
* `hierarchy_id` - (Optional) Hierarchy ID.

### `date_measure_field` Block

* `aggregation_function` - (Optional) Aggregation function.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `field_id` - (Required) Field ID.
* `format_configuration` - (Optional) Format configuration. See [`format_configuration` Block](#format_configuration-block) below.

### `date_time_format_configuration` Block

* `date_time_format` - (Optional) Date time format.
* `null_value_format_configuration` - (Optional) Null value format configuration. See [`null_value_format_configuration` Block](#null_value_format_configuration-block) below.
* `numeric_format_configuration` - (Optional) Numeric format configuration. See [`numeric_format_configuration` Block](#numeric_format_configuration-block) below.

### `date_time_hierarchy` Block

* `drill_down_filters` - (Optional) Drill down filters. See [`drill_down_filters` Block](#drill_down_filters-block) below.
* `hierarchy_id` - (Required) Hierarchy ID.

### `date_time_parameter_declaration` Block

* `default_values` - (Optional) Default values. See [`default_values` Block](#default_values-block) below.
* `name` - (Required) Name.
* `time_granularity` - (Optional) Time granularity.
* `values_when_unset` - (Optional) Values when unset. See [`values_when_unset` Block](#values_when_unset-block) below.

### `date_time_parameters` Block

* `name` - (Required) Name.
* `values` - (Required) Values.

### `date_time_picker` Block

* `display_options` - (Optional) Display options. See [`display_options` Block](#display_options-block) below.
* `filter_control_id` - (Required) Filter control ID.
* `parameter_control_id` - (Required) Parameter control ID.
* `source_filter_id` - (Required) Source filter ID.
* `source_parameter_name` - (Required) Source parameter name.
* `title` - (Required) Title.
* `type` - (Optional) Type.

### `decimal_parameter_declaration` Block

* `default_values` - (Optional) Default values. See [`default_values` Block](#default_values-block) below.
* `name` - (Required) Name.
* `parameter_value_type` - (Required) Parameter value type.
* `values_when_unset` - (Optional) Values when unset. See [`values_when_unset` Block](#values_when_unset-block) below.

### `decimal_parameters` Block

* `name` - (Required) Name.
* `values` - (Required) Values.

### `decimal_places_configuration` Block

* `decimal_places` - (Required) Decimal places.

### `default_new_sheet_configuration` Block

* `interactive_layout_configuration` - (Optional) Interactive layout configuration. See [`interactive_layout_configuration` Block](#interactive_layout_configuration-block) below.
* `paginated_layout_configuration` - (Optional) Paginated layout configuration. See [`paginated_layout_configuration` Block](#paginated_layout_configuration-block) below.
* `sheet_content_type` - (Optional) Sheet content type.

### `default_series_settings` Block

* `axis_binding` - (Optional) Axis binding.
* `line_style_settings` - (Optional) Line style settings. See [`line_style_settings` Block](#line_style_settings-block) below.
* `marker_style_settings` - (Optional) Marker style settings. See [`marker_style_settings` Block](#marker_style_settings-block) below.

### `default_value_column` Block

* `column_name` - (Required) Column name.
* `data_set_identifier` - (Required) Data set identifier.

### `default_values` Block

* `dynamic_value` - (Optional) Dynamic value. See [`dynamic_value` Block](#dynamic_value-block) below.
* `rolling_date` - (Optional) Rolling date. See [`rolling_date` Block](#rolling_date-block) below.
* `static_values` - (Optional) Static values.

### `definition` Block

* `analysis_defaults` - (Optional) Analysis defaults. See [`analysis_defaults` Block](#analysis_defaults-block) below.
* `calculated_fields` - (Optional) Calculated fields. See [`calculated_fields` Block](#calculated_fields-block) below.
* `column_configurations` - (Optional) Column configurations. See [`column_configurations` Block](#column_configurations-block) below.
* `data_set_identifiers_declarations` - (Required) Data set identifiers declarations. See [`data_set_identifiers_declarations` Block](#data_set_identifiers_declarations-block) below.
* `filter_groups` - (Optional) Filter groups. See [`filter_groups` Block](#filter_groups-block) below.
* `parameter_declarations` - (Optional) Parameter declarations. See [`parameter_declarations` Block](#parameter_declarations-block) below.
* `sheets` - (Optional) Sheets. See [`sheets` Block](#sheets-block) below.

### `destination` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `destination_items_limit` Block

* `items_limit` - (Optional) Items limit.
* `other_categories` - (Required) Other categories.

### `display_configuration` Block

* `icon_display_option` - (Optional) Icon display option.

### `display_options` Block

* `date_time_format` - (Optional) Date time format.
* `placeholder_options` - (Optional) Placeholder options. See [`placeholder_options` Block](#placeholder_options-block) below.
* `search_options` - (Optional) Search options. See [`search_options` Block](#search_options-block) below.
* `select_all_options` - (Optional) Select all options. See [`select_all_options` Block](#select_all_options-block) below.
* `title_options` - (Optional) Title options. See [`title_options` Block](#title_options-block) below.

### `donut_center_options` Block

* `label_visibility` - (Optional) Label visibility.

### `donut_options` Block

* `arc_options` - (Optional) Arc options. See [`arc_options` Block](#arc_options-block) below.
* `donut_center_options` - (Optional) Donut center options. See [`donut_center_options` Block](#donut_center_options-block) below.

### `drill_down_filters` Block

* `category_filter` - (Optional) Category filter. See [`category_filter` Block](#category_filter-block) below.
* `numeric_equality_filter` - (Optional) Numeric equality filter. See [`numeric_equality_filter` Block](#numeric_equality_filter-block) below.
* `time_range_filter` - (Optional) Time range filter. See [`time_range_filter` Block](#time_range_filter-block) below.

### `dropdown` Block

* `cascading_control_configuration` - (Optional) Cascading control configuration. See [`cascading_control_configuration` Block](#cascading_control_configuration-block) below.
* `display_options` - (Optional) Display options. See [`display_options` Block](#display_options-block) below.
* `filter_control_id` - (Required) Filter control ID.
* `parameter_control_id` - (Required) Parameter control ID.
* `selectable_values` - (Optional) Selectable values. See [`selectable_values` Block](#selectable_values-block) below.
* `source_filter_id` - (Required) Source filter ID.
* `source_parameter_name` - (Required) Source parameter name.
* `title` - (Required) Title.
* `type` - (Optional) Type.

### `dynamic_configuration` Block

* `calculation` - (Required) Calculation. See [`calculation` Block](#calculation-block) below.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `measure_aggregation_function` - (Required) Measure aggregation function. See [`measure_aggregation_function` Block](#measure_aggregation_function-block) below.

### `dynamic_value` Block

* `default_value_column` - (Required) Default value column. See [`default_value_column` Block](#default_value_column-block) below.
* `group_name_column` - (Optional) Group name column. See [`group_name_column` Block](#group_name_column-block) below.
* `user_name_column` - (Optional) User name column. See [`user_name_column` Block](#user_name_column-block) below.

### `element` Block

* `field_id` - (Required) Field ID.
* `field_value` - (Required) Field value.

### `elements` Block

* `background_style` - (Optional) Background style. See [`background_style` Block](#background_style-block) below.
* `border_style` - (Optional) Border style. See [`border_style` Block](#border_style-block) below.
* `column_index` - (Optional) Column index.
* `column_span` - (Required) Column span.
* `element_id` - (Required) Element ID.
* `element_type` - (Required) Element type.
* `height` - (Required) Height.
* `loading_animation` - (Optional) Loading animation. See [`loading_animation` Block](#loading_animation-block) below.
* `rendering_rules` - (Optional) Rendering rules. See [`rendering_rules` Block](#rendering_rules-block) below.
* `row_index` - (Optional) Row index.
* `row_span` - (Required) Row span.
* `selected_border_style` - (Optional) Selected border style. See [`selected_border_style` Block](#selected_border_style-block) below.
* `visibility` - (Optional) Visibility.
* `width` - (Required) Width.
* `x_axis_location` - (Required) X axis location.
* `y_axis_location` - (Required) Y axis location.

### `empty_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `data_set_identifier` - (Required) Data set identifier.
* `visual_id` - (Required) Visual ID.

### `exclude_period_configuration` Block

* `amount` - (Required) Amount.
* `granularity` - (Required) Granularity.
* `status` - (Optional) Status.

### `explicit_hierarchy` Block

* `columns` - (Required) Columns. See [`columns` Block](#columns-block) below.
* `drill_down_filters` - (Optional) Drill down filters. See [`drill_down_filters` Block](#drill_down_filters-block) below.
* `hierarchy_id` - (Required) Hierarchy ID.

### `field` Block

* `direction` - (Required) Direction.
* `field_id` - (Required) Field ID.

### `field_base_tooltip` Block

* `aggregation_visibility` - (Optional) Aggregation visibility.
* `tooltip_fields` - (Optional) Tooltip fields. See [`tooltip_fields` Block](#tooltip_fields-block) below.
* `tooltip_title_type` - (Optional) Tooltip title type.

### `field_label_type` Block

* `field_id` - (Optional) Field ID.
* `visibility` - (Optional) Visibility.

### `field_level_options` Block

* `field_id` - (Optional) Field ID.

### `field_options` Block

* `data_path_options` - (Optional) Data path options. See [`data_path_options` Block](#data_path_options-block) below.
* `order` - (Optional) Order.
* `selected_field_options` - (Optional) Selected field options. See [`selected_field_options` Block](#selected_field_options-block) below.

### `field_series_item` Block

* `axis_binding` - (Required) Axis binding.
* `field_id` - (Required) Field ID.
* `settings` - (Optional) Settings. See [`settings` Block](#settings-block) below.

### `field_sort` Block

* `direction` - (Required) Direction.
* `field_id` - (Required) Field ID.

### `field_sort_options` Block

* `field_id` - (Required) Field ID.
* `sort_by` - (Required) Sort by. See [`sort_by` Block](#sort_by-block) below.

### `field_tooltip_item` Block

* `field_id` - (Required) Field ID.
* `label` - (Optional) Label.
* `visibility` - (Optional) Visibility.

### `field_wells` Block

* `bar_chart_aggregated_field_wells` - (Optional) Bar chart aggregated field wells. See [`bar_chart_aggregated_field_wells` Block](#bar_chart_aggregated_field_wells-block) below.
* `box_plot_aggregated_field_wells` - (Optional) Box plot aggregated field wells. See [`box_plot_aggregated_field_wells` Block](#box_plot_aggregated_field_wells-block) below.
* `combo_chart_aggregated_field_wells` - (Optional) Combo chart aggregated field wells. See [`combo_chart_aggregated_field_wells` Block](#combo_chart_aggregated_field_wells-block) below.
* `filled_map_aggregated_field_wells` - (Optional) Filled map aggregated field wells. See [`filled_map_aggregated_field_wells` Block](#filled_map_aggregated_field_wells-block) below.
* `funnel_chart_aggregated_field_wells` - (Optional) Funnel chart aggregated field wells. See [`funnel_chart_aggregated_field_wells` Block](#funnel_chart_aggregated_field_wells-block) below.
* `geospatial_map_aggregated_field_wells` - (Optional) Geospatial map aggregated field wells. See [`geospatial_map_aggregated_field_wells` Block](#geospatial_map_aggregated_field_wells-block) below.
* `heat_map_aggregated_field_wells` - (Optional) Heat map aggregated field wells. See [`heat_map_aggregated_field_wells` Block](#heat_map_aggregated_field_wells-block) below.
* `histogram_aggregated_field_wells` - (Optional) Histogram aggregated field wells. See [`histogram_aggregated_field_wells` Block](#histogram_aggregated_field_wells-block) below.
* `line_chart_aggregated_field_wells` - (Optional) Line chart aggregated field wells. See [`line_chart_aggregated_field_wells` Block](#line_chart_aggregated_field_wells-block) below.
* `pie_chart_aggregated_field_wells` - (Optional) Pie chart aggregated field wells. See [`pie_chart_aggregated_field_wells` Block](#pie_chart_aggregated_field_wells-block) below.
* `pivot_table_aggregated_field_wells` - (Optional) Pivot table aggregated field wells. See [`pivot_table_aggregated_field_wells` Block](#pivot_table_aggregated_field_wells-block) below.
* `radar_chart_aggregated_field_wells` - (Optional) Radar chart aggregated field wells. See [`radar_chart_aggregated_field_wells` Block](#radar_chart_aggregated_field_wells-block) below.
* `sankey_diagram_aggregated_field_wells` - (Optional) Sankey diagram aggregated field wells. See [`sankey_diagram_aggregated_field_wells` Block](#sankey_diagram_aggregated_field_wells-block) below.
* `scatter_plot_categorically_aggregated_field_wells` - (Optional) Scatter plot categorically aggregated field wells. See [`scatter_plot_categorically_aggregated_field_wells` Block](#scatter_plot_categorically_aggregated_field_wells-block) below.
* `scatter_plot_unaggregated_field_wells` - (Optional) Scatter plot unaggregated field wells. See [`scatter_plot_unaggregated_field_wells` Block](#scatter_plot_unaggregated_field_wells-block) below.
* `table_aggregated_field_wells` - (Optional) Table aggregated field wells. See [`table_aggregated_field_wells` Block](#table_aggregated_field_wells-block) below.
* `table_unaggregated_field_wells` - (Optional) Table unaggregated field wells. See [`table_unaggregated_field_wells` Block](#table_unaggregated_field_wells-block) below.
* `target_values` - (Optional) Target values. See [`target_values` Block](#target_values-block) below.
* `tree_map_aggregated_field_wells` - (Optional) Tree map aggregated field wells. See [`tree_map_aggregated_field_wells` Block](#tree_map_aggregated_field_wells-block) below.
* `trend_groups` - (Optional) Trend groups. See [`trend_groups` Block](#trend_groups-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.
* `waterfall_chart_aggregated_field_wells` - (Optional) Waterfall chart aggregated field wells. See [`waterfall_chart_aggregated_field_wells` Block](#waterfall_chart_aggregated_field_wells-block) below.
* `word_cloud_aggregated_field_wells` - (Optional) Word cloud aggregated field wells. See [`word_cloud_aggregated_field_wells` Block](#word_cloud_aggregated_field_wells-block) below.

### `filled_map_aggregated_field_wells` Block

* `geospatial` - (Optional) Geospatial. See [`geospatial` Block](#geospatial-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `filled_map_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `conditional_formatting` - (Optional) Conditional formatting. See [`conditional_formatting` Block](#conditional_formatting-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `filter_controls` Block

* `date_time_picker` - (Optional) Date time picker. See [`date_time_picker` Block](#date_time_picker-block) below.
* `dropdown` - (Optional) Dropdown. See [`dropdown` Block](#dropdown-block) below.
* `list` - (Optional) List. See [`list` Block](#list-block) below.
* `relative_date_time` - (Optional) Relative date time. See [`relative_date_time` Block](#relative_date_time-block) below.
* `slider` - (Optional) Slider. See [`slider` Block](#slider-block) below.
* `text_area` - (Optional) Text area. See [`text_area` Block](#text_area-block) below.
* `text_field` - (Optional) Text field. See [`text_field` Block](#text_field-block) below.

### `filter_groups` Block

* `cross_dataset` - (Required) Cross dataset.
* `filter_group_id` - (Required) Filter group ID.
* `filters` - (Required) Filters. See [`filters` Block](#filters-block) below.
* `scope_configuration` - (Required) Scope configuration. See [`scope_configuration` Block](#scope_configuration-block) below.
* `status` - (Optional) Status.

### `filter_list_configuration` Block

* `category_values` - (Optional) Category values.
* `match_operator` - (Required) Match operator.
* `select_all_options` - (Optional) Select all options.

### `filter_operation` Block

* `selected_fields_configuration` - (Required) Selected fields configuration. See [`selected_fields_configuration` Block](#selected_fields_configuration-block) below.
* `target_visuals_configuration` - (Required) Target visuals configuration. See [`target_visuals_configuration` Block](#target_visuals_configuration-block) below.

### `filters` Block

* `category_filter` - (Optional) Category filter. See [`category_filter` Block](#category_filter-block) below.
* `numeric_equality_filter` - (Optional) Numeric equality filter. See [`numeric_equality_filter` Block](#numeric_equality_filter-block) below.
* `numeric_range_filter` - (Optional) Numeric range filter. See [`numeric_range_filter` Block](#numeric_range_filter-block) below.
* `relative_dates_filter` - (Optional) Relative dates filter. See [`relative_dates_filter` Block](#relative_dates_filter-block) below.
* `time_equality_filter` - (Optional) Time equality filter. See [`time_equality_filter` Block](#time_equality_filter-block) below.
* `time_range_filter` - (Optional) Time range filter. See [`time_range_filter` Block](#time_range_filter-block) below.
* `top_bottom_filter` - (Optional) Top bottom filter. See [`top_bottom_filter` Block](#top_bottom_filter-block) below.

### `font_configuration` Block

* `font_color` - (Optional) Font color.
* `font_decoration` - (Optional) Font decoration.
* `font_size` - (Optional) Font size. See [`font_size` Block](#font_size-block) below.
* `font_style` - (Optional) Font style.
* `font_weight` - (Optional) Font weight. See [`font_weight` Block](#font_weight-block) below.

### `font_size` Block

* `relative` - (Optional) Relative.

### `font_weight` Block

* `name` - (Optional) Name.

### `footer_sections` Block

* `layout` - (Optional) Layout. See [`layout` Block](#layout-block) below.
* `section_id` - (Required) Section ID.
* `style` - (Optional) Style. See [`style` Block](#style-block) below.

### `forecast` Block

* `computation_id` - (Required) Computation ID.
* `custom_seasonality_value` - (Optional) Custom seasonality value.
* `lower_boundary` - (Optional) Lower boundary.
* `name` - (Optional) Name.
* `periods_backward` - (Optional) Periods backward.
* `periods_forward` - (Optional) Periods forward.
* `prediction_interval` - (Optional) Prediction interval.
* `seasonality` - (Required) Seasonality.
* `time` - (Optional) Time. See [`time` Block](#time-block) below.
* `upper_boundary` - (Optional) Upper boundary.
* `value` - (Optional) Value. See [`value` Block](#value-block) below.

### `forecast_configurations` Block

* `forecast_properties` - (Optional) Forecast properties. See [`forecast_properties` Block](#forecast_properties-block) below.
* `scenario` - (Optional) Scenario. See [`scenario` Block](#scenario-block) below.

### `forecast_properties` Block

* `lower_boundary` - (Optional) Lower boundary.
* `periods_backward` - (Optional) Periods backward.
* `periods_forward` - (Optional) Periods forward.
* `prediction_interval` - (Optional) Prediction interval.
* `seasonality` - (Optional) Seasonality.
* `upper_boundary` - (Optional) Upper boundary.

### `foreground_color` Block

* `gradient` - (Optional) Gradient. See [`gradient` Block](#gradient-block) below.
* `solid` - (Optional) Solid. See [`solid` Block](#solid-block) below.

### `format` Block

* `background_color` - (Required) Background color. See [`background_color` Block](#background_color-block) below.

### `format_configuration` Block

* `currency_display_format_configuration` - (Optional) Currency display format configuration. See [`currency_display_format_configuration` Block](#currency_display_format_configuration-block) below.
* `date_time_format` - (Optional) Date time format.
* `date_time_format_configuration` - (Optional) Date time format configuration. See [`date_time_format_configuration` Block](#date_time_format_configuration-block) below.
* `null_value_format_configuration` - (Optional) Null value format configuration. See [`null_value_format_configuration` Block](#null_value_format_configuration-block) below.
* `number_display_format_configuration` - (Optional) Number display format configuration. See [`number_display_format_configuration` Block](#number_display_format_configuration-block) below.
* `number_format_configuration` - (Optional) Number format configuration. See [`number_format_configuration` Block](#number_format_configuration-block) below.
* `numeric_format_configuration` - (Optional) Numeric format configuration. See [`numeric_format_configuration` Block](#numeric_format_configuration-block) below.
* `percentage_display_format_configuration` - (Optional) Percentage display format configuration. See [`percentage_display_format_configuration` Block](#percentage_display_format_configuration-block) below.
* `string_format_configuration` - (Optional) String format configuration. See [`string_format_configuration` Block](#string_format_configuration-block) below.

### `format_text` Block

* `plain_text` - (Optional) Plain text.
* `rich_text` - (Optional) Rich text.

### `free_form` Block

* `canvas_size_options` - (Required) Canvas size options. See [`canvas_size_options` Block](#canvas_size_options-block) below.

### `free_form_layout` Block

* `canvas_size_options` - (Optional) Canvas size options. See [`canvas_size_options` Block](#canvas_size_options-block) below.
* `elements` - (Required) Elements. See [`elements` Block](#elements-block) below.

### `from_value` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `funnel_chart_aggregated_field_wells` Block

* `category` - (Optional) Category. See [`category` Block](#category-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `funnel_chart_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `gauge_chart_options` Block

* `arc` - (Optional) Arc. See [`arc` Block](#arc-block) below.
* `arc_axis` - (Optional) Arc axis. See [`arc_axis` Block](#arc_axis-block) below.
* `comparison` - (Optional) Comparison. See [`comparison` Block](#comparison-block) below.
* `primary_value_display_type` - (Optional) Primary value display type.
* `primary_value_font_configuration` - (Optional) Primary value font configuration. See [`primary_value_font_configuration` Block](#primary_value_font_configuration-block) below.

### `gauge_chart_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `conditional_formatting` - (Optional) Conditional formatting. See [`conditional_formatting` Block](#conditional_formatting-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `geospatial` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `geospatial_map_aggregated_field_wells` Block

* `colors` - (Optional) Colors. See [`colors` Block](#colors-block) below.
* `geospatial` - (Optional) Geospatial. See [`geospatial` Block](#geospatial-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `geospatial_map_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `gradient` Block

* `color` - (Required) Color. See [`color` Block](#color-block) below.
* `expression` - (Required) Expression.

### `grid` Block

* `canvas_size_options` - (Required) Canvas size options. See [`canvas_size_options` Block](#canvas_size_options-block) below.

### `grid_layout` Block

* `canvas_size_options` - (Optional) Canvas size options. See [`canvas_size_options` Block](#canvas_size_options-block) below.
* `elements` - (Required) Elements. See [`elements` Block](#elements-block) below.

### `group_by` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `group_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

### `group_name_column` Block

* `column_name` - (Required) Column name.
* `data_set_identifier` - (Required) Data set identifier.

### `groups` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `growth_rate` Block

* `computation_id` - (Required) Computation ID.
* `name` - (Optional) Name.
* `period_size` - (Optional) Period size.
* `time` - (Optional) Time. See [`time` Block](#time-block) below.
* `value` - (Optional) Value. See [`value` Block](#value-block) below.

### `header_sections` Block

* `layout` - (Optional) Layout. See [`layout` Block](#layout-block) below.
* `section_id` - (Required) Section ID.
* `style` - (Optional) Style. See [`style` Block](#style-block) below.

### `header_style` Block

* `background_color` - (Optional) Background color.
* `border` - (Optional) Border. See [`border` Block](#border-block) below.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `height` - (Optional) Height.
* `horizontal_text_alignment` - (Optional) Horizontal text alignment.
* `text_wrap` - (Optional) Text wrap.
* `vertical_text_alignment` - (Optional) Vertical text alignment.
* `visibility` - (Optional) Visibility.

### `heat_map_aggregated_field_wells` Block

* `columns` - (Optional) Columns. See [`columns` Block](#columns-block) below.
* `rows` - (Optional) Rows. See [`rows` Block](#rows-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `heat_map_column_items_limit_configuration` Block

* `items_limit` - (Optional) Items limit.
* `other_categories` - (Required) Other categories.

### `heat_map_column_sort` Block

* `column_sort` - (Optional) Column sort. See [`column_sort` Block](#column_sort-block) below.
* `field_sort` - (Optional) Field sort. See [`field_sort` Block](#field_sort-block) below.

### `heat_map_row_items_limit_configuration` Block

* `items_limit` - (Optional) Items limit.
* `other_categories` - (Required) Other categories.

### `heat_map_row_sort` Block

* `column_sort` - (Optional) Column sort. See [`column_sort` Block](#column_sort-block) below.
* `field_sort` - (Optional) Field sort. See [`field_sort` Block](#field_sort-block) below.

### `heat_map_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `histogram_aggregated_field_wells` Block

* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `histogram_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `icon` Block

* `custom_condition` - (Optional) Custom condition. See [`custom_condition` Block](#custom_condition-block) below.
* `icon_set` - (Optional) Icon set. See [`icon_set` Block](#icon_set-block) below.

### `icon_options` Block

* `icon` - (Optional) Icon.
* `unicode_icon` - (Optional) Unicode icon.

### `icon_set` Block

* `expression` - (Required) Expression.
* `icon_set_type` - (Optional) Icon set type.

### `image_configuration` Block

* `sizing_options` - (Optional) Sizing options. See [`sizing_options` Block](#sizing_options-block) below.

### `inner_horizontal` Block

* `color` - (Optional) Color.
* `style` - (Optional) Style.
* `thickness` - (Optional) Thickness.

### `inner_vertical` Block

* `color` - (Optional) Color.
* `style` - (Optional) Style.
* `thickness` - (Optional) Thickness.

### `insight_configuration` Block

* `computation` - (Optional) Computation. See [`computation` Block](#computation-block) below.
* `custom_narrative` - (Optional) Custom narrative. See [`custom_narrative` Block](#custom_narrative-block) below.

### `insight_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `data_set_identifier` - (Required) Data set identifier.
* `insight_configuration` - (Optional) Insight configuration. See [`insight_configuration` Block](#insight_configuration-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `integer_parameter_declaration` Block

* `default_values` - (Optional) Default values. See [`default_values` Block](#default_values-block) below.
* `name` - (Required) Name.
* `parameter_value_type` - (Required) Parameter value type.
* `values_when_unset` - (Optional) Values when unset. See [`values_when_unset` Block](#values_when_unset-block) below.

### `integer_parameters` Block

* `name` - (Required) Name.
* `values` - (Required) Values.

### `interactive_layout_configuration` Block

* `free_form` - (Optional) Free form. See [`free_form` Block](#free_form-block) below.
* `grid` - (Optional) Grid. See [`grid` Block](#grid-block) below.

### `kpi_options` Block

* `comparison` - (Optional) Comparison. See [`comparison` Block](#comparison-block) below.
* `primary_value_display_type` - (Optional) Primary value display type.
* `primary_value_font_configuration` - (Optional) Primary value font configuration. See [`primary_value_font_configuration` Block](#primary_value_font_configuration-block) below.
* `progress_bar` - (Optional) Progress bar. See [`progress_bar` Block](#progress_bar-block) below.
* `secondary_value` - (Optional) Secondary value. See [`secondary_value` Block](#secondary_value-block) below.
* `secondary_value_font_configuration` - (Optional) Secondary value font configuration. See [`secondary_value_font_configuration` Block](#secondary_value_font_configuration-block) below.
* `sparkline` - (Optional) Sparkline. See [`sparkline` Block](#sparkline-block) below.
* `trend_arrows` - (Optional) Trend arrows. See [`trend_arrows` Block](#trend_arrows-block) below.
* `visual_layout_options` - (Optional) Visual layout options. See [`visual_layout_options` Block](#visual_layout_options-block) below.

### `kpi_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `conditional_formatting` - (Optional) Conditional formatting. See [`conditional_formatting` Block](#conditional_formatting-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `label_configuration` Block

* `custom_label_configuration` - (Optional) Custom label configuration. See [`custom_label_configuration` Block](#custom_label_configuration-block) below.
* `font_color` - (Optional) Font color.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `horizontal_position` - (Optional) Horizontal position.
* `value_label_configuration` - (Optional) Value label configuration. See [`value_label_configuration` Block](#value_label_configuration-block) below.
* `vertical_position` - (Optional) Vertical position.

### `label_font_configuration` Block

* `font_color` - (Optional) Font color.
* `font_decoration` - (Optional) Font decoration.
* `font_size` - (Optional) Font size. See [`font_size` Block](#font_size-block) below.
* `font_style` - (Optional) Font style.
* `font_weight` - (Optional) Font weight. See [`font_weight` Block](#font_weight-block) below.

### `label_options` Block

* `custom_label` - (Optional) Custom label.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `visibility` - (Optional) Visibility.

### `layout` Block

* `free_form_layout` - (Required) Free form layout. See [`free_form_layout` Block](#free_form_layout-block) below.

### `layouts` Block

* `configuration` - (Required) Configuration. See [`configuration` Block](#configuration-block) below.

### `left` Block

* `color` - (Optional) Color.
* `style` - (Optional) Style.
* `thickness` - (Optional) Thickness.

### `legend` Block

* `height` - (Optional) Height.
* `position` - (Optional) Position.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visibility` - (Optional) Visibility.
* `width` - (Optional) Width.

### `line_chart_aggregated_field_wells` Block

* `category` - (Optional) Category. See [`category` Block](#category-block) below.
* `colors` - (Optional) Colors. See [`colors` Block](#colors-block) below.
* `small_multiples` - (Optional) Small multiples. See [`small_multiples` Block](#small_multiples-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `line_chart_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `line_data_labels` Block

* `category_label_visibility` - (Optional) Category label visibility.
* `data_label_types` - (Optional) Data label types. See [`data_label_types` Block](#data_label_types-block) below.
* `label_color` - (Optional) Label color.
* `label_content` - (Optional) Label content.
* `label_font_configuration` - (Optional) Label font configuration. See [`label_font_configuration` Block](#label_font_configuration-block) below.
* `measure_label_visibility` - (Optional) Measure label visibility.
* `overlap` - (Optional) Overlap.
* `position` - (Optional) Position.
* `visibility` - (Optional) Visibility.

### `line_style_settings` Block

* `line_interpolation` - (Optional) Line interpolation.
* `line_style` - (Optional) Line style.
* `line_visibility` - (Optional) Line visibility.
* `line_width` - (Optional) Line width.

### `line_values` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `linear` Block

* `step_count` - (Optional) Step count.
* `step_size` - (Optional) Step size.

### `link_configuration` Block

* `content` - (Optional) Content. See [`content` Block](#content-block) below.
* `target` - (Optional) Target.

### `link_to_data_set_column` Block

* `column_name` - (Required) Column name.
* `data_set_identifier` - (Required) Data set identifier.

### `list` Block

* `cascading_control_configuration` - (Optional) Cascading control configuration. See [`cascading_control_configuration` Block](#cascading_control_configuration-block) below.
* `display_options` - (Optional) Display options. See [`display_options` Block](#display_options-block) below.
* `filter_control_id` - (Required) Filter control ID.
* `parameter_control_id` - (Required) Parameter control ID.
* `selectable_values` - (Optional) Selectable values. See [`selectable_values` Block](#selectable_values-block) below.
* `source_filter_id` - (Required) Source filter ID.
* `source_parameter_name` - (Required) Source parameter name.
* `title` - (Required) Title.
* `type` - (Optional) Type.

### `loading_animation` Block

* `visibility` - (Optional) Visibility.

### `local_navigation_configuration` Block

* `target_sheet_id` - (Required) Target sheet ID.

### `logarithmic` Block

* `base` - (Optional) Base.

### `map_style_options` Block

* `base_map_style` - (Optional) Base map style.

### `marker_style_settings` Block

* `marker_color` - (Optional) Marker color.
* `marker_shape` - (Optional) Marker shape.
* `marker_size` - (Optional) Marker size.
* `marker_visibility` - (Optional) Marker visibility.

### `maximum_label_type` Block

* `visibility` - (Optional) Visibility.

### `maximum_minimum` Block

* `computation_id` - (Required) Computation ID.
* `name` - (Optional) Name.
* `time` - (Optional) Time. See [`time` Block](#time-block) below.
* `type` - (Required) Type.
* `value` - (Optional) Value. See [`value` Block](#value-block) below.

### `measure_aggregation_function` Block

* `categorical_aggregation_function` - (Optional) Categorical aggregation function.
* `date_aggregation_function` - (Optional) Date aggregation function.
* `numerical_aggregation_function` - (Optional) Numerical aggregation function. See [`numerical_aggregation_function` Block](#numerical_aggregation_function-block) below.

### `metric_comparison` Block

* `computation_id` - (Required) Computation ID.
* `from_value` - (Optional) From value. See [`from_value` Block](#from_value-block) below.
* `name` - (Optional) Name.
* `target_value` - (Optional) Target value. See [`target_value` Block](#target_value-block) below.
* `time` - (Optional) Time. See [`time` Block](#time-block) below.

### `metric_header_cell_style` Block

* `background_color` - (Optional) Background color.
* `border` - (Optional) Border. See [`border` Block](#border-block) below.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `height` - (Optional) Height.
* `horizontal_text_alignment` - (Optional) Horizontal text alignment.
* `text_wrap` - (Optional) Text wrap.
* `vertical_text_alignment` - (Optional) Vertical text alignment.
* `visibility` - (Optional) Visibility.

### `min_max` Block

* `maximum` - (Optional) Maximum.
* `minimum` - (Optional) Minimum.

### `minimum_label_type` Block

* `visibility` - (Optional) Visibility.

### `missing_data_configuration` Block

* `treatment_option` - (Optional) Treatment option.

### `navigation_operation` Block

* `local_navigation_configuration` - (Optional) Local navigation configuration. See [`local_navigation_configuration` Block](#local_navigation_configuration-block) below.

### `negative_value_configuration` Block

* `display_mode` - (Required) Display mode.

### `null_value_color` Block

* `color` - (Optional) Color.
* `data_value` - (Optional) Data value.

### `null_value_format_configuration` Block

* `null_string` - (Required) Null string.

### `number_display_format_configuration` Block

* `decimal_places_configuration` - (Optional) Decimal places configuration. See [`decimal_places_configuration` Block](#decimal_places_configuration-block) below.
* `negative_value_configuration` - (Optional) Negative value configuration. See [`negative_value_configuration` Block](#negative_value_configuration-block) below.
* `null_value_format_configuration` - (Optional) Null value format configuration. See [`null_value_format_configuration` Block](#null_value_format_configuration-block) below.
* `number_scale` - (Optional) Number scale.
* `prefix` - (Optional) Prefix.
* `separator_configuration` - (Optional) Separator configuration. See [`separator_configuration` Block](#separator_configuration-block) below.
* `suffix` - (Optional) Suffix.

### `number_format_configuration` Block

* `numeric_format_configuration` - (Optional) Numeric format configuration. See [`numeric_format_configuration` Block](#numeric_format_configuration-block) below.

### `numeric_axis_options` Block

* `range` - (Optional) Range. See [`range` Block](#range-block) below.
* `scale` - (Optional) Scale. See [`scale` Block](#scale-block) below.

### `numeric_equality_filter` Block

* `aggregation_function` - (Optional) Aggregation function. See [`aggregation_function` Block](#aggregation_function-block) below.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `filter_id` - (Required) Filter ID.
* `match_operator` - (Required) Match operator.
* `null_option` - (Required) Null option.
* `parameter_name` - (Optional) Parameter name.
* `select_all_options` - (Optional) Select all options.
* `value` - (Optional) Value.

### `numeric_format_configuration` Block

* `currency_display_format_configuration` - (Optional) Currency display format configuration. See [`currency_display_format_configuration` Block](#currency_display_format_configuration-block) below.
* `number_display_format_configuration` - (Optional) Number display format configuration. See [`number_display_format_configuration` Block](#number_display_format_configuration-block) below.
* `percentage_display_format_configuration` - (Optional) Percentage display format configuration. See [`percentage_display_format_configuration` Block](#percentage_display_format_configuration-block) below.

### `numeric_range_filter` Block

* `aggregation_function` - (Optional) Aggregation function. See [`aggregation_function` Block](#aggregation_function-block) below.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `filter_id` - (Required) Filter ID.
* `include_maximum` - (Optional) Include maximum.
* `include_minimum` - (Optional) Include minimum.
* `null_option` - (Required) Null option.
* `range_maximum` - (Optional) Range maximum. See [`range_maximum` Block](#range_maximum-block) below.
* `range_minimum` - (Optional) Range minimum. See [`range_minimum` Block](#range_minimum-block) below.
* `select_all_options` - (Optional) Select all options.

### `numerical_aggregation_function` Block

* `percentile_aggregation` - (Optional) Percentile aggregation. See [`percentile_aggregation` Block](#percentile_aggregation-block) below.
* `simple_numerical_aggregation` - (Optional) Simple numerical aggregation.

### `numerical_dimension_field` Block

* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `field_id` - (Required) Field ID.
* `format_configuration` - (Optional) Format configuration. See [`format_configuration` Block](#format_configuration-block) below.
* `hierarchy_id` - (Optional) Hierarchy ID.

### `numerical_measure_field` Block

* `aggregation_function` - (Optional) Aggregation function. See [`aggregation_function` Block](#aggregation_function-block) below.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `field_id` - (Required) Field ID.
* `format_configuration` - (Optional) Format configuration. See [`format_configuration` Block](#format_configuration-block) below.

### `padding` Block

* `bottom` - (Optional) Bottom.
* `left` - (Optional) Left.
* `right` - (Optional) Right.
* `top` - (Optional) Top.

### `page_break_configuration` Block

* `after` - (Optional) After. See [`after` Block](#after-block) below.

### `paginated_layout_configuration` Block

* `section_based` - (Optional) Section based. See [`section_based` Block](#section_based-block) below.

### `paginated_report_options` Block

* `overflow_column_header_visibility` - (Optional) Overflow column header visibility.
* `vertical_overflow_visibility` - (Optional) Vertical overflow visibility.

### `pagination_configuration` Block

* `page_number` - (Required) Page number.
* `page_size` - (Required) Page size.

### `panel_configuration` Block

* `background_color` - (Optional) Background color.
* `background_visibility` - (Optional) Background visibility.
* `border_color` - (Optional) Border color.
* `border_style` - (Optional) Border style.
* `border_thickness` - (Optional) Border thickness.
* `border_visibility` - (Optional) Border visibility.
* `gutter_spacing` - (Optional) Gutter spacing.
* `gutter_visibility` - (Optional) Gutter visibility.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.

### `paper_canvas_size_options` Block

* `paper_margin` - (Optional) Paper margin. See [`paper_margin` Block](#paper_margin-block) below.
* `paper_orientation` - (Optional) Paper orientation.
* `paper_size` - (Optional) Paper size.

### `paper_margin` Block

* `bottom` - (Optional) Bottom.
* `left` - (Optional) Left.
* `right` - (Optional) Right.
* `top` - (Optional) Top.

### `parameter_controls` Block

* `date_time_picker` - (Optional) Date time picker. See [`date_time_picker` Block](#date_time_picker-block) below.
* `dropdown` - (Optional) Dropdown. See [`dropdown` Block](#dropdown-block) below.
* `list` - (Optional) List. See [`list` Block](#list-block) below.
* `slider` - (Optional) Slider. See [`slider` Block](#slider-block) below.
* `text_area` - (Optional) Text area. See [`text_area` Block](#text_area-block) below.
* `text_field` - (Optional) Text field. See [`text_field` Block](#text_field-block) below.

### `parameter_declarations` Block

* `date_time_parameter_declaration` - (Optional) Date time parameter declaration. See [`date_time_parameter_declaration` Block](#date_time_parameter_declaration-block) below.
* `decimal_parameter_declaration` - (Optional) Decimal parameter declaration. See [`decimal_parameter_declaration` Block](#decimal_parameter_declaration-block) below.
* `integer_parameter_declaration` - (Optional) Integer parameter declaration. See [`integer_parameter_declaration` Block](#integer_parameter_declaration-block) below.
* `string_parameter_declaration` - (Optional) String parameter declaration. See [`string_parameter_declaration` Block](#string_parameter_declaration-block) below.

### `parameter_value_configurations` Block

* `destination_parameter_name` - (Required) Destination parameter name.
* `value` - (Required) Value. See [`value` Block](#value-block) below.

### `parameters` Block

* `date_time_parameters` - (Optional) Date time parameters. See [`date_time_parameters` Block](#date_time_parameters-block) below.
* `decimal_parameters` - (Optional) Decimal parameters. See [`decimal_parameters` Block](#decimal_parameters-block) below.
* `integer_parameters` - (Optional) Integer parameters. See [`integer_parameters` Block](#integer_parameters-block) below.
* `string_parameters` - (Optional) String parameters. See [`string_parameters` Block](#string_parameters-block) below.

### `percent_range` Block

* `from` - (Optional) From.
* `to` - (Optional) To.

### `percentage_display_format_configuration` Block

* `decimal_places_configuration` - (Optional) Decimal places configuration. See [`decimal_places_configuration` Block](#decimal_places_configuration-block) below.
* `negative_value_configuration` - (Optional) Negative value configuration. See [`negative_value_configuration` Block](#negative_value_configuration-block) below.
* `null_value_format_configuration` - (Optional) Null value format configuration. See [`null_value_format_configuration` Block](#null_value_format_configuration-block) below.
* `prefix` - (Optional) Prefix.
* `separator_configuration` - (Optional) Separator configuration. See [`separator_configuration` Block](#separator_configuration-block) below.
* `suffix` - (Optional) Suffix.

### `percentile_aggregation` Block

* `percentile_value` - (Optional) Percentile value.

### `period_over_period` Block

* `computation_id` - (Required) Computation ID.
* `name` - (Optional) Name.
* `time` - (Optional) Time. See [`time` Block](#time-block) below.
* `value` - (Optional) Value. See [`value` Block](#value-block) below.

### `period_to_date` Block

* `computation_id` - (Required) Computation ID.
* `name` - (Optional) Name.
* `period_time_granularity` - (Required) Period time granularity.
* `time` - (Optional) Time. See [`time` Block](#time-block) below.
* `value` - (Optional) Value. See [`value` Block](#value-block) below.

### `permissions` Block

* `actions` - (Required) Actions.
* `principal` - (Required) Principal.

### `pie_chart_aggregated_field_wells` Block

* `category` - (Optional) Category. See [`category` Block](#category-block) below.
* `small_multiples` - (Optional) Small multiples. See [`small_multiples` Block](#small_multiples-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `pie_chart_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `pivot_table_aggregated_field_wells` Block

* `columns` - (Optional) Columns. See [`columns` Block](#columns-block) below.
* `rows` - (Optional) Rows. See [`rows` Block](#rows-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `pivot_table_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `conditional_formatting` - (Optional) Conditional formatting. See [`conditional_formatting` Block](#conditional_formatting-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `placeholder_options` Block

* `visibility` - (Optional) Visibility.

### `point_style_options` Block

* `cluster_marker_configuration` - (Optional) Cluster marker configuration. See [`cluster_marker_configuration` Block](#cluster_marker_configuration-block) below.
* `selected_point_style` - (Optional) Selected point style.

### `predefined_hierarchy` Block

* `columns` - (Required) Columns. See [`columns` Block](#columns-block) below.
* `drill_down_filters` - (Optional) Drill down filters. See [`drill_down_filters` Block](#drill_down_filters-block) below.
* `hierarchy_id` - (Required) Hierarchy ID.

### `primary_value` Block

* `icon` - (Optional) Icon. See [`icon` Block](#icon-block) below.
* `text_color` - (Required) Text color. See [`text_color` Block](#text_color-block) below.

### `primary_value_font_configuration` Block

* `font_color` - (Optional) Font color.
* `font_decoration` - (Optional) Font decoration.
* `font_size` - (Optional) Font size. See [`font_size` Block](#font_size-block) below.
* `font_style` - (Optional) Font style.
* `font_weight` - (Optional) Font weight. See [`font_weight` Block](#font_weight-block) below.

### `primary_y_axis_display_options` Block

* `axis_line_visibility` - (Optional) Axis line visibility.
* `axis_offset` - (Optional) Axis offset.
* `axis_options` - (Optional) Axis options. See [`axis_options` Block](#axis_options-block) below.
* `data_options` - (Optional) Data options. See [`data_options` Block](#data_options-block) below.
* `grid_line_visibility` - (Optional) Grid line visibility.
* `missing_data_configuration` - (Optional) Missing data configuration. See [`missing_data_configuration` Block](#missing_data_configuration-block) below.
* `scrollbar_options` - (Optional) Scrollbar options. See [`scrollbar_options` Block](#scrollbar_options-block) below.
* `tick_label_options` - (Optional) Tick label options. See [`tick_label_options` Block](#tick_label_options-block) below.

### `primary_y_axis_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

### `progress_bar` Block

* `foreground_color` - (Required) Foreground color. See [`foreground_color` Block](#foreground_color-block) below.
* `visibility` - (Optional) Visibility.

### `radar_chart_aggregated_field_wells` Block

* `category` - (Optional) Category. See [`category` Block](#category-block) below.
* `color` - (Optional) Color. See [`color` Block](#color-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `radar_chart_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `range` Block

* `data_driven` - (Optional) Data driven. See [`data_driven` Block](#data_driven-block) below.
* `max` - (Optional) Max.
* `min` - (Optional) Min.
* `min_max` - (Optional) Min max. See [`min_max` Block](#min_max-block) below.

### `range_ends_label_type` Block

* `visibility` - (Optional) Visibility.

### `range_maximum` Block

* `parameter` - (Optional) Parameter.
* `static_value` - (Optional) Static value.

### `range_maximum_value` Block

* `parameter` - (Optional) Parameter.
* `rolling_date` - (Optional) Rolling date. See [`rolling_date` Block](#rolling_date-block) below.
* `static_value` - (Optional) Static value.

### `range_minimum` Block

* `parameter` - (Optional) Parameter.
* `static_value` - (Optional) Static value.

### `range_minimum_value` Block

* `parameter` - (Optional) Parameter.
* `rolling_date` - (Optional) Rolling date. See [`rolling_date` Block](#rolling_date-block) below.
* `static_value` - (Optional) Static value.

### `reference_lines` Block

* `data_configuration` - (Required) Data configuration. See [`data_configuration` Block](#data_configuration-block) below.
* `label_configuration` - (Optional) Label configuration. See [`label_configuration` Block](#label_configuration-block) below.
* `status` - (Optional) Status.
* `style_configuration` - (Optional) Style configuration. See [`style_configuration` Block](#style_configuration-block) below.

### `relative_date_time` Block

* `display_options` - (Optional) Display options. See [`display_options` Block](#display_options-block) below.
* `filter_control_id` - (Required) Filter control ID.
* `source_filter_id` - (Required) Source filter ID.
* `title` - (Required) Title.

### `relative_dates_filter` Block

* `anchor_date_configuration` - (Required) Anchor date configuration. See [`anchor_date_configuration` Block](#anchor_date_configuration-block) below.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `exclude_period_configuration` - (Optional) Exclude period configuration. See [`exclude_period_configuration` Block](#exclude_period_configuration-block) below.
* `filter_id` - (Required) Filter ID.
* `minimum_granularity` - (Required) Minimum granularity.
* `null_option` - (Required) Null option.
* `parameter_name` - (Optional) Parameter name.
* `relative_date_type` - (Required) Relative date type.
* `relative_date_value` - (Optional) Relative date value.
* `time_granularity` - (Required) Time granularity.

### `rendering_rules` Block

* `configuration_overrides` - (Required) Configuration overrides. See [`configuration_overrides` Block](#configuration_overrides-block) below.
* `expression` - (Required) Expression.

### `right` Block

* `color` - (Optional) Color.
* `style` - (Optional) Style.
* `thickness` - (Optional) Thickness.

### `rolling_date` Block

* `data_set_identifier` - (Optional) Data set identifier.
* `expression` - (Required) Expression.

### `row` Block

* `background_color` - (Required) Background color. See [`background_color` Block](#background_color-block) below.
* `text_color` - (Required) Text color. See [`text_color` Block](#text_color-block) below.

### `row_alternate_color_options` Block

* `row_alternate_colors` - (Optional) Row alternate colors.
* `status` - (Optional) Status.

### `row_field_names_style` Block

* `background_color` - (Optional) Background color.
* `border` - (Optional) Border. See [`border` Block](#border-block) below.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `height` - (Optional) Height.
* `horizontal_text_alignment` - (Optional) Horizontal text alignment.
* `text_wrap` - (Optional) Text wrap.
* `vertical_text_alignment` - (Optional) Vertical text alignment.
* `visibility` - (Optional) Visibility.

### `row_header_style` Block

* `background_color` - (Optional) Background color.
* `border` - (Optional) Border. See [`border` Block](#border-block) below.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `height` - (Optional) Height.
* `horizontal_text_alignment` - (Optional) Horizontal text alignment.
* `text_wrap` - (Optional) Text wrap.
* `vertical_text_alignment` - (Optional) Vertical text alignment.
* `visibility` - (Optional) Visibility.

### `row_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

### `row_sort` Block

* `column_sort` - (Optional) Column sort. See [`column_sort` Block](#column_sort-block) below.
* `field_sort` - (Optional) Field sort. See [`field_sort` Block](#field_sort-block) below.

### `row_subtotal_options` Block

* `custom_label` - (Optional) Custom label.
* `field_level` - (Optional) Field level.
* `field_level_options` - (Optional) Field level options. See [`field_level_options` Block](#field_level_options-block) below.
* `metric_header_cell_style` - (Optional) Metric header cell style. See [`metric_header_cell_style` Block](#metric_header_cell_style-block) below.
* `total_cell_style` - (Optional) Total cell style. See [`total_cell_style` Block](#total_cell_style-block) below.
* `totals_visibility` - (Optional) Totals visibility.
* `value_cell_style` - (Optional) Value cell style. See [`value_cell_style` Block](#value_cell_style-block) below.

### `row_total_options` Block

* `custom_label` - (Optional) Custom label.
* `metric_header_cell_style` - (Optional) Metric header cell style. See [`metric_header_cell_style` Block](#metric_header_cell_style-block) below.
* `placement` - (Optional) Placement.
* `scroll_status` - (Optional) Scroll status.
* `total_cell_style` - (Optional) Total cell style. See [`total_cell_style` Block](#total_cell_style-block) below.
* `totals_visibility` - (Optional) Totals visibility.
* `value_cell_style` - (Optional) Value cell style. See [`value_cell_style` Block](#value_cell_style-block) below.

### `rows` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `same_sheet_target_visual_configuration` Block

* `target_visual_option` - (Optional) Target visual option.
* `target_visuals` - (Optional) Target visuals.

### `sankey_diagram_aggregated_field_wells` Block

* `destination` - (Optional) Destination. See [`destination` Block](#destination-block) below.
* `source` - (Optional) Source. See [`source` Block](#source-block) below.
* `weight` - (Optional) Weight. See [`weight` Block](#weight-block) below.

### `sankey_diagram_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `scale` Block

* `linear` - (Optional) Linear. See [`linear` Block](#linear-block) below.
* `logarithmic` - (Optional) Logarithmic. See [`logarithmic` Block](#logarithmic-block) below.

### `scatter_plot_categorically_aggregated_field_wells` Block

* `category` - (Optional) Category. See [`category` Block](#category-block) below.
* `size` - (Optional) Size. See [`size` Block](#size-block) below.
* `x_axis` - (Optional) X axis. See [`x_axis` Block](#x_axis-block) below.
* `y_axis` - (Optional) Y axis. See [`y_axis` Block](#y_axis-block) below.

### `scatter_plot_unaggregated_field_wells` Block

* `size` - (Optional) Size. See [`size` Block](#size-block) below.
* `x_axis` - (Optional) X axis. See [`x_axis` Block](#x_axis-block) below.
* `y_axis` - (Optional) Y axis. See [`y_axis` Block](#y_axis-block) below.

### `scatter_plot_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `scenario` Block

* `what_if_point_scenario` - (Optional) What if point scenario. See [`what_if_point_scenario` Block](#what_if_point_scenario-block) below.
* `what_if_range_scenario` - (Optional) What if range scenario. See [`what_if_range_scenario` Block](#what_if_range_scenario-block) below.

### `scope` Block

* `role` - (Optional) Role.

### `scope_configuration` Block

* `selected_sheets` - (Optional) Selected sheets. See [`selected_sheets` Block](#selected_sheets-block) below.

### `screen_canvas_size_options` Block

* `optimized_view_port_width` - (Optional) Optimized view port width.
* `resize_option` - (Required) Resize option.

### `scrollbar_options` Block

* `visibility` - (Optional) Visibility.
* `visible_range` - (Optional) Visible range. See [`visible_range` Block](#visible_range-block) below.

### `search_options` Block

* `visibility` - (Optional) Visibility.

### `secondary_value` Block

* `visibility` - (Optional) Visibility.

### `secondary_value_font_configuration` Block

* `font_color` - (Optional) Font color.
* `font_decoration` - (Optional) Font decoration.
* `font_size` - (Optional) Font size. See [`font_size` Block](#font_size-block) below.
* `font_style` - (Optional) Font style.
* `font_weight` - (Optional) Font weight. See [`font_weight` Block](#font_weight-block) below.

### `secondary_y_axis_display_options` Block

* `axis_line_visibility` - (Optional) Axis line visibility.
* `axis_offset` - (Optional) Axis offset.
* `axis_options` - (Optional) Axis options. See [`axis_options` Block](#axis_options-block) below.
* `data_options` - (Optional) Data options. See [`data_options` Block](#data_options-block) below.
* `grid_line_visibility` - (Optional) Grid line visibility.
* `missing_data_configuration` - (Optional) Missing data configuration. See [`missing_data_configuration` Block](#missing_data_configuration-block) below.
* `scrollbar_options` - (Optional) Scrollbar options. See [`scrollbar_options` Block](#scrollbar_options-block) below.
* `tick_label_options` - (Optional) Tick label options. See [`tick_label_options` Block](#tick_label_options-block) below.

### `secondary_y_axis_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

### `section_based` Block

* `canvas_size_options` - (Required) Canvas size options. See [`canvas_size_options` Block](#canvas_size_options-block) below.

### `section_based_layout` Block

* `body_sections` - (Required) Body sections. See [`body_sections` Block](#body_sections-block) below.
* `canvas_size_options` - (Optional) Canvas size options. See [`canvas_size_options` Block](#canvas_size_options-block) below.
* `footer_sections` - (Required) Footer sections. See [`footer_sections` Block](#footer_sections-block) below.
* `header_sections` - (Required) Header sections. See [`header_sections` Block](#header_sections-block) below.

### `select_all_options` Block

* `visibility` - (Optional) Visibility.

### `selectable_values` Block

* `link_to_data_set_column` - (Optional) Link to data set column. See [`link_to_data_set_column` Block](#link_to_data_set_column-block) below.
* `values` - (Optional) Values.

### `selected_border_style` Block

* `color` - (Optional) Color.
* `visibility` - (Optional) Visibility.

### `selected_field_options` Block

* `custom_label` - (Optional) Custom label.
* `field_id` - (Required) Field ID.
* `url_styling` - (Optional) URL styling. See [`url_styling` Block](#url_styling-block) below.
* `visibility` - (Optional) Visibility.
* `width` - (Optional) Width.

### `selected_fields_configuration` Block

* `selected_field_option` - (Optional) Selected field option.
* `selected_fields` - (Optional) Selected fields.

### `selected_sheets` Block

* `sheet_visual_scoping_configurations` - (Optional) Sheet visual scoping configurations. See [`sheet_visual_scoping_configurations` Block](#sheet_visual_scoping_configurations-block) below.

### `separator_configuration` Block

* `decimal_separator` - (Optional) Decimal separator.
* `thousands_separator` - (Optional) Thousands separator. See [`thousands_separator` Block](#thousands_separator-block) below.

### `series` Block

* `data_field_series_item` - (Optional) Data field series item. See [`data_field_series_item` Block](#data_field_series_item-block) below.
* `field_series_item` - (Optional) Field series item. See [`field_series_item` Block](#field_series_item-block) below.

### `set_parameters_operation` Block

* `parameter_value_configurations` - (Required) Parameter value configurations. See [`parameter_value_configurations` Block](#parameter_value_configurations-block) below.

### `settings` Block

* `line_style_settings` - (Optional) Line style settings. See [`line_style_settings` Block](#line_style_settings-block) below.
* `marker_style_settings` - (Optional) Marker style settings. See [`marker_style_settings` Block](#marker_style_settings-block) below.

### `shape` Block

* `field_id` - (Required) Field ID.
* `format` - (Optional) Format. See [`format` Block](#format-block) below.

### `sheet_control_layouts` Block

* `configuration` - (Required) Configuration. See [`configuration` Block](#configuration-block) below.

### `sheet_visual_scoping_configurations` Block

* `scope` - (Required) Scope.
* `sheet_id` - (Required) Sheet ID.
* `visual_ids` - (Optional) Visual ids.

### `sheets` Block

* `content_type` - (Optional) Content type.
* `description` - (Optional) Description.
* `filter_controls` - (Optional) Filter controls. See [`filter_controls` Block](#filter_controls-block) below.
* `layouts` - (Optional) Layouts. See [`layouts` Block](#layouts-block) below.
* `name` - (Optional) Name.
* `parameter_controls` - (Optional) Parameter controls. See [`parameter_controls` Block](#parameter_controls-block) below.
* `sheet_control_layouts` - (Optional) Sheet control layouts. See [`sheet_control_layouts` Block](#sheet_control_layouts-block) below.
* `sheet_id` - (Required) Sheet ID.
* `text_boxes` - (Optional) Text boxes. See [`text_boxes` Block](#text_boxes-block) below.
* `title` - (Optional) Title.
* `visuals` - (Optional) Visuals. See [`visuals` Block](#visuals-block) below.

### `side_specific_border` Block

* `bottom` - (Required) Bottom. See [`bottom` Block](#bottom-block) below.
* `inner_horizontal` - (Required) Inner horizontal. See [`inner_horizontal` Block](#inner_horizontal-block) below.
* `inner_vertical` - (Required) Inner vertical. See [`inner_vertical` Block](#inner_vertical-block) below.
* `left` - (Required) Left. See [`left` Block](#left-block) below.
* `right` - (Required) Right. See [`right` Block](#right-block) below.
* `top` - (Required) Top. See [`top` Block](#top-block) below.

### `simple_cluster_marker` Block

* `color` - (Optional) Color.

### `size` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `size_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

### `sizes` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `sizing_options` Block

* `table_cell_image_scaling_configuration` - (Optional) Table cell image scaling configuration.

### `slider` Block

* `display_options` - (Optional) Display options. See [`display_options` Block](#display_options-block) below.
* `filter_control_id` - (Required) Filter control ID.
* `maximum_value` - (Required) Maximum value.
* `minimum_value` - (Required) Minimum value.
* `parameter_control_id` - (Required) Parameter control ID.
* `source_filter_id` - (Required) Source filter ID.
* `source_parameter_name` - (Required) Source parameter name.
* `step_size` - (Required) Step size.
* `title` - (Required) Title.
* `type` - (Optional) Type.

### `small_multiples` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `small_multiples_limit_configuration` Block

* `items_limit` - (Optional) Items limit.
* `other_categories` - (Required) Other categories.

### `small_multiples_options` Block

* `max_visible_columns` - (Optional) Max visible columns.
* `max_visible_rows` - (Optional) Max visible rows.
* `panel_configuration` - (Optional) Panel configuration. See [`panel_configuration` Block](#panel_configuration-block) below.

### `small_multiples_sort` Block

* `column_sort` - (Optional) Column sort. See [`column_sort` Block](#column_sort-block) below.
* `field_sort` - (Optional) Field sort. See [`field_sort` Block](#field_sort-block) below.

### `solid` Block

* `color` - (Optional) Color.
* `expression` - (Required) Expression.

### `sort_by` Block

* `column` - (Optional) Column. See [`column` Block](#column-block) below.
* `column_name` - (Required) Column name.
* `data_path` - (Optional) Data path. See [`data_path` Block](#data_path-block) below.
* `data_set_identifier` - (Required) Data set identifier.
* `field` - (Optional) Field. See [`field` Block](#field-block) below.

### `sort_configuration` Block

* `breakdown_items_limit` - (Optional) Breakdown items limit. See [`breakdown_items_limit` Block](#breakdown_items_limit-block) below.
* `category_items_limit` - (Optional) Category items limit. See [`category_items_limit` Block](#category_items_limit-block) below.
* `category_items_limit_configuration` - (Optional) Category items limit configuration. See [`category_items_limit_configuration` Block](#category_items_limit_configuration-block) below.
* `category_sort` - (Optional) Category sort. See [`category_sort` Block](#category_sort-block) below.
* `color_items_limit` - (Optional) Color items limit. See [`color_items_limit` Block](#color_items_limit-block) below.
* `color_items_limit_configuration` - (Optional) Color items limit configuration. See [`color_items_limit_configuration` Block](#color_items_limit_configuration-block) below.
* `color_sort` - (Optional) Color sort. See [`color_sort` Block](#color_sort-block) below.
* `destination_items_limit` - (Optional) Destination items limit. See [`destination_items_limit` Block](#destination_items_limit-block) below.
* `field_sort_options` - (Optional) Field sort options. See [`field_sort_options` Block](#field_sort_options-block) below.
* `heat_map_column_items_limit_configuration` - (Optional) Heat map column items limit configuration. See [`heat_map_column_items_limit_configuration` Block](#heat_map_column_items_limit_configuration-block) below.
* `heat_map_column_sort` - (Optional) Heat map column sort. See [`heat_map_column_sort` Block](#heat_map_column_sort-block) below.
* `heat_map_row_items_limit_configuration` - (Optional) Heat map row items limit configuration. See [`heat_map_row_items_limit_configuration` Block](#heat_map_row_items_limit_configuration-block) below.
* `heat_map_row_sort` - (Optional) Heat map row sort. See [`heat_map_row_sort` Block](#heat_map_row_sort-block) below.
* `pagination_configuration` - (Optional) Pagination configuration. See [`pagination_configuration` Block](#pagination_configuration-block) below.
* `row_sort` - (Optional) Row sort. See [`row_sort` Block](#row_sort-block) below.
* `small_multiples_limit_configuration` - (Optional) Small multiples limit configuration. See [`small_multiples_limit_configuration` Block](#small_multiples_limit_configuration-block) below.
* `small_multiples_sort` - (Optional) Small multiples sort. See [`small_multiples_sort` Block](#small_multiples_sort-block) below.
* `source_items_limit` - (Optional) Source items limit. See [`source_items_limit` Block](#source_items_limit-block) below.
* `tree_map_group_items_limit_configuration` - (Optional) Tree map group items limit configuration. See [`tree_map_group_items_limit_configuration` Block](#tree_map_group_items_limit_configuration-block) below.
* `tree_map_sort` - (Optional) Tree map sort. See [`tree_map_sort` Block](#tree_map_sort-block) below.
* `trend_group_sort` - (Optional) Trend group sort. See [`trend_group_sort` Block](#trend_group_sort-block) below.
* `weight_sort` - (Optional) Weight sort. See [`weight_sort` Block](#weight_sort-block) below.

### `sort_paths` Block

* `field_id` - (Required) Field ID.
* `field_value` - (Required) Field value.

### `source` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `source_controls` Block

* `column_to_match` - (Required) Column to match. See [`column_to_match` Block](#column_to_match-block) below.
* `source_sheet_control_id` - (Optional) Source sheet control ID.

### `source_entity` Block

* `source_template` - (Optional) Source template. See [`source_template` Block](#source_template-block) below.

### `source_items_limit` Block

* `items_limit` - (Optional) Items limit.
* `other_categories` - (Required) Other categories.

### `source_template` Block

* `arn` - (Required) ARN.
* `data_set_references` - (Required) Data set references. See [`data_set_references` Block](#data_set_references-block) below.

### `sparkline` Block

* `color` - (Optional) Color.
* `tooltip_visibility` - (Optional) Tooltip visibility.
* `type` - (Required) Type.
* `visibility` - (Optional) Visibility.

### `standard_layout` Block

* `type` - (Required) Type.

### `static_configuration` Block

* `value` - (Required) Value.

### `stops` Block

* `color` - (Optional) Color.
* `data_value` - (Optional) Data value.
* `gradient_offset` - (Required) Gradient offset.

### `string_format_configuration` Block

* `null_value_format_configuration` - (Optional) Null value format configuration. See [`null_value_format_configuration` Block](#null_value_format_configuration-block) below.
* `numeric_format_configuration` - (Optional) Numeric format configuration. See [`numeric_format_configuration` Block](#numeric_format_configuration-block) below.

### `string_parameter_declaration` Block

* `default_values` - (Optional) Default values. See [`default_values` Block](#default_values-block) below.
* `name` - (Required) Name.
* `parameter_value_type` - (Required) Parameter value type.
* `values_when_unset` - (Optional) Values when unset. See [`values_when_unset` Block](#values_when_unset-block) below.

### `string_parameters` Block

* `name` - (Required) Name.
* `values` - (Required) Values.

### `style` Block

* `height` - (Optional) Height.
* `padding` - (Optional) Padding. See [`padding` Block](#padding-block) below.

### `style_configuration` Block

* `color` - (Optional) Color.
* `pattern` - (Optional) Pattern.

### `style_options` Block

* `fill_style` - (Optional) Fill style.

### `subtitle` Block

* `format_text` - (Optional) Format text. See [`format_text` Block](#format_text-block) below.
* `visibility` - (Optional) Visibility.

### `table_aggregated_field_wells` Block

* `group_by` - (Optional) Group by. See [`group_by` Block](#group_by-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `table_inline_visualizations` Block

* `data_bars` - (Optional) Data bars. See [`data_bars` Block](#data_bars-block) below.

### `table_options` Block

* `cell_style` - (Optional) Cell style. See [`cell_style` Block](#cell_style-block) below.
* `collapsed_row_dimensions_visibility` - (Optional) Collapsed row dimensions visibility.
* `column_header_style` - (Optional) Column header style. See [`column_header_style` Block](#column_header_style-block) below.
* `column_names_visibility` - (Optional) Column names visibility.
* `header_style` - (Optional) Header style. See [`header_style` Block](#header_style-block) below.
* `metric_placement` - (Optional) Metric placement.
* `orientation` - (Optional) Orientation.
* `row_alternate_color_options` - (Optional) Row alternate color options. See [`row_alternate_color_options` Block](#row_alternate_color_options-block) below.
* `row_field_names_style` - (Optional) Row field names style. See [`row_field_names_style` Block](#row_field_names_style-block) below.
* `row_header_style` - (Optional) Row header style. See [`row_header_style` Block](#row_header_style-block) below.
* `single_metric_visibility` - (Optional) Single metric visibility.
* `toggle_buttons_visibility` - (Optional) Toggle buttons visibility.

### `table_unaggregated_field_wells` Block

* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `table_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `conditional_formatting` - (Optional) Conditional formatting. See [`conditional_formatting` Block](#conditional_formatting-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `target_value` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `target_values` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `target_visuals_configuration` Block

* `same_sheet_target_visual_configuration` - (Optional) Same sheet target visual configuration. See [`same_sheet_target_visual_configuration` Block](#same_sheet_target_visual_configuration-block) below.

### `text_area` Block

* `delimiter` - (Optional) Delimiter.
* `display_options` - (Optional) Display options. See [`display_options` Block](#display_options-block) below.
* `filter_control_id` - (Required) Filter control ID.
* `parameter_control_id` - (Required) Parameter control ID.
* `source_filter_id` - (Required) Source filter ID.
* `source_parameter_name` - (Required) Source parameter name.
* `title` - (Required) Title.

### `text_boxes` Block

* `content` - (Optional) Content.
* `sheet_text_box_id` - (Required) Sheet text box ID.

### `text_color` Block

* `gradient` - (Optional) Gradient. See [`gradient` Block](#gradient-block) below.
* `solid` - (Optional) Solid. See [`solid` Block](#solid-block) below.

### `text_field` Block

* `display_options` - (Optional) Display options. See [`display_options` Block](#display_options-block) below.
* `filter_control_id` - (Required) Filter control ID.
* `parameter_control_id` - (Required) Parameter control ID.
* `source_filter_id` - (Required) Source filter ID.
* `source_parameter_name` - (Required) Source parameter name.
* `title` - (Required) Title.

### `text_format` Block

* `background_color` - (Required) Background color. See [`background_color` Block](#background_color-block) below.
* `icon` - (Optional) Icon. See [`icon` Block](#icon-block) below.
* `text_color` - (Required) Text color. See [`text_color` Block](#text_color-block) below.

### `thousands_separator` Block

* `symbol` - (Optional) Symbol.
* `visibility` - (Optional) Visibility.

### `tick_label_options` Block

* `label_options` - (Optional) Label options. See [`label_options` Block](#label_options-block) below.
* `rotation_angle` - (Optional) Rotation angle.

### `time` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `time_equality_filter` Block

* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `filter_id` - (Required) Filter ID.
* `parameter_name` - (Optional) Parameter name.
* `time_granularity` - (Required) Time granularity.
* `value` - (Optional) Value.

### `time_range_filter` Block

* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `exclude_period_configuration` - (Optional) Exclude period configuration. See [`exclude_period_configuration` Block](#exclude_period_configuration-block) below.
* `filter_id` - (Required) Filter ID.
* `include_maximum` - (Optional) Include maximum.
* `include_minimum` - (Optional) Include minimum.
* `null_option` - (Required) Null option.
* `range_maximum` - (Required) Range maximum.
* `range_maximum_value` - (Optional) Range maximum value. See [`range_maximum_value` Block](#range_maximum_value-block) below.
* `range_minimum` - (Required) Range minimum.
* `range_minimum_value` - (Optional) Range minimum value. See [`range_minimum_value` Block](#range_minimum_value-block) below.
* `time_granularity` - (Required) Time granularity.

### `title` Block

* `custom_label` - (Optional) Custom label.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `format_text` - (Optional) Format text. See [`format_text` Block](#format_text-block) below.
* `horizontal_text_alignment` - (Optional) Horizontal text alignment.
* `visibility` - (Optional) Visibility.

### `title_options` Block

* `custom_label` - (Optional) Custom label.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `visibility` - (Optional) Visibility.

### `tooltip` Block

* `field_base_tooltip` - (Optional) Field base tooltip. See [`field_base_tooltip` Block](#field_base_tooltip-block) below.
* `selected_tooltip_type` - (Optional) Selected tooltip type.
* `tooltip_visibility` - (Optional) Tooltip visibility.

### `tooltip_fields` Block

* `column_tooltip_item` - (Optional) Column tooltip item. See [`column_tooltip_item` Block](#column_tooltip_item-block) below.
* `field_tooltip_item` - (Optional) Field tooltip item. See [`field_tooltip_item` Block](#field_tooltip_item-block) below.

### `top` Block

* `color` - (Optional) Color.
* `style` - (Optional) Style.
* `thickness` - (Optional) Thickness.

### `top_bottom_filter` Block

* `aggregation_sort_configuration` - (Required) Aggregation sort configuration. See [`aggregation_sort_configuration` Block](#aggregation_sort_configuration-block) below.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `filter_id` - (Required) Filter ID.
* `limit` - (Optional) Limit.
* `parameter_name` - (Optional) Parameter name.
* `time_granularity` - (Required) Time granularity.

### `top_bottom_movers` Block

* `category` - (Optional) Category. See [`category` Block](#category-block) below.
* `computation_id` - (Required) Computation ID.
* `mover_size` - (Optional) Mover size.
* `name` - (Optional) Name.
* `sort_order` - (Required) Sort order.
* `time` - (Optional) Time. See [`time` Block](#time-block) below.
* `type` - (Required) Type.
* `value` - (Optional) Value. See [`value` Block](#value-block) below.

### `top_bottom_ranked` Block

* `category` - (Optional) Category. See [`category` Block](#category-block) below.
* `computation_id` - (Required) Computation ID.
* `name` - (Optional) Name.
* `result_size` - (Optional) Result size.
* `type` - (Required) Type.
* `value` - (Optional) Value. See [`value` Block](#value-block) below.

### `total_aggregation` Block

* `computation_id` - (Required) Computation ID.
* `name` - (Optional) Name.
* `value` - (Optional) Value. See [`value` Block](#value-block) below.

### `total_cell_style` Block

* `background_color` - (Optional) Background color.
* `border` - (Optional) Border. See [`border` Block](#border-block) below.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `height` - (Optional) Height.
* `horizontal_text_alignment` - (Optional) Horizontal text alignment.
* `text_wrap` - (Optional) Text wrap.
* `vertical_text_alignment` - (Optional) Vertical text alignment.
* `visibility` - (Optional) Visibility.

### `total_options` Block

* `column_subtotal_options` - (Optional) Column subtotal options. See [`column_subtotal_options` Block](#column_subtotal_options-block) below.
* `column_total_options` - (Optional) Column total options. See [`column_total_options` Block](#column_total_options-block) below.
* `custom_label` - (Optional) Custom label.
* `placement` - (Optional) Placement.
* `row_subtotal_options` - (Optional) Row subtotal options. See [`row_subtotal_options` Block](#row_subtotal_options-block) below.
* `row_total_options` - (Optional) Row total options. See [`row_total_options` Block](#row_total_options-block) below.
* `scroll_status` - (Optional) Scroll status.
* `total_cell_style` - (Optional) Total cell style. See [`total_cell_style` Block](#total_cell_style-block) below.
* `totals_visibility` - (Optional) Totals visibility.

### `tree_map_aggregated_field_wells` Block

* `colors` - (Optional) Colors. See [`colors` Block](#colors-block) below.
* `groups` - (Optional) Groups. See [`groups` Block](#groups-block) below.
* `sizes` - (Optional) Sizes. See [`sizes` Block](#sizes-block) below.

### `tree_map_group_items_limit_configuration` Block

* `items_limit` - (Optional) Items limit.
* `other_categories` - (Required) Other categories.

### `tree_map_sort` Block

* `column_sort` - (Optional) Column sort. See [`column_sort` Block](#column_sort-block) below.
* `field_sort` - (Optional) Field sort. See [`field_sort` Block](#field_sort-block) below.

### `tree_map_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `trend_arrows` Block

* `visibility` - (Optional) Visibility.

### `trend_group_sort` Block

* `column_sort` - (Optional) Column sort. See [`column_sort` Block](#column_sort-block) below.
* `field_sort` - (Optional) Field sort. See [`field_sort` Block](#field_sort-block) below.

### `trend_groups` Block

* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.

### `uniform_border` Block

* `color` - (Optional) Color.
* `style` - (Optional) Style.
* `thickness` - (Optional) Thickness.

### `unique_values` Block

* `category` - (Optional) Category. See [`category` Block](#category-block) below.
* `computation_id` - (Required) Computation ID.
* `name` - (Optional) Name.

### `url_operation` Block

* `url_target` - (Required) URL target.
* `url_template` - (Required) URL template.

### `url_styling` Block

* `image_configuration` - (Optional) Image configuration. See [`image_configuration` Block](#image_configuration-block) below.
* `link_configuration` - (Optional) Link configuration. See [`link_configuration` Block](#link_configuration-block) below.

### `user_name_column` Block

* `column_name` - (Required) Column name.
* `data_set_identifier` - (Required) Data set identifier.

### `value` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `custom_values_configuration` - (Optional) Custom values configuration. See [`custom_values_configuration` Block](#custom_values_configuration-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.
* `select_all_value_options` - (Optional) Select all value options.
* `source_field` - (Optional) Source field.
* `source_parameter_name` - (Optional) Source parameter name.

### `value_axis` Block

* `axis_line_visibility` - (Optional) Axis line visibility.
* `axis_offset` - (Optional) Axis offset.
* `data_options` - (Optional) Data options. See [`data_options` Block](#data_options-block) below.
* `grid_line_visibility` - (Optional) Grid line visibility.
* `scrollbar_options` - (Optional) Scrollbar options. See [`scrollbar_options` Block](#scrollbar_options-block) below.
* `tick_label_options` - (Optional) Tick label options. See [`tick_label_options` Block](#tick_label_options-block) below.

### `value_cell_style` Block

* `background_color` - (Optional) Background color.
* `border` - (Optional) Border. See [`border` Block](#border-block) below.
* `font_configuration` - (Optional) Font configuration. See [`font_configuration` Block](#font_configuration-block) below.
* `height` - (Optional) Height.
* `horizontal_text_alignment` - (Optional) Horizontal text alignment.
* `text_wrap` - (Optional) Text wrap.
* `vertical_text_alignment` - (Optional) Vertical text alignment.
* `visibility` - (Optional) Visibility.

### `value_label_configuration` Block

* `format_configuration` - (Optional) Format configuration. See [`format_configuration` Block](#format_configuration-block) below.
* `relative_position` - (Optional) Relative position.

### `value_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

### `values` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `column` - (Required) Column. See [`column` Block](#column-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `field_id` - (Required) Field ID.
* `format_configuration` - (Optional) Format configuration. See [`format_configuration` Block](#format_configuration-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `values_when_unset` Block

* `custom_value` - (Optional) Custom value.
* `value_when_unset_option` - (Optional) Value when unset option.

### `visible_range` Block

* `percent_range` - (Optional) Percent range. See [`percent_range` Block](#percent_range-block) below.

### `visual_layout_options` Block

* `standard_layout` - (Optional) Standard layout. See [`standard_layout` Block](#standard_layout-block) below.

### `visual_palette` Block

* `chart_color` - (Optional) Chart color.
* `color_map` - (Optional) Color map. See [`color_map` Block](#color_map-block) below.

### `visuals` Block

* `bar_chart_visual` - (Optional) Bar chart visual. See [`bar_chart_visual` Block](#bar_chart_visual-block) below.
* `box_plot_visual` - (Optional) Box plot visual. See [`box_plot_visual` Block](#box_plot_visual-block) below.
* `combo_chart_visual` - (Optional) Combo chart visual. See [`combo_chart_visual` Block](#combo_chart_visual-block) below.
* `custom_content_visual` - (Optional) Custom content visual. See [`custom_content_visual` Block](#custom_content_visual-block) below.
* `empty_visual` - (Optional) Empty visual. See [`empty_visual` Block](#empty_visual-block) below.
* `filled_map_visual` - (Optional) Filled map visual. See [`filled_map_visual` Block](#filled_map_visual-block) below.
* `funnel_chart_visual` - (Optional) Funnel chart visual. See [`funnel_chart_visual` Block](#funnel_chart_visual-block) below.
* `gauge_chart_visual` - (Optional) Gauge chart visual. See [`gauge_chart_visual` Block](#gauge_chart_visual-block) below.
* `geospatial_map_visual` - (Optional) Geospatial map visual. See [`geospatial_map_visual` Block](#geospatial_map_visual-block) below.
* `heat_map_visual` - (Optional) Heat map visual. See [`heat_map_visual` Block](#heat_map_visual-block) below.
* `histogram_visual` - (Optional) Histogram visual. See [`histogram_visual` Block](#histogram_visual-block) below.
* `insight_visual` - (Optional) Insight visual. See [`insight_visual` Block](#insight_visual-block) below.
* `kpi_visual` - (Optional) Kpi visual. See [`kpi_visual` Block](#kpi_visual-block) below.
* `line_chart_visual` - (Optional) Line chart visual. See [`line_chart_visual` Block](#line_chart_visual-block) below.
* `pie_chart_visual` - (Optional) Pie chart visual. See [`pie_chart_visual` Block](#pie_chart_visual-block) below.
* `pivot_table_visual` - (Optional) Pivot table visual. See [`pivot_table_visual` Block](#pivot_table_visual-block) below.
* `radar_chart_visual` - (Optional) Radar chart visual. See [`radar_chart_visual` Block](#radar_chart_visual-block) below.
* `sankey_diagram_visual` - (Optional) Sankey diagram visual. See [`sankey_diagram_visual` Block](#sankey_diagram_visual-block) below.
* `scatter_plot_visual` - (Optional) Scatter plot visual. See [`scatter_plot_visual` Block](#scatter_plot_visual-block) below.
* `table_visual` - (Optional) Table visual. See [`table_visual` Block](#table_visual-block) below.
* `tree_map_visual` - (Optional) Tree map visual. See [`tree_map_visual` Block](#tree_map_visual-block) below.
* `waterfall_visual` - (Optional) Waterfall visual. See [`waterfall_visual` Block](#waterfall_visual-block) below.
* `word_cloud_visual` - (Optional) Word cloud visual. See [`word_cloud_visual` Block](#word_cloud_visual-block) below.

### `waterfall_chart_aggregated_field_wells` Block

* `breakdowns` - (Optional) Breakdowns. See [`breakdowns` Block](#breakdowns-block) below.
* `categories` - (Optional) Categories. See [`categories` Block](#categories-block) below.
* `values` - (Optional) Values. See [`values` Block](#values-block) below.

### `waterfall_chart_options` Block

* `total_bar_label` - (Optional) Total bar label.

### `waterfall_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `weight` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `weight_sort` Block

* `column_sort` - (Optional) Column sort. See [`column_sort` Block](#column_sort-block) below.
* `field_sort` - (Optional) Field sort. See [`field_sort` Block](#field_sort-block) below.

### `what_if_point_scenario` Block

* `date` - (Required) Date.
* `value` - (Required) Value.

### `what_if_range_scenario` Block

* `end_date` - (Required) End date.
* `start_date` - (Required) Start date.
* `value` - (Required) Value.

### `window_options` Block

* `bounds` - (Optional) Bounds. See [`bounds` Block](#bounds-block) below.
* `map_zoom_mode` - (Optional) Map zoom mode.

### `word_cloud_aggregated_field_wells` Block

* `group_by` - (Optional) Group by. See [`group_by` Block](#group_by-block) below.
* `size` - (Optional) Size. See [`size` Block](#size-block) below.

### `word_cloud_options` Block

* `cloud_layout` - (Optional) Cloud layout.
* `maximum_string_length` - (Optional) Maximum string length.
* `word_casing` - (Optional) Word casing.
* `word_orientation` - (Optional) Word orientation.
* `word_padding` - (Optional) Word padding.
* `word_scaling` - (Optional) Word scaling.

### `word_cloud_visual` Block

* `actions` - (Optional) Actions. See [`actions` Block](#actions-block) below.
* `chart_configuration` - (Optional) Chart configuration. See [`chart_configuration` Block](#chart_configuration-block) below.
* `column_hierarchies` - (Optional) Column hierarchies. See [`column_hierarchies` Block](#column_hierarchies-block) below.
* `subtitle` - (Optional) Subtitle. See [`subtitle` Block](#subtitle-block) below.
* `title` - (Optional) Title. See [`title` Block](#title-block) below.
* `visual_id` - (Required) Visual ID.

### `x_axis` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `x_axis_display_options` Block

* `axis_line_visibility` - (Optional) Axis line visibility.
* `axis_offset` - (Optional) Axis offset.
* `data_options` - (Optional) Data options. See [`data_options` Block](#data_options-block) below.
* `grid_line_visibility` - (Optional) Grid line visibility.
* `scrollbar_options` - (Optional) Scrollbar options. See [`scrollbar_options` Block](#scrollbar_options-block) below.
* `tick_label_options` - (Optional) Tick label options. See [`tick_label_options` Block](#tick_label_options-block) below.

### `x_axis_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

### `y_axis` Block

* `calculated_measure_field` - (Optional) Calculated measure field. See [`calculated_measure_field` Block](#calculated_measure_field-block) below.
* `categorical_dimension_field` - (Optional) Categorical dimension field. See [`categorical_dimension_field` Block](#categorical_dimension_field-block) below.
* `categorical_measure_field` - (Optional) Categorical measure field. See [`categorical_measure_field` Block](#categorical_measure_field-block) below.
* `date_dimension_field` - (Optional) Date dimension field. See [`date_dimension_field` Block](#date_dimension_field-block) below.
* `date_measure_field` - (Optional) Date measure field. See [`date_measure_field` Block](#date_measure_field-block) below.
* `numerical_dimension_field` - (Optional) Numerical dimension field. See [`numerical_dimension_field` Block](#numerical_dimension_field-block) below.
* `numerical_measure_field` - (Optional) Numerical measure field. See [`numerical_measure_field` Block](#numerical_measure_field-block) below.

### `y_axis_display_options` Block

* `axis_line_visibility` - (Optional) Axis line visibility.
* `axis_offset` - (Optional) Axis offset.
* `data_options` - (Optional) Data options. See [`data_options` Block](#data_options-block) below.
* `grid_line_visibility` - (Optional) Grid line visibility.
* `scrollbar_options` - (Optional) Scrollbar options. See [`scrollbar_options` Block](#scrollbar_options-block) below.
* `tick_label_options` - (Optional) Tick label options. See [`tick_label_options` Block](#tick_label_options-block) below.

### `y_axis_label_options` Block

* `axis_label_options` - (Optional) Axis label options. See [`axis_label_options` Block](#axis_label_options-block) below.
* `sort_icon_visibility` - (Optional) Sort icon visibility.
* `visibility` - (Optional) Visibility.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the analysis.
* `created_time` - Time that the analysis was created.
* `id` - Comma-delimited string joining AWS account ID and analysis ID.
* `last_published_time` - Time that the analysis was last published.
* `last_updated_time` - Time that the analysis was last updated.
* `status` - Analysis creation status.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a QuickSight Analysis using the AWS account ID and analysis ID separated by a comma (`,`). For example:

```terraform
import {
  to = aws_quicksight_analysis.example
  id = "123456789012,example-id"
}
```

Using `terraform import`, import a QuickSight Analysis using the AWS account ID and analysis ID separated by a comma (`,`). For example:

```console
% terraform import aws_quicksight_analysis.example 123456789012,example-id
```
