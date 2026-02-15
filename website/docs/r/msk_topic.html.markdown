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

  configs = base64encode(jsonencode({
    "retention.ms"        = "604800000"
    "retention.bytes"     = "-1",
    "cleanup.policy"      = "delete",
    "min.insync.replicas" = "2"
  }))
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of Topic.
* `cluster_arn` - (Required) Amazon Resource Name (ARN) that uniquely identifies MSK Cluster.
* `replication_factor` - (Required) Replication factor for Topic.
* `partition_count` - (Required) Number of partitions for Topic.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `configs` - (Optional) Base64-encoded Kafka configuration in JSON format for Topic.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Topic.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

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

* `name` (String) Name of Topic.
* `cluster_arn` (String) Amazon Resource Name (ARN) that uniquely identifies MSK Cluster.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Managed Streaming for Kafka Topic using the `name` and `cluster_arn`. For example:

```terraform
import {
  to = aws_kafka_topic.example
  id = "topicname,arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3"
}
```

Using `terraform import`, import Managed Streaming for Kafka Topic using the `name`and `cluster_arn`. For example:

```console
% terraform import aws_kafka_topic.example topicname,arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3
```
