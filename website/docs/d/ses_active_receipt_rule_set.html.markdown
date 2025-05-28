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

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - SES receipt rule set ARN.
* `rule_set_name` - Name of the rule set
