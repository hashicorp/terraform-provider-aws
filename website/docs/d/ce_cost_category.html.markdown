---
subcategory: "CE (Cost Explorer)"
layout: "aws"
page_title: "AWS: aws_ce_cost_category"
description: |-
  Provides details about a specific CostExplorer Cost Category Definition
---

# Resource: aws_ce_cost_category

Provides details about a specific CostExplorer Cost Category.

## Example Usage

```terraform
data "aws_ce_cost_category" "example" {
  cost_category_arn = "costCategoryARN"
}
```

## Argument Reference

The following arguments are required:

* `cost_category_arn` - (Required) Unique name for the Cost Category.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the cost category.
* `effective_end` - Effective end data of your Cost Category.
* `effective_start` - Effective state data of your Cost Category.
* `id` - Unique ID of the cost category.
* `rule` - Configuration block for the Cost Category rules used to categorize costs. See below.
* `rule_version` - Rule schema version in this particular Cost Category.
* `split_charge_rule` - Configuration block for the split charge rules used to allocate your charges between your Cost Category values. See below.

### `rule`

* `inherited_value` - Configuration block for the value the line item is categorized as if the line item contains the matched dimension. See below.
* `rule` - Configuration block for the `Expression` object used to categorize costs. See below.
* `type` - You can define the CostCategoryRule rule type as either `REGULAR` or `INHERITED_VALUE`.
* `value` - Default value for the cost category.

### `inherited_value`

* `dimension_key` - Key to extract cost category values.
* `dimension_name` - Name of the dimension that's used to group costs. If you specify `LINKED_ACCOUNT_NAME`, the cost category value is based on account name. If you specify `TAG`, the cost category value will be based on the value of the specified tag key. Valid values are `LINKED_ACCOUNT_NAME`, `TAG`

### `rule`

* `and` - Return results that match both `Dimension` objects.
* `cost_category` - Configuration block for the filter that's based on `CostCategory` values. See below.
* `dimension` - Configuration block for the specific `Dimension` to use for `Expression`. See below.
* `not` - Return results that match both `Dimension` object.
* `or` - Return results that match both `Dimension` object.
* `tag` - Configuration block for the specific `Tag` to use for `Expression`. See below.

### `cost_category`

* `key` - Unique name of the Cost Category.
* `match_options` - Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - Specific value of the Cost Category.

### `dimension`

* `key` - Unique name of the Cost Category.
* `match_options` - Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - Specific value of the Cost Category.

### `tag`

* `key` - Key for the tag.
* `match_options` - Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - Specific value of the Cost Category.

### `split_charge_rule`

* `method` - Method that's used to define how to split your source costs across your targets. Valid values are `FIXED`, `PROPORTIONAL`, `EVEN`
* `parameter` - Configuration block for the parameters for a split charge method. This is only required for the `FIXED` method. See below.
* `source` - Cost Category value that you want to split.
* `targets` - Cost Category values that you want to split costs across. These values can't be used as a source in other split charge rules.

### `parameter`

* `type` - Parameter type.
* `values` - Parameter values.
