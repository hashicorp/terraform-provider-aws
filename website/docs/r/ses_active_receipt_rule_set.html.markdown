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

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import active SES receipt rule sets using the rule set name. For example:

```terraform
import {
  to = aws_ses_active_receipt_rule_set.my_rule_set
  id = "my_rule_set_name"
}
```

Using `terraform import`, import active SES receipt rule sets using the rule set name. For example:

```console
% terraform import aws_ses_active_receipt_rule_set.my_rule_set my_rule_set_name
```
