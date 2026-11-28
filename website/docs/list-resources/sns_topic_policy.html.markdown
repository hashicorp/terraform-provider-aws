---
subcategory: "SNS (Simple Notification)"
layout: "aws"
page_title: "AWS: aws_sns_topic_policy"
description: |-
  Lists SNS (Simple Notification) Topic Policy resources.
---

# List Resource: aws_sns_topic_policy

Lists SNS (Simple Notification) Topic Policy resources.

## Example Usage

```terraform
list "aws_sns_topic_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
