---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_topic"
description: |-
  Get information on an Amazon MSK Topic
---

# Data Source: aws_msk_topic

Get information on an Amazon MSK Topic.

## Example Usage

```terraform
data "aws_msk_topic" "example" {
  cluster_arn = aws_msk_cluster.example.arn
  name        = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `cluster_arn` - (Required) ARN of the MSK cluster.
* `name` - (Required) Name of the MSK topic.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the MSK topic.
* `configs` - Aggregated Kafka configuration in JSON format for the topic.
* `partition_count` - Number of partitions for the topic.
* `replication_factor` - Replication factor for the topic.
