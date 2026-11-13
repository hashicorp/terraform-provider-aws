---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_automation_rule_v2"
description: |-
  Lists Security Hub Automation RuleV2 resources.
---

# List Resource: aws_securityhub_automation_rule_v2

Lists Security Hub Automation RuleV2 resources.

## Example Usage

```terraform
list "aws_securityhub_automation_rule_v2" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
