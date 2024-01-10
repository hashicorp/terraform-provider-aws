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

This resource supports the following arguments:

* `rule_set_name` - (Required) Name of the rule set.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - SES receipt rule set ARN.
* `id` - SES receipt rule set name.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SES receipt rule sets using the rule set name. For example:

```terraform
import {
  to = aws_ses_receipt_rule_set.my_rule_set
  id = "my_rule_set_name"
}
```

Using `terraform import`, import SES receipt rule sets using the rule set name. For example:

```console
% terraform import aws_ses_receipt_rule_set.my_rule_set my_rule_set_name
```
