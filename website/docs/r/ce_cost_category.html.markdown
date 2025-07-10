---
subcategory: "CE (Cost Explorer)"
layout: "aws"
page_title: "AWS: aws_ce_cost_category"
description: |-
  Provides a CE Cost Category Definition
---

# Resource: aws_ce_cost_category

Provides a CE Cost Category.

## Example Usage

```terraform
resource "aws_ce_cost_category" "test" {
  name         = "NAME"
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-prod"]
        match_options = ["ENDS_WITH"]
      }
    }
  }
  rule {
    value = "staging"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-stg"]
        match_options = ["ENDS_WITH"]
      }
    }
  }
  rule {
    value = "testing"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-dev"]
        match_options = ["ENDS_WITH"]
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Unique name for the Cost Category.
* `rule` - (Required) Configuration block for the Cost Category rules used to categorize costs. See below.
* `rule_version` - (Required) Rule schema version in this particular Cost Category.

The following arguments are optional:

* `default_value` - (Optional) Default value for the cost category.
* `effective_start`- (Optional)  The Cost Category's effective start date. It can only be a billing start date (first day of the month). If the date isn't provided, it's the first day of the current month. Dates can't be before the previous twelve months, or in the future. For example `2022-11-01T00:00:00Z`.
* `split_charge_rule` - (Optional) Configuration block for the split charge rules used to allocate your charges between your Cost Category values. See below.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `rule`

* `inherited_value` - (Optional) Configuration block for the value the line item is categorized as if the line item contains the matched dimension. See below.
* `rule` - (Optional) Configuration block for the `Expression` object used to categorize costs. See below.
* `type` - (Optional) You can define the CostCategoryRule rule type as either `REGULAR` or `INHERITED_VALUE`.
* `value` - (Optional) Default value for the cost category.

### `inherited_value`

* `dimension_key` - (Optional) Key to extract cost category values.
* `dimension_name` - (Optional) Name of the dimension that's used to group costs. If you specify `LINKED_ACCOUNT_NAME`, the cost category value is based on account name. If you specify `TAG`, the cost category value will be based on the value of the specified tag key. Valid values are `LINKED_ACCOUNT_NAME`, `TAG`

### `rule`

* `and` - (Optional) Return results that match both `Dimension` objects.
* `cost_category` - (Optional) Configuration block for the filter that's based on `CostCategory` values. See below.
* `dimension` - (Optional) Configuration block for the specific `Dimension` to use for `Expression`. See below.
* `not` - (Optional) Return results that match both `Dimension` object.
* `or` - (Optional) Return results that match both `Dimension` object.
* `tags` - (Optional) Configuration block for the specific `Tag` to use for `Expression`. See below.

### `cost_category`

* `key` - (Optional) Unique name of the Cost Category.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

### `dimension`

* `key` - (Optional) Unique name of the Cost Category.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

### `tags`

* `key` - (Optional) Key for the tag.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

### `split_charge_rule`

* `method` - (Required) Method that's used to define how to split your source costs across your targets. Valid values are `FIXED`, `PROPORTIONAL`, `EVEN`
* `parameter` - (Optional) Configuration block for the parameters for a split charge method. This is only required for the `FIXED` method. See below.
* `source` - (Required) Cost Category value that you want to split.
* `targets` - (Required) Cost Category values that you want to split costs across. These values can't be used as a source in other split charge rules.

### `parameter`

* `type` - (Optional) Parameter type.
* `values` - (Optional) Parameter values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the cost category.
* `effective_end` - Effective end data of your Cost Category.
* `id` - Unique ID of the cost category.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ce_cost_category` using the id. For example:

```terraform
import {
  to = aws_ce_cost_category.example
  id = "costCategoryARN"
}
```

Using `terraform import`, import `aws_ce_cost_category` using the id. For example:

```console
% terraform import aws_ce_cost_category.example costCategoryARN
```
