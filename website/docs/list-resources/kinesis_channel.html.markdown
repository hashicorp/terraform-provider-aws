---
subcategory: "Kinesis"
layout: "aws"
page_title: "AWS: aws_kinesis_channel"
description: |-
  Lists Kinesis Channel resources.
---

# List Resource: aws_kinesis_channel

Lists Kinesis Channel resources.

## Example Usage

```terraform
list "aws_kinesis_channel" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
