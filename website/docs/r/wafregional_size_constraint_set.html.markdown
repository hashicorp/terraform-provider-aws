---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_size_constraint_set"
description: |-
  Provides an AWS WAF Regional Size Constraint Set resource for use with ALB.
---

# Resource: aws_wafregional_size_constraint_set

Provides a WAF Regional Size Constraint Set Resource for use with Application Load Balancer.

## Example Usage

```terraform
resource "aws_wafregional_size_constraint_set" "size_constraint_set" {
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
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `size_constraints` - (Optional) Parts of web requests that you want to inspect the size of. See [`size_constraints` Block](#size_constraints-block) below for details.

### `size_constraints` Block

The `size_constraints` block supports the following arguments:

* `comparison_operator` - (Required) Type of comparison you want to perform. For example, `EQ`, `NE`, `LT`, `GT`. See [docs](https://docs.aws.amazon.com/waf/latest/APIReference/API_wafRegional_SizeConstraint.html) for all supported values.
* `field_to_match` - (Required) Where in a web request to look for the size constraint. See [`field_to_match` Block](#field_to_match-block) below for details.
* `size` - (Required) Size in bytes that you want to compare against the size of the specified `field_to_match`. Valid values are between 0 - 21474836480 bytes (0 - 20 GB).
* `text_transformation` - (Required) Text transformations used to eliminate unusual formatting that attackers use in web requests in an effort to bypass AWS WAF. If you specify a transformation, AWS WAF performs the transformation on `field_to_match` before inspecting a request for a match. For example, `CMD_LINE`, `HTML_ENTITY_DECODE` or `NONE`. See [docs](http://docs.aws.amazon.com/waf/latest/APIReference/API_SizeConstraint.html#WAF-Type-SizeConstraint-TextTransformation) for all supported values. **Note:** if you choose `BODY` as `type`, you must choose `NONE` because CloudFront forwards only the first 8192 bytes for inspection.

### `field_to_match` Block

The `field_to_match` block supports the following arguments:

* `data` - (Optional) When `type` is `HEADER`, enter the name of the header that you want to search, e.g., `User-Agent` or `Referer`. If `type` is any other value, omit this field.
* `type` - (Required) Part of the web request that you want AWS WAF to search for a specified string. For example, `HEADER`, `METHOD` or `BODY`. See [docs](http://docs.aws.amazon.com/waf/latest/APIReference/API_FieldToMatch.html) for all supported values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the WAF Size Constraint Set.
* `id` - ID of the WAF Size Constraint Set.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Size Constraint Set using the ID. For example:

```terraform
import {
  to = aws_wafregional_size_constraint_set.size_constraint_set
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Size Constraint Set using the ID. For example:

```console
% terraform import aws_wafregional_size_constraint_set.size_constraint_set a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
