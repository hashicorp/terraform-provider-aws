---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_network_performance_metric_subscription"
description: |-
  Provides a resource to manage an Infrastructure Performance subscription.
---

# Resource: aws_vpc_network_performance_metric_subscription

Provides a resource to manage an Infrastructure Performance subscription.

## Example Usage

```terraform
resource "aws_vpc_network_performance_metric_subscription" "example" {
  source      = "us-east-1"
  destination = "us-west-1"
}
```

## Argument Reference

This resource supports the following arguments:

* `destination` - (Required) The target Region or Availability Zone that the metric subscription is enabled for. For example, `eu-west-1`.
* `metric` - (Optional) The metric used for the enabled subscription. Valid values: `aggregate-latency`. Default: `aggregate-latency`.
* `source` - (Required) The source Region or Availability Zone that the metric subscription is enabled for. For example, `us-east-1`.
* `statistic` - (Optional) The statistic used for the enabled subscription. Valid values: `p50`. Default: `p50`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `period` - The data aggregation time for the subscription.
