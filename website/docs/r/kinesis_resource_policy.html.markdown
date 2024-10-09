---
subcategory: "Kinesis"
layout: "aws"
page_title: "AWS: aws_kinesis_resource_policy"
description: |-
  Provides a resource to manage an Amazon Kinesis Streams resource policy.
---

# Resource: aws_kinesis_resource_policy

Provides a resource to manage an Amazon Kinesis Streams resource policy.
Use a resource policy to manage cross-account access to your data streams or consumers.

## Example Usage

```terraform
resource "aws_kinesis_resource_policy" "example" {
  resource_arn = aws_kinesis_stream.example.arn

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "writePolicy",
  "Statement": [{
    "Sid": "writestatement",
    "Effect": "Allow",
    "Principal": {
      "AWS": "123456789456"
    },
    "Action": [
      "kinesis:DescribeStreamSummary",
      "kinesis:ListShards",
      "kinesis:PutRecord",
      "kinesis:PutRecords"
    ],
    "Resource": "${aws_kinesis_stream.example.arn}"
  }]
}
EOF
}
```

## Argument Reference

This resource supports the following arguments:

* `policy` - (Required) The policy document.
* `resource_arn` - (Required) The Amazon Resource Name (ARN) of the data stream or consumer.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Kinesis resource policies using the `resource_arn`. For example:

```terraform
import {
  to = aws_kinesis_resource_policy.example
  id = "arn:aws:kinesis:us-west-2:123456789012:stream/example"
}
```

Using `terraform import`, import Kinesis resource policies using the `resource_arn`. For example:

```console
% terraform import aws_kinesis_resource_policy.example arn:aws:kinesis:us-west-2:123456789012:stream/example
```
