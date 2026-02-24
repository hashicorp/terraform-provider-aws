---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_topic"
description: |-
  Terraform resource for managing an AWS Managed Streaming for Kafka Topic.
---
# Resource: aws_msk_topic

Terraform resource for managing an AWS Managed Streaming for Kafka Topic.

## Example Usage

```terraform
resource "aws_msk_topic" "example" {
  cluster_arn        = aws_msk_cluster.example.arn
  topic_name         = "example-topic"
  partition_count    = 3
  replication_factor = 3
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_arn` - (Required, Forces new resource) The ARN of the MSK cluster.
* `topic_name` - (Required, Forces new resource) The name of the Kafka topic.
* `partition_count` - (Required) The number of partitions for the topic. Can only be increased.
* `replication_factor` - (Required, Forces new resource) The replication factor for the topic.
* `configs` - (Optional) Base64-encoded Kafka topic configuration string.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the topic.
* `id` - Composite ID of the topic, formatted as `cluster_arn,topic_name`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MSK Topics using `cluster_arn,topic_name`. For example:

```terraform
import {
  to = aws_msk_topic.example
  id = "arn:aws:kafka:us-east-1:123456789012:cluster/example/uuid,my-topic"
}
```

Using `terraform import`, import MSK Topics using `cluster_arn,topic_name`. For example:

```console
% terraform import aws_msk_topic.example arn:aws:kafka:us-east-1:123456789012:cluster/example/uuid,my-topic
```
