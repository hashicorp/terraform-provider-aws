---
subcategory: "SES Mail Manager"
layout: "aws"
page_title: "AWS: aws_mailmanager_rule_set"
description: |-
  Lists SES Mail Manager Rule Set resources.
---

# List Resource: aws_mailmanager_rule_set

Lists SES Mail Manager Rule Set resources.

## Example Usage

```terraform
list "aws_mailmanager_rule_set" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
