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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
