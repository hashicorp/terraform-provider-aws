---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_sql_injection_match_set"
description: |-
  Provides a AWS WAF SQL Injection Match Set resource.
---

# Resource: aws_waf_sql_injection_match_set

Provides a WAF SQL Injection Match Set Resource

## Example Usage

```terraform
resource "aws_waf_sql_injection_match_set" "sql_injection_match_set" {
  name = "tf-sql_injection_match_set"

  sql_injection_match_tuples {
    text_transformation = "URL_DECODE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name or description of the SQL Injection Match Set.
* `sql_injection_match_tuples` - (Optional) The parts of web requests that you want AWS WAF to inspect for malicious SQL code and, if you want AWS WAF to inspect a header, the name of the header.

## Nested Blocks

### `sql_injection_match_tuples`

* `field_to_match` - (Required) Specifies where in a web request to look for snippets of malicious SQL code.
* `text_transformation` - (Required) Text transformations used to eliminate unusual formatting that attackers use in web requests in an effort to bypass AWS WAF.
  If you specify a transformation, AWS WAF performs the transformation on `field_to_match` before inspecting a request for a match.
  e.g., `CMD_LINE`, `HTML_ENTITY_DECODE` or `NONE`.
  See [docs](http://docs.aws.amazon.com/waf/latest/APIReference/API_SqlInjectionMatchTuple.html#WAF-Type-SqlInjectionMatchTuple-TextTransformation)
  for all supported values.

### `field_to_match`

#### Arguments

* `data` - (Optional) When `type` is `HEADER`, enter the name of the header that you want to search, e.g., `User-Agent` or `Referer`.
  If `type` is any other value, omit this field.
* `type` - (Required) The part of the web request that you want AWS WAF to search for a specified string.
  e.g., `HEADER`, `METHOD` or `BODY`.
  See [docs](http://docs.aws.amazon.com/waf/latest/APIReference/API_FieldToMatch.html)
  for all supported values.

## Remarks

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the WAF SQL Injection Match Set.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS WAF SQL Injection Match Set using their ID. For example:

```terraform
import {
  to = aws_waf_sql_injection_match_set.example
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import AWS WAF SQL Injection Match Set using their ID. For example:

```console
% terraform import aws_waf_sql_injection_match_set.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
