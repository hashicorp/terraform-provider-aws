---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_regex_pattern_set"
description: |-
  Retrieves the summary of a WAFv2 Regex Pattern Set.
---

# Data Source: aws_wafv2_regex_pattern_set

Retrieves the summary of a WAFv2 Regex Pattern Set.

## Example Usage

```terraform
data "aws_wafv2_regex_pattern_set" "example" {
  name  = "some-regex-pattern-set"
  scope = "REGIONAL"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the WAFv2 Regex Pattern Set.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the entity.
* `description` - Description of the set that helps with identification.
* `id` - Unique identifier for the set.
* `regular_expression` - One or more blocks of regular expression patterns that AWS WAF is searching for. See [Regular Expression](#regular-expression) below for details.

### Regular Expression

Each `regular_expression` supports the following argument:

* `regex_string` - (Required) String representing the regular expression, see the AWS WAF [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-regex-pattern-set-creating.html) for more information.
