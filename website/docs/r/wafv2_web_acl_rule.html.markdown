---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl_rule"
description: |-
  Manages an AWS WAF Web ACL Rule.
---

# Resource: aws_wafv2_web_acl_rule

Manages an AWS WAF Web ACL Rule.

## Example Usage

### Basic Usage

```terraform
resource "aws_wafv2_web_acl_rule" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Web ACL Rule.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Web ACL Rule using the `example_id_arg`. For example:

```terraform
import {
  to = aws_wafv2_web_acl_rule.example
  id = "web_acl_rule-id-12345678"
}
```

Using `terraform import`, import WAF Web ACL Rule using the `example_id_arg`. For example:

```console
% terraform import aws_wafv2_web_acl_rule.example web_acl_rule-id-12345678
```
