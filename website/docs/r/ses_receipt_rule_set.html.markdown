---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_receipt_rule_set"
description: |-
  Provides an SES receipt rule set resource
---

# Resource: aws_ses_receipt_rule_set

Provides an SES receipt rule set resource.

## Example Usage

```terraform
resource "aws_ses_receipt_rule_set" "main" {
  rule_set_name = "primary-rules"
}
```

## Argument Reference

The following arguments are supported:

* `rule_set_name` - (Required) Name of the rule set.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - SES receipt rule set ARN.
* `id` - SES receipt rule set name.

## Import

SES receipt rule sets can be imported using the rule set name.

```
$ terraform import aws_ses_receipt_rule_set.my_rule_set my_rule_set_name
```
