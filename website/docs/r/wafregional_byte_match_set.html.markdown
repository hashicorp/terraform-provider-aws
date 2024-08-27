---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_byte_match_set"
description: |-
  Provides a AWS WAF Regional ByteMatchSet resource for use with ALB.
---

# Resource: aws_wafregional_byte_match_set

Provides a WAF Regional Byte Match Set Resource for use with Application Load Balancer.

## Example Usage

```terraform
resource "aws_wafregional_byte_match_set" "byte_set" {
  name = "tf_waf_byte_match_set"

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer1"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name or description of the ByteMatchSet.
* `byte_match_tuples` - (Optional)Settings for the ByteMatchSet, such as the bytes (typically a string that corresponds with ASCII characters) that you want AWS WAF to search for in web requests. ByteMatchTuple documented below.

ByteMatchTuples(byte_match_tuples) support the following:

* `field_to_match` - (Required) Settings for the ByteMatchTuple. FieldToMatch documented below.
* `positional_constraint` - (Required) Within the portion of a web request that you want to search.
* `target_string` - (Required) The value that you want AWS WAF to search for. The maximum length of the value is 50 bytes.
* `text_transformation` - (Required) The formatting way for web request.

FieldToMatch(field_to_match) support following:

* `data` - (Optional) When the value of Type is HEADER, enter the name of the header that you want AWS WAF to search, for example, User-Agent or Referer. If the value of Type is any other value, omit Data.
* `type` - (Required) The part of the web request that you want AWS WAF to search for a specified string.

## Remarks

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the WAF ByteMatchSet.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Regional Byte Match Set using the id. For example:

```terraform
import {
  to = aws_wafregional_byte_match_set.byte_set
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Regional Byte Match Set using the id. For example:

```console
% terraform import aws_wafregional_byte_match_set.byte_set a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
