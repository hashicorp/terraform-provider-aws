---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_group"
description: |-
  Lists Auto Scaling Group resources.
---

# List Resource: aws_autoscaling_group

Lists Auto Scaling Group resources.

## Example Usage

```terraform
list "aws_autoscaling_group" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
