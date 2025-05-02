---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_regex_pattern_set"
description: |-
  Provides an AWS WAFv2 Regex Pattern Set resource.
---

# Resource: aws_wafv2_regex_pattern_set

Provides an AWS WAFv2 Regex Pattern Set Resource

## Example Usage

```terraform
resource "aws_wafv2_regex_pattern_set" "example" {
  name        = "example"
  description = "Example regex pattern set"
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  regular_expression {
    regex_string = "two"
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Optional, Forces new resource) A friendly name of the regular expression pattern set. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Optional) A friendly description of the regular expression pattern set.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.
* `regular_expression` - (Optional) One or more blocks of regular expression patterns that you want AWS WAF to search for, such as `B[a@]dB[o0]t`. See [Regular Expression](#regular-expression) below for details. A maximum of 10 `regular_expression` blocks may be specified.
* `tags` - (Optional) An array of key:value pairs to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Regular Expression

* `regex_string` - (Required) The string representing the regular expression, see the AWS WAF [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-regex-pattern-set-creating.html) for more information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A unique identifier for the set.
* `arn` - The Amazon Resource Name (ARN) that identifies the cluster.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 Regex Pattern Sets using `ID/name/scope`. For example:

```terraform
import {
  to = aws_wafv2_regex_pattern_set.example
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc/example/REGIONAL"
}
```

Using `terraform import`, import WAFv2 Regex Pattern Sets using `ID/name/scope`. For example:

```console
% terraform import aws_wafv2_regex_pattern_set.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc/example/REGIONAL
```
