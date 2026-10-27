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

* `destination` - (Required) Target Region or Availability Zone that the metric subscription is enabled for. For example, `eu-west-1`.
* `metric` - (Optional) Metric used for the enabled subscription. Valid values: `aggregate-latency`. Default: `aggregate-latency`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `source` - (Required) Source Region or Availability Zone that the metric subscription is enabled for. For example, `us-east-1`.
* `statistic` - (Optional) Statistic used for the enabled subscription. Valid values: `p50`. Default: `p50`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `period` - Data aggregation time for the subscription.
