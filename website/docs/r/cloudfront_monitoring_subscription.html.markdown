---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_monitoring_subscription"
description: |-
  Provides a CloudFront monitoring subscription resource.
---

# Resource: aws_cloudfront_monitoring_subscription

Provides a CloudFront real-time log configuration resource.

## Example Usage

```terraform
resource "aws_cloudfront_monitoring_subscription" "example" {
  distribution_id = aws_cloudfront_distribution.example.id

  monitoring_subscription {
    realtime_metrics_subscription_config {
      realtime_metrics_subscription_status = "Enabled"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `distribution_id` - (Required) The ID of the distribution that you are enabling metrics for.
* `monitoring_subscription` - (Required) A monitoring subscription. This structure contains information about whether additional CloudWatch metrics are enabled for a given CloudFront distribution.

### monitoring_subscription

* `realtime_metrics_subscription_config` - (Required) A subscription configuration for additional CloudWatch metrics. See below.

### realtime_metrics_subscription_config

* `realtime_metrics_subscription_status` - (Required) A flag that indicates whether additional CloudWatch metrics are enabled for a given CloudFront distribution. Valid values are `Enabled` and `Disabled`. See below.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the CloudFront monitoring subscription, which corresponds to the `distribution_id`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront monitoring subscription using the id. For example:

```terraform
import {
  to = aws_cloudfront_monitoring_subscription.example
  id = "E3QYSUHO4VYRGB"
}
```

Using `terraform import`, import CloudFront monitoring subscription using the id. For example:

```console
% terraform import aws_cloudfront_monitoring_subscription.example E3QYSUHO4VYRGB
```
