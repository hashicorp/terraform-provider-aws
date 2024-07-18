---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_replicator"
description: |-
  Terraform resource for managing an AWS Managed Streaming for Kafka Replicator.
---

# Resource: aws_msk_replicator

Terraform resource for managing an AWS Managed Streaming for Kafka Replicator.

## Example Usage

### Basic Usage

```terraform
resource "aws_msk_replicator" "test" {
  replicator_name            = "test-name"
  description                = "test-description"
  service_execution_role_arn = aws_iam_role.source.arn

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.source.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.source[*].id
      security_groups_ids = [aws_security_group.source.id]
    }
  }

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.target.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.target[*].id
      security_groups_ids = [aws_security_group.target.id]
    }
  }

  replication_info_list {
    source_kafka_cluster_arn = aws_msk_cluster.source.arn
    target_kafka_cluster_arn = aws_msk_cluster.target.arn
    target_compression_type  = "NONE"


    topic_replication {
      topics_to_replicate = [".*"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = [".*"]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `replicator_name` - (Required) The name of the replicator.
* `kafka_cluster` - (Required) A list of Kafka clusters which are targets of the replicator.
* `service_execution_role_arn` - (Required) The ARN of the IAM role used by the replicator to access resources in the customer's account (e.g source and target clusters).
* `replication_info_list` - (Required) A list of replication configurations, where each configuration targets a given source cluster to target cluster replication flow.
* `description` - (Optional) A summary description of the replicator.

### kafka_cluster Argument Reference

* `amazon_msk_cluster` - (Required) Details of an Amazon MSK cluster.
* `vpc_config` - (Required) Details of an Amazon VPC which has network connectivity to the Apache Kafka cluster.

### amazon_msk_cluster Argument Reference

* `msk_cluster_arn` - (Required) The ARN of an Amazon MSK cluster.

### vpc_config Argument Reference

* `subnet_ids` - (Required) The list of subnets to connect to in the virtual private cloud (VPC). AWS creates elastic network interfaces inside these subnets to allow communication between your Kafka Cluster and the replicator.
* `security_groups_ids` - (Required) The AWS security groups to associate with the ENIs used by the replicator. If a security group is not specified, the default security group associated with the VPC is used.

### replication_info_list Argument Reference

* `source_kafka_cluster_arn` - (Required) The ARN of the source Kafka cluster.
* `target_kafka_cluster_arn` - (Required) The ARN of the target Kafka cluster.
* `target_compression_type` - (Required) The type of compression to use writing records to target Kafka cluster.
* `topic_replication` - (Required) Configuration relating to topic replication.
* `starting_position` - (Optional) Configuration for specifying the position in the topics to start replicating from.
* `consumer_group_replication` - (Required) Configuration relating to consumer group replication.

### topic_replication Argument Reference

* `topics_to_replicate` - (Required) List of regular expression patterns indicating the topics to copy.
* `topics_to_exclude` - (Optional) List of regular expression patterns indicating the topics that should not be replica.
* `detect_and_copy_new_topics` - (Optional) Whether to periodically check for new topics and partitions.
* `copy_access_control_lists_for_topics` - (Optional) Whether to periodically configure remote topic ACLs to match their corresponding upstream topics.
* `copy_topic_configurations` - (Optional) Whether to periodically configure remote topics to match their corresponding upstream topics.

### consumer_group_replication Argument Reference

* `consumer_groups_to_replicate` - (Required) List of regular expression patterns indicating the consumer groups to copy.
* `consumer_groups_to_exclude` - (Optional) List of regular expression patterns indicating the consumer groups that should not be replicated.
* `detect_and_copy_new_consumer_groups` - (Optional) Whether to periodically check for new consumer groups.
* `synchronise_consumer_group_offsets` - (Optional) Whether to periodically write the translated offsets to __consumer_offsets topic in target cluster.

### starting_position

* `type` - (Optional) The type of replication starting position. Supports `LATEST` and `EARLIEST`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Replicator. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MSK relicators using the replicator ARN. For example:

```terraform
import {
  to = aws_msk_replicator.example
  id = "arn:aws:kafka:us-west-2:123456789012:configuration/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3"
}
```

Using `terraform import`, import MSK replicators using the replicator ARN. For example:

```console
% terraform import aws_msk_replicator.example arn:aws:kafka:us-west-2:123456789012:configuration/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3
```
