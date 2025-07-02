---
subcategory: "Kinesis Firehose"
layout: "aws"
page_title: "AWS: aws_kinesis_firehose_delivery_stream"
description: |-
  Provides an AWS Kinesis Firehose Delivery Stream data source.
---

# Data Source: aws_kinesis_firehose_delivery_stream

Use this data source to get information about a Kinesis Firehose Delivery Stream for use in other resources.

For more details, see the [Amazon Kinesis Firehose Documentation][1].

## Example Usage

```terraform
data "aws_kinesis_firehose_delivery_stream" "stream" {
  name = "stream-name"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the Kinesis Firehose Delivery Stream.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ARN of the Kinesis Firehose Delivery Stream.
* `arn` - ARN of the Kinesis Firehose Delivery Stream (same as `id`).

[1]: https://aws.amazon.com/documentation/firehose/
