---
subcategory: "Kinesis Firehose"
layout: "aws"
page_title: "AWS: aws_kinesis_firehose_delivery_stream"
description: |-
  Lists Kinesis Firehose Delivery Stream resources.
---

# List Resource: aws_kinesis_firehose_delivery_stream

Lists Kinesis Firehose Delivery Stream resources.

## Example Usage

```terraform
list "aws_kinesis_firehose_delivery_stream" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
