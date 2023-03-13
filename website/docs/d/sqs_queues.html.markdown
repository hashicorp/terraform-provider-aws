---
subcategory: "SQS (Simple Queue)"
layout: "aws"
page_title: "AWS: aws_sqs_queues"
description: |-
  Terraform data source for managing an AWS SQS (Simple Queue) Queues.
---

# Data Source: aws_sqs_queues

Terraform data source for managing an AWS SQS (Simple Queue) Queues.

## Example Usage

### Basic Usage

```terraform
data "aws_sqs_queues" "example" {
  queue_name_prefix = "example"
}
```

## Argument Reference

The following arguments are optional:

* `queue_name_prefix` - (Optional) A string to use for filtering the list results. Only those queues whose name begins with the specified string are returned. Queue URLs and names are case-sensitive.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `queue_urls` - A list of queue URLs.
