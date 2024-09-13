---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_size_constraint_set"
description: |-
  The `aws_waf_size_constraint_set` resource provides an AWS WAF Size Constraint Set.
---

# Resource: aws_waf_size_constraint_set

Use the `aws_waf_size_constraint_set` resource to manage WAF size constraint sets.

## Example Usage

```terraform
resource "aws_waf_size_constraint_set" "size_constraint_set" {
  name = "tfsize_constraints"

  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "EQ"
    size                = "4096"

    field_to_match {
      type = "BODY"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name or description of the Size Constraint Set.
* `size_constraints` - (Optional) Parts of web requests that you want to inspect the size of.

## Nested Blocks

### `size_constraints`

#### Arguments

* `field_to_match` - (Required) Parameter that specifies where in a web request to look for the size constraint.
* `comparison_operator` - (Required) Type of comparison you want to perform, such as `EQ`, `NE`, `LT`, or `GT`. Please refer to the [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_wafRegional_SizeConstraint.html) for a complete list of supported values.
* `size` - (Required) Size in bytes that you want to compare against the size of the specified `field_to_match`. Valid values for `size` are between 0 and 21474836480 bytes (0 and 20 GB).
* `text_transformation` - (Required) Parameter is used to eliminate unusual formatting that attackers may use in web requests to bypass AWS WAF. When a transformation is specified, AWS WAF performs the transformation on the `field_to_match` before inspecting the request for a match. Some examples of supported transformations are `CMD_LINE`, `HTML_ENTITY_DECODE`, and `NONE`. You can find a complete list of supported values in the [AWS WAF API Reference](http://docs.aws.amazon.com/waf/latest/APIReference/API_SizeConstraint.html#WAF-Type-SizeConstraint-TextTransformation).
  **Note:** If you choose `BODY` as the `type`, you must also choose `NONE` because CloudFront only forwards the first 8192 bytes for inspection.

### `field_to_match`

#### Arguments

* `data` - (Optional) When the `type` is `HEADER`, specify the name of the header that you want to search using the `data` field, for example, `User-Agent` or `Referer`. If the `type` is any other value, you can omit this field.
* `type` - (Required) Part of the web request that you want AWS WAF to search for a specified string. For example, `HEADER`, `METHOD`, or `BODY`. See the [docs](http://docs.aws.amazon.com/waf/latest/APIReference/API_FieldToMatch.html) for all supported values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF Size Constraint Set.
* `arn` - Amazon Resource Name (ARN).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS WAF Size Constraint Set using their ID. For example:

```terraform
import {
  to = aws_waf_size_constraint_set.example
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import AWS WAF Size Constraint Set using their ID. For example:

```console
% terraform import aws_waf_size_constraint_set.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
