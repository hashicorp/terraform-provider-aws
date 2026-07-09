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
      topic_name_configuration {
        type = "PREFIXED_WITH_SOURCE_CLUSTER_ALIAS"
      }
      topics_to_replicate = [".*"]
      starting_position {
        type = "LATEST"
      }
    }

    consumer_group_replication {
      consumer_groups_to_replicate = [".*"]
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `replicator_name` - (Required) The name of the replicator.
* `kafka_cluster` - (Required) A list of Kafka clusters which are targets of the replicator.
* `service_execution_role_arn` - (Required) The ARN of the IAM role used by the replicator to access resources in the customer's account (e.g source and target clusters).
* `replication_info_list` - (Required) A list of replication configurations, where each configuration targets a given source cluster to target cluster replication flow.
* `description` - (Optional) A summary description of the replicator.
* `log_delivery` - (Optional) Configuration block for delivering replicator logs to customer destinations. Detailed below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

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
* `consumer_group_replication` - (Required) Configuration relating to consumer group replication.

### topic_replication Argument Reference

* `topic_name_configuration` - (Optional) Configuration for specifying replicated topic names should be the same as their corresponding upstream topics or prefixed with source cluster alias.
* `topics_to_replicate` - (Required) List of regular expression patterns indicating the topics to copy.
* `topics_to_exclude` - (Optional) List of regular expression patterns indicating the topics that should not be replica.
* `detect_and_copy_new_topics` - (Optional) Whether to periodically check for new topics and partitions.
* `copy_access_control_lists_for_topics` - (Optional) Whether to periodically configure remote topic ACLs to match their corresponding upstream topics.
* `copy_topic_configurations` - (Optional) Whether to periodically configure remote topics to match their corresponding upstream topics.
* `starting_position` - (Optional) Configuration for specifying the position in the topics to start replicating from.

### consumer_group_replication Argument Reference

* `consumer_groups_to_replicate` - (Required) List of regular expression patterns indicating the consumer groups to copy.
* `consumer_group_offset_sync_mode` - (Optional) The consumer group offset synchronization mode. Valid values are `LEGACY` and `ENHANCED`. With `LEGACY`, offsets are synchronized when producers write to the source cluster. With `ENHANCED`, consumer offsets are synchronized regardless of producer location. `ENHANCED` requires a corresponding replicator that replicates data from the target cluster to the source cluster and requires `topic_name_configuration` to be set to `IDENTICAL`. Defaults to `LEGACY`. Changing this value will force a new resource.
* `consumer_groups_to_exclude` - (Optional) List of regular expression patterns indicating the consumer groups that should not be replicated.
* `detect_and_copy_new_consumer_groups` - (Optional) Whether to periodically check for new consumer groups.
* `synchronise_consumer_group_offsets` - (Optional) Whether to periodically write the translated offsets to __consumer_offsets topic in target cluster.

### topic_name_configuration

* `type` - (optional) The type of topic configuration name. Supports `PREFIXED_WITH_SOURCE_CLUSTER_ALIAS` and `IDENTICAL`.

### starting_position

* `type` - (Optional) The type of replication starting position. Supports `LATEST` and `EARLIEST`.

### log_delivery

* `replicator_log_delivery` - (Optional) Configuration block for replicator log delivery. Detailed below.

### replicator_log_delivery

* `cloudwatch_logs` - (Optional) Configuration block for replicator log delivery to Amazon CloudWatch Logs. Detailed below.
* `firehose` - (Optional) Configuration block for replicator log delivery to Amazon Data Firehose. Detailed below.
* `s3` - (Optional) Configuration block for replicator log delivery to Amazon S3. Detailed below.

### cloudwatch_logs

* `enabled` - (Required) Boolean whether to enable log delivery to CloudWatch Logs.
* `log_group` - (Optional) Name of CloudWatch Logs log group. Required if `enabled` is `true`. If `enabled` is `false`, this value must not be set.

### firehose

* `enabled` - (Required) Boolean whether to enable log delivery to Firehose.
* `delivery_stream` - (Optional) Name of the Firehose delivery stream. Required if `enabled` is `true`. If `enabled` is `false`, this value must not be set.

### s3

* `enabled` - (Required) Boolean whether to enable log delivery to S3.
* `bucket` - (Optional) Name of the S3 bucket. Required if `enabled` is `true`. If `enabled` is `false`, this value must not be set.
* `prefix` - (Optional) Prefix to use when storing replicator logs in S3. If `enabled` is `false`, this value must not be set.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Replicator.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_msk_replicator.example
  identity = {
    arn = "arn:aws:kafka:us-west-2:123456789012:replicator/example-replicator/b3a16098-f408-4995-8e36-482db4f1b46b"
  }
}

resource "aws_msk_replicator" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the MSK replicator.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MSK replicators using `arn`. For example:

```terraform
import {
  to = aws_msk_replicator.example
  id = "arn:aws:kafka:us-west-2:123456789012:replicator/example-replicator/b3a16098-f408-4995-8e36-482db4f1b46b"
}
```

Using `terraform import`, import MSK replicators using `arn`. For example:

```console
% terraform import aws_msk_replicator.example arn:aws:kafka:us-west-2:123456789012:replicator/example-replicator/b3a16098-f408-4995-8e36-482db4f1b46b
```
