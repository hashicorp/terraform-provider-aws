---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_active_receipt_rule_set"
description: |-
  Retrieve the active SES receipt rule set
---

# Data Source: aws_ses_active_receipt_rule_set

Retrieve the active SES receipt rule set

## Example Usage

```terraform
data "aws_ses_active_receipt_rule_set" "main" {}
```

## Attributes Reference

The following attributes are exported:

* `arn` - SES receipt rule set ARN.
* `rule_set_name` - Name of the rule set
