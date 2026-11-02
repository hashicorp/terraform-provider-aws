---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_theme"
description: |-
  Use this data source to fetch information about a QuickSight Theme.
---

# Data Source: aws_quicksight_theme

Terraform data source for managing an AWS QuickSight Theme.

## Example Usage

### Basic Usage

```terraform
data "aws_quicksight_theme" "example" {
  theme_id = "example"
}
```

## Argument Reference

The following arguments are required:

* `theme_id` - (Required) Identifier of the theme.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the theme.
* `base_theme_id` - ID of the theme that a custom theme will inherit from. All themes inherit from one of the starting themes defined by Amazon QuickSight.
* `configuration` - Theme configuration, which contains the theme display properties. See [configuration](#configuration-block).
* `created_time` - Time that the theme was created.
* `id` - Comma-delimited string joining AWS account ID and theme ID.
* `last_updated_time` - Time that the theme was last updated.
* `name` - Display name of the theme.
* `permissions` - Set of resource permissions on the theme. See [permissions](#permissions-block).
* `status` - Theme creation status.
* `tags` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `version_description` - Description of the current theme version being created/updated.
* `version_number` - Version number of the theme version.

### `permissions` Block

* `actions` - List of IAM actions to grant or revoke permissions on.
* `principal` - ARN of the principal. See the [ResourcePermission documentation](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ResourcePermission.html) for the applicable ARN values.

### `configuration` Block

* `data_color_palette` - Color properties that apply to chart data colors. See [data_color_palette](#data_color_palette-block).
* `sheet` - Display options related to sheets. See [sheet](#sheet-block).
* `typography` - Typography options. See [typography](#typography-block).
* `ui_color_palette` - Color properties that apply to the UI and to charts, excluding the colors that apply to data. See [ui_color_palette](#ui_color_palette-block).

### `data_color_palette` Block

* `colors` - List of hexadecimal codes for the colors. Minimum of 8 items and maximum of 20 items.
* `empty_fill_color` - Hexadecimal code of a color that applies to charts where a lack of data is highlighted.
* `min_max_gradient` - Minimum and maximum hexadecimal codes that describe a color gradient. List of exactly 2 items.

### `sheet` Block

* `tile` - Display options for tiles. See [tile](#tile-block).
* `tile_layout` - Layout options for tiles. See [tile_layout](#tile_layout-block).

### `tile` Block

* `border` - Border around a tile. See [border](#border-block).

### `border` Block

* `show` - Option to enable display of borders for visuals.

### `tile_layout` Block

* `gutter` - Gutter settings that apply between tiles. See [gutter](#gutter-block).
* `margin` - Margin settings that apply around the outside edge of sheets. See [margin](#margin-block).

### `gutter` Block

* `show` - Whether to display a gutter space between sheet tiles.

### `margin` Block

* `show` - Whether to display sheet margins.

### `typography` Block

* `font_families` - List of font families. Maximum number of 5 items. See [font_families](#font_families-block).

### `font_families` Block

* `font_family` - Font family name.

### `ui_color_palette` Block

* `accent` - Color (hexadecimal) that applies to selected states and buttons.
* `accent_foreground` - Color (hexadecimal) that applies to any text or other elements that appear over the accent color.
* `danger` - Color (hexadecimal) that applies to error messages.
* `danger_foreground` - Color (hexadecimal) that applies to any text or other elements that appear over the error color.
* `dimension` - Color (hexadecimal) that applies to the names of fields that are identified as dimensions.
* `dimension_foreground` - Color (hexadecimal) that applies to any text or other elements that appear over the dimension color.
* `measure` - Color (hexadecimal) that applies to the names of fields that are identified as measures.
* `measure_foreground` - Color (hexadecimal) that applies to any text or other elements that appear over the measure color.
* `primary_background` - Color (hexadecimal) that applies to visuals and other high emphasis UI.
* `primary_foreground` - Color (hexadecimal) of text and other foreground elements that appear over the primary background regions, such as grid lines, borders, table banding, icons, and so on.
* `secondary_background` - Color (hexadecimal) that applies to the sheet background and sheet controls.
* `secondary_foreground` - Color (hexadecimal) that applies to any sheet title, sheet control text, or UI that appears over the secondary background.
* `success` - Color (hexadecimal) that applies to success messages, for example the check mark for a successful download.
* `success_foreground` - Color (hexadecimal) that applies to any text or other elements that appear over the success color.
* `warning` - Color (hexadecimal) that applies to warning and informational messages.
* `warning_foreground` - Color (hexadecimal) that applies to any text or other elements that appear over the warning color.
