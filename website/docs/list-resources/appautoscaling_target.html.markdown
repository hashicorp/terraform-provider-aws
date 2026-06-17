---
subcategory: "Application Auto Scaling"
layout: "aws"
page_title: "AWS: aws_appautoscaling_target"
description: |-
  Lists Application Auto Scaling scalable targets.
---

# List Resource: aws_appautoscaling_target

Lists Application Auto Scaling scalable targets in the specified service namespace.

## Example Usage

```terraform
list "aws_appautoscaling_target" "example" {
  provider = aws

  config {
    service_namespace = "ecs"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `service_namespace` - (Required) Namespace of the AWS service that owns the scalable target. Valid values are documented in the AWS [`DescribeScalableTargets`](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_DescribeScalableTargets.html) API reference.
* `region` - (Optional) Region to query. Defaults to provider region.
