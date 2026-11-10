---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_dashboard"
description: |-
  Lists CloudWatch Omni Dashboard resources.
---

# List Resource: aws_cloudwatchomni_dashboard

Lists CloudWatch Omni Dashboard resources.

## Example Usage

```terraform
list "aws_cloudwatchomni_dashboard" "example" {
  provider = aws

  config {
    space_id = aws_cloudwatchomni_space.example.space_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `space_id` - (Required) ID of the space to list dashboards from.
