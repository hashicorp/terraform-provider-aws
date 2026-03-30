---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_topic"
description: |-
  Manages an AWS Managed Streaming for Kafka Topic.
---

# Resource: aws_msk_topic

Manages an AWS Managed Streaming for Kafka Topic.

## Example Usage

### Basic Usage

```terraform
resource "aws_msk_topic" "example" {
  name        = "Example"
  cluster_arn = aws_msk_cluster.example.arn

  partition_count    = 2
  replication_factor = 2

  configs = jsonencode({
    "retention.ms"        = "604800000"
    "retention.bytes"     = "-1",
    "cleanup.policy"      = "delete",
    "min.insync.replicas" = "2"
  })
}
```

## Argument Reference

The following arguments are required:

* `cluster_arn` - (Required) Amazon Resource Name (ARN) that uniquely identifies MSK Cluster.
* `name` - (Required) Name of Topic.
* `partition_count` - (Required) Number of partitions for Topic.
* `replication_factor` - (Required) Replication factor for Topic.

The following arguments are optional:

* `configs` - (Optional) Explicit configured Kafka configuration in JSON format for Topic.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Topic.
* `configs_actual` - Aggregated Kafka configuration in JSON format for Topic, both explicit set values from `configs` and implicit set values (AWS default configuration, historically set values or manual configuration from outside Terraform).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_msk_topic.example
  identity = {
    name        = "Example"
    cluster_arn = "arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3"
  }
}

resource "aws_msk_topic" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `cluster_arn` (String) Amazon Resource Name (ARN) that uniquely identifies MSK Cluster.
* `name` (String) Name of Topic.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Managed Streaming for Kafka Topic using the `cluster_arn` and `name` . For example:

```terraform
import {
  to = aws_kafka_topic.example
  id = "arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3,topicname"
}
```

Using `terraform import`, import Managed Streaming for Kafka Topic using the `cluster_arn` and `name`. For example:

```console
% terraform import aws_kafka_topic.example arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3,topicname
```
