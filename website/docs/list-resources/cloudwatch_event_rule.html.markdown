---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_rule"
description: |-
  Lists EventBridge Rule resources.
---

# List Resource: aws_cloudwatch_event_rule

Lists EventBridge Rule resources.

## Example Usage

```terraform
list "aws_cloudwatch_event_rule" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
