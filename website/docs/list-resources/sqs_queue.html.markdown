---
subcategory: "SQS (Simple Queue)"
layout: "aws"
page_title: "AWS: aws_sqs_queue"
description: |-
  Lists SQS Queue resources.
---

# List Resource: aws_sqs_queue

Lists SQS Queue resources.

## Example Usage

```terraform
list "aws_sqs_queue" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
