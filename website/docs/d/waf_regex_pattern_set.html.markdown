---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_regex_pattern_set"
description: |-
  Retrieves the summary of a AWS WAF Regex Pattern Set.
---

# Resource: aws_waf_regex_pattern_set

Retrieves the summary of a AWS WAF Regex Pattern Set

## Example Usage

```terraform
data "aws_waf_regex_pattern_set" "example" {
  name = "tf_waf_regex_pattern_set"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) The name of the WAF Regex Pattern Set.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the WAF Regex Pattern Set.
* `arn` - Amazon Resource Name (ARN)
* `regex_pattern_strings` - A list of regular expression (regex) patterns that AWS WAF is searching for, such as `B[a@]dB[o0]t`.
