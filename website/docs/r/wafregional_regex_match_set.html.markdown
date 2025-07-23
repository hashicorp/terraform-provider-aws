---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_regex_match_set"
description: |-
  Provides a AWS WAF Regional Regex Match Set resource.
---

# Resource: aws_wafregional_regex_match_set

Provides a WAF Regional Regex Match Set Resource

## Example Usage

```terraform
resource "aws_wafregional_regex_match_set" "example" {
  name = "example"

  regex_match_tuple {
    field_to_match {
      data = "User-Agent"
      type = "HEADER"
    }

    regex_pattern_set_id = aws_wafregional_regex_pattern_set.example.id
    text_transformation  = "NONE"
  }
}

resource "aws_wafregional_regex_pattern_set" "example" {
  name                  = "example"
  regex_pattern_strings = ["one", "two"]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name or description of the Regex Match Set.
* `regex_match_tuple` - (Required) The regular expression pattern that you want AWS WAF to search for in web requests, the location in requests that you want AWS WAF to search, and other settings. See below.

### Nested Arguments

#### `regex_match_tuple`

* `field_to_match` - (Required) The part of a web request that you want to search, such as a specified header or a query string.
* `regex_pattern_set_id` - (Required) The ID of a [Regex Pattern Set](/docs/providers/aws/r/waf_regex_pattern_set.html).
* `text_transformation` - (Required) Text transformations used to eliminate unusual formatting that attackers use in web requests in an effort to bypass AWS WAF.
  e.g., `CMD_LINE`, `HTML_ENTITY_DECODE` or `NONE`.
  See [docs](http://docs.aws.amazon.com/waf/latest/APIReference/API_ByteMatchTuple.html#WAF-Type-ByteMatchTuple-TextTransformation)
  for all supported values.

#### `field_to_match`

* `data` - (Optional) When `type` is `HEADER`, enter the name of the header that you want to search, e.g., `User-Agent` or `Referer`.
  If `type` is any other value, omit this field.
* `type` - (Required) The part of the web request that you want AWS WAF to search for a specified string.
  e.g., `HEADER`, `METHOD` or `BODY`.
  See [docs](http://docs.aws.amazon.com/waf/latest/APIReference/API_FieldToMatch.html)
  for all supported values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the WAF Regional Regex Match Set.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Regional Regex Match Set using the id. For example:

```terraform
import {
  to = aws_wafregional_regex_match_set.example
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Regional Regex Match Set using the id. For example:

```console
% terraform import aws_wafregional_regex_match_set.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
