---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_topic"
description: |-
  Get information on an Amazon MSK Topic.
---
# Data Source: aws_msk_topic

Get information on an Amazon MSK Topic.

## Example Usage

```terraform
data "aws_msk_topic" "example" {
  cluster_arn = aws_msk_cluster.example.arn
  topic_name  = "example-topic"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_arn` - (Required) The ARN of the MSK cluster.
* `topic_name` - (Required) The name of the Kafka topic.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the topic.
* `configs` - Base64-encoded Kafka topic configuration string.
* `partition_count` - The number of partitions for the topic.
* `replication_factor` - The replication factor for the topic.
