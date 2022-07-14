---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_usage_limit"
description: |-
  Provides a Redshift Usage Limit resource.
---

# Resource: aws_redshift_usage_limit

Creates a new Amazon Redshift Usage Limit.

## Example Usage

```terraform
resource "aws_redshift_usage_limit" "example" {
  cluster_identifier = aws_redshift_cluster.example.id
  feature_type       = "concurrency-scaling"
  limit_type         = "time"
  amount             = 60
}
```

## Argument Reference

The following arguments are supported:

* `amount` - (Required) The limit amount. If time-based, this amount is in minutes. If data-based, this amount is in terabytes (TB). The value must be a positive number.
* `breach_action` - (Optional) The action that Amazon Redshift takes when the limit is reached. The default is `log`. Valid values are `log`, `emit-metric`, and `disable`.
* `cluster_identifier` - (Required) The identifier of the cluster that you want to limit usage.
* `feature_type` - (Required) The Amazon Redshift feature that you want to limit. Valid values are `spectrum`, `concurrency-scaling`, and `cross-region-datasharing`.
* `limit_type` - (Required) The type of limit. Depending on the feature type, this can be based on a time duration or data size. If FeatureType is `spectrum`, then LimitType must be `data-scanned`. If FeatureType is `concurrency-scaling`, then LimitType must be `time`. If FeatureType is `cross-region-datasharing`, then LimitType must be `data-scanned`. Valid values are `data-scanned`, and `time`.
* `period` - (Optional) The time period that the amount applies to. A weekly period begins on Sunday. The default is `monthly`. Valid values are `daily`, `weekly`, and `monthly`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the Redshift Usage Limit.
* `id` - The Redshift Usage Limit ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Redshift usage limits can be imported using the `id`, e.g.,

```
$ terraform import aws_redshift_usage_limit.example example-id
```
