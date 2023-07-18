---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_regex_pattern_set"
description: |-
  Provides a AWS WAF Regional Regex Pattern Set resource.
---

# Resource: aws_wafregional_regex_pattern_set

Provides a WAF Regional Regex Pattern Set Resource

## Example Usage

```terraform
resource "aws_wafregional_regex_pattern_set" "example" {
  name                  = "example"
  regex_pattern_strings = ["one", "two"]
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name or description of the Regex Pattern Set.
* `regex_pattern_strings` - (Optional) A list of regular expression (regex) patterns that you want AWS WAF to search for, such as `B[a@]dB[o0]t`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the WAF Regional Regex Pattern Set.

## Import

Import WAF Regional Regex Pattern Set using the id. For example:

```
$ terraform import aws_wafregional_regex_pattern_set.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
