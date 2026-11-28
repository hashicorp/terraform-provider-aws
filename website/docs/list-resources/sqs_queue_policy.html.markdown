---
subcategory: "SQS (Simple Queue)"
layout: "aws"
page_title: "AWS: aws_sqs_queue_policy"
description: |-
  Lists SQS queue policies for queues in the configured region.
---

# List Resource: aws_sqs_queue_policy

Lists SQS queue policies for queues in the configured region.

## Example Usage

```terraform
list "aws_sqs_queue_policy" "example" {
  provider = aws
}

list "aws_sqs_queue_policy" "example_with_resource_data" {
  provider = aws

  include_resource = true
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
