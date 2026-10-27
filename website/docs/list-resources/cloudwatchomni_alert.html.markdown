---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_alert"
description: |-
  Lists CloudWatch Omni Alert resources.
---

# List Resource: aws_cloudwatchomni_alert

Lists CloudWatch Omni Alert resources.

## Example Usage

```terraform
list "aws_cloudwatchomni_alert" "example" {
  provider = aws

  config {
    space_id = aws_cloudwatchomni_space.example.space_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `space_id` - (Required) ID of the space to list alerts from.
