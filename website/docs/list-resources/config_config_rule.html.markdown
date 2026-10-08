---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_config_rule"
description: |-
  Lists AWS Config Rule resources.
---

# List Resource: aws_config_config_rule

Lists AWS Config Rule resources.

## Example Usage

```terraform
list "aws_config_config_rule" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
