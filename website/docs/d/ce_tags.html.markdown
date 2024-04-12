---
subcategory: "CE (Cost Explorer)"
layout: "aws"
page_title: "AWS: aws_ce_tags"
description: |-
  Provides the available cost allocation tag keys and tag values for a specified period.
---

# Data source: aws_ce_tags

Provides the available cost allocation tag keys and tag values for a specified period.

## Example Usage

```terraform
data "aws_ce_tags" "test" {
  time_period {
    start = "2021-01-01"
    end   = "2022-12-01"
  }
}
```

## Argument Reference

The following arguments are required:

* `time_period` - (Required) Configuration block for the start and end dates for retrieving the dimension values. See [`time_period` block](#time_period-block) below for details.

The following arguments are optional:

* `filter` - (Optional) Configuration block for the `Expression` object used to categorize costs. See [`filter` block](#filter-block) below for details.
* `search_string` - (Optional) Value that you want to search for.
* `sort_by` - (Optional) Configuration block for the value by which you want to sort the data. [`sort_by` block](#sort_by-block) below for details.
* `tag_key` - (Optional) Key of the tag that you want to return values for.

### `filter` block

The `filter` configuration block supports the following arguments:

* `and` - (Optional) Return results that match both `Dimension` objects.
* `cost_category` - (Optional) Configuration block for the filter that's based on `CostCategory` values. See [`cost_category` block](#cost_category-block) below for details.
* `dimension` - (Optional) Configuration block for the specific `Dimension` to use for `Expression`. See [`dimension` block](#dimension-block) below for details.
* `not` - (Optional) Return results that match both `Dimension` object.
* `or` - (Optional) Return results that match both `Dimension` object.
* `tag` - (Optional) Configuration block for the specific `Tag` to use for `Expression`. See [`tag` block](#tag-block) below for details.

#### `cost_category` block

The `cost_category` configuration block supports the following arguments:

* `key` - (Optional) Unique name of the Cost Category.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

#### `dimension` block

The `dimension` configuration block supports the following arguments:

* `key` - (Optional) Unique name of the Cost Category.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

#### `tag` block

The `tag` configuration block supports the following arguments:

* `key` - (Optional) Key for the tag.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

### `time_period` block

The `time_period` configuration block supports the following arguments:

* `start` - (Required) End of the time period.
* `end` - (Required) Beginning of the time period.

### `sort_by` block

The `sort_by` configuration block supports the following arguments:

* `key` - (Required) key that's used to sort the data. Valid values are: `BlendedCost`,  `UnblendedCost`, `AmortizedCost`, `NetAmortizedCost`, `NetUnblendedCost`, `UsageQuantity`, `NormalizedUsageAmount`.
* `sort_order` - (Optional) order that's used to sort the data. Valid values are: `ASCENDING`,  `DESCENDING`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Unique ID of the tag.
* `tags` - Tags that match your request.
