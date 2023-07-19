---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_active_receipt_rule_set"
description: |-
  Provides a resource to designate the active SES receipt rule set
---

# Resource: aws_ses_active_receipt_rule_set

Provides a resource to designate the active SES receipt rule set

## Example Usage

```terraform
resource "aws_ses_active_receipt_rule_set" "main" {
  rule_set_name = "primary-rules"
}
```

## Argument Reference

This resource supports the following arguments:

* `rule_set_name` - (Required) The name of the rule set

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The SES receipt rule set name.
* `arn` - The SES receipt rule set ARN.

## Import

Import Active SES receipt rule sets using the rule set name. For example:

```
$ terraform import aws_ses_active_receipt_rule_set.my_rule_set my_rule_set_name
```
