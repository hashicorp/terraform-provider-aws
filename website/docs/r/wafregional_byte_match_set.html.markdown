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

* `byte_match_tuples` - (Optional) Settings for the ByteMatchSet, such as the bytes (typically a string that corresponds with ASCII characters) that you want AWS WAF to search for in web requests. See [`byte_match_tuples` Block](#byte_match_tuples-block) below for details.
* `name` - (Required) Name or description of the ByteMatchSet.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `byte_match_tuples` Block

The `byte_match_tuples` block supports the following arguments:

* `field_to_match` - (Required) Settings for the ByteMatchTuple. See [`field_to_match` Block](#field_to_match-block) below for details.
* `positional_constraint` - (Required) Within the portion of a web request that you want to search.
* `target_string` - (Optional) Value that you want AWS WAF to search for. The maximum length of the value is 50 bytes.
* `text_transformation` - (Required) Formatting way for web request.

### `field_to_match` Block

The `field_to_match` block supports the following arguments:

* `data` - (Optional) When the value of Type is HEADER, enter the name of the header that you want AWS WAF to search, for example, User-Agent or Referer. If the value of Type is any other value, omit Data.
* `type` - (Required) Part of the web request that you want AWS WAF to search for a specified string.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the WAF ByteMatchSet.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Regional Byte Match Set using the ID. For example:

```terraform
import {
  to = aws_wafregional_byte_match_set.byte_set
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Regional Byte Match Set using the ID. For example:

```console
% terraform import aws_wafregional_byte_match_set.byte_set a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
