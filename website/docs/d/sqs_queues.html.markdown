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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `queue_name_prefix` - (Optional) A string to use for filtering the list results. Only those queues whose name begins with the specified string are returned. Queue URLs and names are case-sensitive.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `queue_urls` - A list of queue URLs.
