---
subcategory: "CE (Cost Explorer)"
layout: "aws"
page_title: "AWS: aws_ce_tags"
description: |-
  Provides details about a specific CE Tags
---

# Resource: aws_ce_tags

Provides details about a specific CE Tags.

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

* `time_period` - (Required) Configuration block for the start and end dates for retrieving the dimension values.

The following arguments are optional:

* `filter` - (Optional) Configuration block for the `Expression` object used to categorize costs. See below.
* `search_string` - (Optional) Value that you want to search for.
* `sort_by` - (Optional) Configuration block for the value by which you want to sort the data. See below.
* `tag_key` - (Optional) Key of the tag that you want to return values for.


### `time_period`

* `start` - (Required) End of the time period.
* `end` - (Required) Beginning of the time period.

### `filter`

* `and` - (Optional) Return results that match both `Dimension` objects.
* `cost_category` - (Optional) Configuration block for the filter that's based on `CostCategory` values. See below.
* `dimension` - (Optional) Configuration block for the specific `Dimension` to use for `Expression`. See below.
* `not` - (Optional) Return results that match both `Dimension` object.
* `or` - (Optional) Return results that match both `Dimension` object.
* `tag` - (Optional) Configuration block for the specific `Tag` to use for `Expression`. See below.

### `cost_category`

* `key` - (Optional) Unique name of the Cost Category.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

### `dimension`

* `key` - (Optional) Unique name of the Cost Category.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

### `tag`

* `key` - (Optional) Key for the tag.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.


### `sort_by`

* `key` - (Required) key that's used to sort the data. Valid values are: `BlendedCost`,  `UnblendedCost`, `AmortizedCost`, `NetAmortizedCost`, `NetUnblendedCost`, `UsageQuantity`, `NormalizedUsageAmount`.
* `sort_order` - (Optional) order that's used to sort the data. Valid values are: `ASCENDING`,  `DESCENDING`.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique ID of the tag.
* `tags` - Tags that match your request.

