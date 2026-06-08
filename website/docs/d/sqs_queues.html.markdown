---
subcategory: "SQS (Simple Queue)"
layout: "aws"
page_title: "AWS: aws_sqs_queues"
description: |-
  Terraform data source for managing an AWS SQS (Simple Queue) Queues.
---

# Data Source: aws_sqs_queues

Terraform data source for managing an AWS SQS (Simple Queue) Queues.

Backed by the AWS SQS [`ListQueues`](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ListQueues.html) API.

## Example Usage

### Basic Usage

```terraform
data "aws_sqs_queues" "example" {
  queue_name_prefix = "example"
}
```

### Tuning Page Size

The data source paginates through all results automatically. `max_results` controls the page size passed to the AWS `ListQueues` API; it does not cap the total number of queues returned.

```terraform
data "aws_sqs_queues" "example" {
  queue_name_prefix = "example"
  max_results       = 500
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `queue_name_prefix` - (Optional) A string to use for filtering the list results. Only those queues whose name begins with the specified string are returned. Queue URLs and names are case-sensitive.
* `max_results` - (Optional) Maximum number of queue URLs returned per [`ListQueues`](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ListQueues.html) API call (page size). Valid values are between `1` and `1000`. Defaults to `1000`. This setting only affects paging behaviour; all matching queue URLs are returned across pages regardless of the value.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `queue_urls` - A list of queue URLs.
