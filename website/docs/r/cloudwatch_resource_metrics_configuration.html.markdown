---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_resource_metrics_configuration"
description: |-
  Manages a CloudWatch resource metrics configuration, which enables detailed metric collection for an AWS resource.
---

# Resource: aws_cloudwatch_resource_metrics_configuration

Manages a CloudWatch resource metrics configuration. Creating a configuration makes CloudWatch collect detailed metrics for the targeted AWS resource.

Each AWS resource can have at most one resource metrics configuration. Detailed metrics incur charges, so limit collection with `metric_selections` when you only need a subset of metrics. See [Amazon CloudWatch pricing](https://aws.amazon.com/cloudwatch/pricing/) for details.

~> **Note:** Not every AWS resource type is eligible for detailed metrics. At the time of writing, only Amazon ElastiCache replication groups are supported. Targeting any other resource type, including ElastiCache serverless caches and cache clusters, returns a `ValidationError` with the message `Unsupported resource type`.

## Example Usage

### Collect All Detailed Metrics

Omit `metric_selections` to collect every detailed metric that is available for the resource.

```terraform
resource "aws_elasticache_replication_group" "example" {
  replication_group_id = "example"
  description          = "example"
  engine               = "valkey"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 1
}

resource "aws_cloudwatch_resource_metrics_configuration" "example" {
  resource_arn = aws_elasticache_replication_group.example.arn
}
```

### Collect Selected Metrics

```terraform
resource "aws_cloudwatch_resource_metrics_configuration" "example" {
  resource_arn = aws_elasticache_replication_group.example.arn

  metric_selections {
    include_metrics = [
      "CacheHits",
      "CacheMisses",
      "CurrConnections",
    ]
  }
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` - (Required) ARN of the AWS resource to enable detailed monitoring for. Changing this forces a new resource to be created.

The following arguments are optional:

* `metric_selections` - (Optional) Configuration block that limits which metrics CloudWatch collects. At most one block is supported. When omitted, CloudWatch collects all available detailed metrics for the resource. See [`metric_selections`](#metric_selections) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `metric_selections` Block

* `include_metrics` - (Required) Set of metric names to collect for the resource. CloudWatch collects only the metrics listed here. Between 1 and 500 names, each between 1 and 255 characters. For the metrics published by ElastiCache, see [Monitoring use with CloudWatch Metrics](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/CacheMetrics.html).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_resource_metrics_configuration.example
  identity = {
    resource_arn = "arn:aws:elasticache:us-west-2:012345678901:replicationgroup:example"
  }
}

resource "aws_cloudwatch_resource_metrics_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `resource_arn` - ARN of the AWS resource that the resource metrics configuration applies to.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a CloudWatch resource metrics configuration using the `resource_arn`. For example:

```terraform
import {
  to = aws_cloudwatch_resource_metrics_configuration.example
  id = "arn:aws:elasticache:us-west-2:012345678901:replicationgroup:example"
}
```

Using `terraform import`, import a CloudWatch resource metrics configuration using the `resource_arn`. For example:

```console
% terraform import aws_cloudwatch_resource_metrics_configuration.example arn:aws:elasticache:us-west-2:012345678901:replicationgroup:example
```
