---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_cluster"
description: |-
  Get information on an Amazon MSK Cluster
---

# Data Source: aws_msk_serverless_cluster

Get information on an Amazon MSK Serverless Cluster.

-> **Note:** This data sources returns information on _serverless_ clusters.

## Example Usage

```terraform
data "aws_msk_serverless_cluster" "example" {
  cluster_name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `cluster_name` - (Required) Name of the cluster.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the MSK cluster.
* `bootstrap_brokers_sasl_iam` - The DNS name (or IP addresses) and SASL IAM port pair. For example, `b-1.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9098`
* `cluster_uuid` - UUID of the MSK cluster, for use in IAM policies.
* `tags` - Map of key-value pairs assigned to the cluster.
