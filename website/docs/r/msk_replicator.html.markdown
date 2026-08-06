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

### Self-Managed Apache Kafka Cluster Target

Replicate from an Amazon MSK cluster to a self-managed or on-premises Apache Kafka cluster, authenticating to the Apache Kafka cluster with SASL/SCRAM and trusting a custom root CA chain.

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
    apache_kafka_cluster {
      apache_kafka_cluster_id = "target-apache-kafka-cluster"
      bootstrap_broker_string = "b-1.example.com:9096,b-2.example.com:9096"
    }

    client_authentication {
      sasl_scram {
        mechanism  = "SHA512"
        secret_arn = aws_secretsmanager_secret.target.arn
      }
    }

    encryption_in_transit {
      root_ca_certificate = aws_secretsmanager_secret.root_ca.arn
    }
  }

  replication_info_list {
    source_kafka_cluster_arn = aws_msk_cluster.source.arn
    target_kafka_cluster_id  = "target-apache-kafka-cluster"
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

### With Log Delivery

Deliver replicator logs to CloudWatch Logs, Amazon Data Firehose, and Amazon S3.

```terraform
resource "aws_msk_replicator" "test" {
  replicator_name            = "test-name"
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

  log_delivery {
    replicator_log_delivery {
      cloudwatch_logs {
        enabled   = true
        log_group = aws_cloudwatch_log_group.test.name
      }

      firehose {
        enabled         = true
        delivery_stream = aws_kinesis_firehose_delivery_stream.test.name
      }

      s3 {
        enabled = true
        bucket  = aws_s3_bucket.test.bucket
        prefix  = "replicator-logs"
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `replicator_name` - (Required) The name of the replicator.
* `kafka_cluster` - (Required) The source and target Kafka clusters for the replicator. Exactly two blocks are required. Detailed below.
* `service_execution_role_arn` - (Required) The ARN of the IAM role used by the replicator to access resources in the customer's account (e.g source and target clusters).
* `replication_info_list` - (Required) A list of replication configurations, where each configuration targets a given source cluster to target cluster replication flow.
* `description` - (Optional) A summary description of the replicator.
* `log_delivery` - (Optional) Configuration block for delivering replicator logs to customer destinations. Detailed below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### kafka_cluster Argument Reference

* `amazon_msk_cluster` - (Optional) Details of an Amazon MSK cluster. Exactly one of `amazon_msk_cluster` or `apache_kafka_cluster` must be specified. Detailed below.
* `apache_kafka_cluster` - (Optional) Details of a self-managed or on-premises Apache Kafka cluster. Exactly one of `amazon_msk_cluster` or `apache_kafka_cluster` must be specified. Detailed below.
* `client_authentication` - (Optional) Details of the client authentication used by the Kafka cluster. Only valid for an `apache_kafka_cluster`. Detailed below.
* `encryption_in_transit` - (Optional) Details of encryption in transit to the Kafka cluster. Only valid for an `apache_kafka_cluster`. TLS encryption in transit is always applied to an `apache_kafka_cluster`; this block is only required to supply a custom root CA chain (for a cluster using a private or self-signed certificate). Detailed below.
* `vpc_config` - (Optional) Details of an Amazon VPC which has network connectivity to the Kafka cluster. Provide this on the `amazon_msk_cluster` entry only; the replicator reaches the Apache Kafka cluster through that VPC.

### amazon_msk_cluster Argument Reference

* `msk_cluster_arn` - (Required) The ARN of an Amazon MSK cluster.

### apache_kafka_cluster Argument Reference

* `apache_kafka_cluster_id` - (Required) The Kafka `cluster.id` of the self-managed or on-premises Apache Kafka cluster (as reported by the cluster itself, e.g. via the Kafka admin tooling), not an arbitrary name. MSK Replicator validates this value against the source cluster. See [Migrate third-party and self-managed Apache Kafka clusters to Amazon MSK](https://aws.amazon.com/blogs/big-data/migrate-third-party-and-self-managed-apache-kafka-clusters-to-amazon-msk-express-and-standard-brokers-with-amazon-msk-replicator/) for how to obtain the cluster ID and the other required inputs.
* `bootstrap_broker_string` - (Required) The bootstrap broker connection string used to connect to the Apache Kafka cluster.

### client_authentication Argument Reference

* `mtls` - (Optional) Details of the mTLS client authentication used by the Kafka cluster. Detailed below.
* `sasl_scram` - (Optional) Details of the SASL/SCRAM client authentication used by the Kafka cluster. Detailed below.

### mtls Argument Reference

* `secret_arn` - (Required) The ARN of the AWS Secrets Manager secret that stores the private key and certificate used for mTLS authentication. See [Set up prerequisites for MSK Replicator with self-managed Apache Kafka clusters](https://docs.aws.amazon.com/msk/latest/developerguide/msk-replicator-external-prereqs.html) for the required secret contents and format.

### sasl_scram Argument Reference

* `mechanism` - (Required) The SASL/SCRAM mechanism used for authentication. Valid values are `SHA256` and `SHA512`.
* `secret_arn` - (Required) The ARN of the AWS Secrets Manager secret that stores the credentials used for SASL/SCRAM authentication. See [Set up prerequisites for MSK Replicator with self-managed Apache Kafka clusters](https://docs.aws.amazon.com/msk/latest/developerguide/msk-replicator-external-prereqs.html) for the required secret contents and format.

### encryption_in_transit Argument Reference

* `root_ca_certificate` - (Required) The ARN of the AWS Secrets Manager secret that stores the custom root CA certificate chain used to trust the certificate authority of the Apache Kafka cluster. See [Set up prerequisites for MSK Replicator with self-managed Apache Kafka clusters](https://docs.aws.amazon.com/msk/latest/developerguide/msk-replicator-external-prereqs.html) for the required secret contents and format.

### vpc_config Argument Reference

* `subnet_ids` - (Required) The list of subnets to connect to in the virtual private cloud (VPC). AWS creates elastic network interfaces inside these subnets to allow communication between your Kafka Cluster and the replicator.
* `security_groups_ids` - (Required) The AWS security groups to associate with the ENIs used by the replicator. If a security group is not specified, the default security group associated with the VPC is used.

~> **Note:** When an `apache_kafka_cluster` uses `client_authentication`, the replicator's network interfaces (created in these subnets, with private IPs only) must be able to reach AWS Secrets Manager and AWS KMS to retrieve and decrypt the credentials. Ensure the subnets have egress to those services via a NAT gateway or Secrets Manager and KMS interface VPC endpoints; otherwise the replicator times out connecting to the source cluster.

### replication_info_list Argument Reference

* `source_kafka_cluster_arn` - (Optional) The ARN of the source Kafka cluster. Use for an Amazon MSK source. Exactly one of `source_kafka_cluster_arn` or `source_kafka_cluster_id` must be specified.
* `source_kafka_cluster_id` - (Optional) The identifier of the source Kafka cluster. Use for a self-managed / on-premises Apache Kafka source (matches `apache_kafka_cluster_id`). Exactly one of `source_kafka_cluster_arn` or `source_kafka_cluster_id` must be specified.
* `target_kafka_cluster_arn` - (Optional) The ARN of the target Kafka cluster. Use for an Amazon MSK target. Exactly one of `target_kafka_cluster_arn` or `target_kafka_cluster_id` must be specified.
* `target_kafka_cluster_id` - (Optional) The identifier of the target Kafka cluster. Use for a self-managed / on-premises Apache Kafka target (matches `apache_kafka_cluster_id`). Exactly one of `target_kafka_cluster_arn` or `target_kafka_cluster_id` must be specified.
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
* `consumer_group_offset_sync_mode` - (Optional) Consumer group offset synchronization mode. Valid values are `LEGACY` and `ENHANCED`. With `LEGACY`, offsets are synchronized when producers write to the source cluster. With `ENHANCED`, consumer offsets are synchronized regardless of producer location. `ENHANCED` requires a corresponding replicator that replicates data from the target cluster to the source cluster and requires `topic_name_configuration.type` to be set to `IDENTICAL`. Defaults to `LEGACY`. Changing this value will force a new resource.
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
