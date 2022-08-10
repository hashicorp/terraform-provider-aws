---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_cluster"
description: |-
  Terraform resource for managing an AWS Managed Streaming for Kafka cluster
---

# Resource: aws_msk_cluster

Manages AWS Managed Streaming for Kafka cluster

## Example Usage

### Basic

```terraform
resource "aws_vpc" "vpc" {
  cidr_block = "192.168.0.0/22"
}

data "aws_availability_zones" "azs" {
  state = "available"
}

resource "aws_subnet" "subnet_az1" {
  availability_zone = data.aws_availability_zones.azs.names[0]
  cidr_block        = "192.168.0.0/24"
  vpc_id            = aws_vpc.vpc.id
}

resource "aws_subnet" "subnet_az2" {
  availability_zone = data.aws_availability_zones.azs.names[1]
  cidr_block        = "192.168.1.0/24"
  vpc_id            = aws_vpc.vpc.id
}

resource "aws_subnet" "subnet_az3" {
  availability_zone = data.aws_availability_zones.azs.names[2]
  cidr_block        = "192.168.2.0/24"
  vpc_id            = aws_vpc.vpc.id
}

resource "aws_security_group" "sg" {
  vpc_id = aws_vpc.vpc.id
}

resource "aws_kms_key" "kms" {
  description = "example"
}

resource "aws_cloudwatch_log_group" "test" {
  name = "msk_broker_logs"
}

resource "aws_s3_bucket" "bucket" {
  bucket = "msk-broker-logs-bucket"
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}

resource "aws_iam_role" "firehose_role" {
  name = "firehose_test_role"

  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
  {
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "firehose.amazonaws.com"
    },
    "Effect": "Allow",
    "Sid": ""
  }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-msk-broker-logs-stream"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose_role.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  tags = {
    LogDeliveryEnabled = "placeholder"
  }

  lifecycle {
    ignore_changes = [
      tags["LogDeliveryEnabled"],
    ]
  }
}

resource "aws_msk_cluster" "example" {
  cluster_name           = "example"
  kafka_version          = "3.2.0"
  number_of_broker_nodes = 3

  broker_node_group_info {
    instance_type = "kafka.m5.large"
    client_subnets = [
      aws_subnet.subnet_az1.id,
      aws_subnet.subnet_az2.id,
      aws_subnet.subnet_az3.id,
    ]
    storage_info {
      ebs_storage_info {
        volume_size = 1000
      }
    }
    security_groups = [aws_security_group.sg.id]
  }

  encryption_info {
    encryption_at_rest_kms_key_arn = aws_kms_key.kms.arn
  }

  open_monitoring {
    prometheus {
      jmx_exporter {
        enabled_in_broker = true
      }
      node_exporter {
        enabled_in_broker = true
      }
    }
  }

  logging_info {
    broker_logs {
      cloudwatch_logs {
        enabled   = true
        log_group = aws_cloudwatch_log_group.test.name
      }
      firehose {
        enabled         = true
        delivery_stream = aws_kinesis_firehose_delivery_stream.test_stream.name
      }
      s3 {
        enabled = true
        bucket  = aws_s3_bucket.bucket.id
        prefix  = "logs/msk-"
      }
    }
  }

  tags = {
    foo = "bar"
  }
}

output "zookeeper_connect_string" {
  value = aws_msk_cluster.example.zookeeper_connect_string
}

output "bootstrap_brokers_tls" {
  description = "TLS connection host:port pairs"
  value       = aws_msk_cluster.example.bootstrap_brokers_tls
}
```

### With volume_throughput argument

```terraform
resource "aws_msk_cluster" "example" {
  cluster_name           = "example"
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    instance_type = "kafka.m5.4xlarge"
    client_subnets = [
      aws_subnet.subnet_az1.id,
      aws_subnet.subnet_az2.id,
      aws_subnet.subnet_az3.id,
    ]
    storage_info {
      ebs_storage_info {
        provisioned_throughput {
          enabled           = true
          volume_throughput = 250
        }
        volume_size = 1000
      }
    }
    security_groups = [aws_security_group.sg.id]
  }
}
```

## Argument Reference

The following arguments are supported:

* `broker_node_group_info` - (Required) Configuration block for the broker nodes of the Kafka cluster.
* `cluster_name` - (Required) Name of the MSK cluster.
* `kafka_version` - (Required) Specify the desired Kafka software version.
* `number_of_broker_nodes` - (Required) The desired total number of broker nodes in the kafka cluster.  It must be a multiple of the number of specified client subnets.
* `client_authentication` - (Optional) Configuration block for specifying a client authentication. See below.
* `configuration_info` - (Optional) Configuration block for specifying a MSK Configuration to attach to Kafka brokers. See below.
* `encryption_info` - (Optional) Configuration block for specifying encryption. See below.
* `enhanced_monitoring` - (Optional) Specify the desired enhanced MSK CloudWatch monitoring level. See [Monitoring Amazon MSK with Amazon CloudWatch](https://docs.aws.amazon.com/msk/latest/developerguide/monitoring.html)
* `open_monitoring` - (Optional) Configuration block for JMX and Node monitoring for the MSK cluster. See below.
* `logging_info` - (Optional) Configuration block for streaming broker logs to Cloudwatch/S3/Kinesis Firehose. See below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### broker_node_group_info Argument Reference

* `client_subnets` - (Required) A list of subnets to connect to in client VPC ([documentation](https://docs.aws.amazon.com/msk/1.0/apireference/clusters.html#clusters-prop-brokernodegroupinfo-clientsubnets)).
* `ebs_volume_size` - (Optional, **Deprecated** use `storage_info.ebs_storage_info.volume_size` instead) The size in GiB of the EBS volume for the data drive on each broker node.
* `instance_type` - (Required) Specify the instance type to use for the kafka brokersE.g., kafka.m5.large. ([Pricing info](https://aws.amazon.com/msk/pricing/))
* `security_groups` - (Required) A list of the security groups to associate with the elastic network interfaces to control who can communicate with the cluster.
* `az_distribution` - (Optional) The distribution of broker nodes across availability zones ([documentation](https://docs.aws.amazon.com/msk/1.0/apireference/clusters.html#clusters-model-brokerazdistribution)). Currently the only valid value is `DEFAULT`.
* `connectivity_info` - (Optional) Information about the cluster access configuration. See below. For security reasons, you can't turn on public access while creating an MSK cluster. However, you can update an existing cluster to make it publicly accessible. You can also create a new cluster and then update it to make it publicly accessible ([documentation](https://docs.aws.amazon.com/msk/latest/developerguide/public-access.html)).
* `storage_info` - (Optional) A block that contains information about storage volumes attached to MSK broker nodes. See below.

### broker_node_group_info connectivity_info Argument Reference

* `public_access` - (Optional) Access control settings for brokers. See below.

### connectivity_info public_access Argument Reference

* `type` - (Optional) Public access type. Valida values: `DISABLED`, `SERVICE_PROVIDED_EIPS`.

### broker_node_group_info storage_info Argument Reference

* `ebs_storage_info` - (Optional) A block that contains EBS volume information. See below.

### storage_info ebs_storage_info Argument Reference

* `provisioned_throughput` - (Optional) A block that contains EBS volume provisioned throughput information. To provision storage throughput, you must choose broker type kafka.m5.4xlarge or larger. See below.
* `volume_size` - (Optional) The size in GiB of the EBS volume for the data drive on each broker node. Minimum value of `1` and maximum value of `16384`.

### ebs_storage_info provisioned_throughput Argument Reference

* `enabled` - (Optional) Controls whether provisioned throughput is enabled or not. Default value: `false`.
* `volume_throughput` - (Optional) Throughput value of the EBS volumes for the data drive on each kafka broker node in MiB per second. The minimum value is `250`. The maximum value varies between broker type. You can refer to the valid values for the maximum volume throughput at the following [documentation on throughput bottlenecks](https://docs.aws.amazon.com/msk/latest/developerguide/msk-provision-throughput.html#throughput-bottlenecks)

### client_authentication Argument Reference

* `sasl` - (Optional) Configuration block for specifying SASL client authentication. See below.
* `tls` - (Optional) Configuration block for specifying TLS client authentication. See below.
* `unauthenticated` - (Optional) Enables unauthenticated access.

#### client_authentication sasl Argument Reference

* `iam` - (Optional) Enables IAM client authentication. Defaults to `false`.
* `scram` - (Optional) Enables SCRAM client authentication via AWS Secrets Manager. Defaults to `false`.

#### client_authentication tls Argument Reference

* `certificate_authority_arns` - (Optional) List of ACM Certificate Authority Amazon Resource Names (ARNs).

### configuration_info Argument Reference

* `arn` - (Required) Amazon Resource Name (ARN) of the MSK Configuration to use in the cluster.
* `revision` - (Required) Revision of the MSK Configuration to use in the cluster.

### encryption_info Argument Reference

* `encryption_in_transit` - (Optional) Configuration block to specify encryption in transit. See below.
* `encryption_at_rest_kms_key_arn` - (Optional) You may specify a KMS key short ID or ARN (it will always output an ARN) to use for encrypting your data at rest.  If no key is specified, an AWS managed KMS ('aws/msk' managed service) key will be used for encrypting the data at rest.

#### encryption_info encryption_in_transit Argument Reference

* `client_broker` - (Optional) Encryption setting for data in transit between clients and brokers. Valid values: `TLS`, `TLS_PLAINTEXT`, and `PLAINTEXT`. Default value is `TLS`.
* `in_cluster` - (Optional) Whether data communication among broker nodes is encrypted. Default value: `true`.

#### open_monitoring Argument Reference

* `prometheus` - (Required) Configuration block for Prometheus settings for open monitoring. See below.

#### open_monitoring prometheus Argument Reference

* `jmx_exporter` - (Optional) Configuration block for JMX Exporter. See below.
* `node_exporter` - (Optional) Configuration block for Node Exporter. See below.

#### open_monitoring prometheus jmx_exporter Argument Reference

* `enabled_in_broker` - (Required) Indicates whether you want to enable or disable the JMX Exporter.

#### open_monitoring prometheus node_exporter Argument Reference

* `enabled_in_broker` - (Required) Indicates whether you want to enable or disable the Node Exporter.

#### logging_info Argument Reference

* `broker_logs` - (Required) Configuration block for Broker Logs settings for logging info. See below.

#### logging_info broker_logs cloudwatch_logs Argument Reference

* `enabled` - (Optional) Indicates whether you want to enable or disable streaming broker logs to Cloudwatch Logs.
* `log_group` - (Optional) Name of the Cloudwatch Log Group to deliver logs to.

#### logging_info broker_logs firehose Argument Reference

* `enabled` - (Optional) Indicates whether you want to enable or disable streaming broker logs to Kinesis Data Firehose.
* `delivery_stream` - (Optional) Name of the Kinesis Data Firehose delivery stream to deliver logs to.

#### logging_info broker_logs s3 Argument Reference

* `enabled` - (Optional) Indicates whether you want to enable or disable streaming broker logs to S3.
* `bucket` - (Optional) Name of the S3 bucket to deliver logs to.
* `prefix` - (Optional) Prefix to append to the folder name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the MSK cluster.
* `bootstrap_brokers` - Comma separated list of one or more hostname:port pairs of kafka brokers suitable to bootstrap connectivity to the kafka cluster. Contains a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `PLAINTEXT` or `TLS_PLAINTEXT`. The resource sorts values alphabetically. AWS may not always return all endpoints so this value is not guaranteed to be stable across applies.
* `bootstrap_brokers_public_sasl_iam` - One or more DNS names (or IP addresses) and SASL IAM port pairs. For example, `b-1-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9198,b-2-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9198,b-3-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9198`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS` and `client_authentication.0.sasl.0.iam` is set to `true` and `broker_node_group_info.0.connectivity_info.0.public_access.0.type` is set to `SERVICE_PROVIDED_EIPS` and the cluster fulfill all other requirements for public access. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `bootstrap_brokers_public_sasl_scram` - One or more DNS names (or IP addresses) and SASL SCRAM port pairs. For example, `b-1-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9196,b-2-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9196,b-3-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9196`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS` and `client_authentication.0.sasl.0.scram` is set to `true` and `broker_node_group_info.0.connectivity_info.0.public_access.0.type` is set to `SERVICE_PROVIDED_EIPS` and the cluster fulfill all other requirements for public access. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `bootstrap_brokers_public_tls` - One or more DNS names (or IP addresses) and TLS port pairs. For example, `b-1-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9194,b-2-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9194,b-3-public.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9194`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS` and `broker_node_group_info.0.connectivity_info.0.public_access.0.type` is set to `SERVICE_PROVIDED_EIPS` and the cluster fulfill all other requirements for public access. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `bootstrap_brokers_sasl_iam` - One or more DNS names (or IP addresses) and SASL IAM port pairs. For example, `b-1.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9098,b-2.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9098,b-3.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9098`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS` and `client_authentication.0.sasl.0.iam` is set to `true`. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `bootstrap_brokers_sasl_scram` - One or more DNS names (or IP addresses) and SASL SCRAM port pairs. For example, `b-1.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9096,b-2.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9096,b-3.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9096`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS` and `client_authentication.0.sasl.0.scram` is set to `true`. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `bootstrap_brokers_tls` - One or more DNS names (or IP addresses) and TLS port pairs. For example, `b-1.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9094,b-2.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9094,b-3.exampleClusterName.abcde.c2.kafka.us-east-1.amazonaws.com:9094`. This attribute will have a value if `encryption_info.0.encryption_in_transit.0.client_broker` is set to `TLS_PLAINTEXT` or `TLS`. The resource sorts the list alphabetically. AWS may not always return all endpoints so the values may not be stable across applies.
* `current_version` - Current version of the MSK Cluster used for updates, e.g., `K13V1IB3VIYZZH`
* `encryption_info.0.encryption_at_rest_kms_key_arn` - The ARN of the KMS key used for encryption at rest of the broker data volumes.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `zookeeper_connect_string` - A comma separated list of one or more hostname:port pairs to use to connect to the Apache Zookeeper cluster. The returned values are sorted alphabetically. The AWS API may not return all endpoints, so this value is not guaranteed to be stable across applies.
* `zookeeper_connect_string_tls` - A comma separated list of one or more hostname:port pairs to use to connect to the Apache Zookeeper cluster via TLS. The returned values are sorted alphabetically. The AWS API may not return all endpoints, so this value is not guaranteed to be stable across applies.

## Timeouts

`aws_msk_cluster` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `120 minutes`) How long to wait for the MSK Cluster to be created.
* `update` - (Default `120 minutes`) How long to wait for the MSK Cluster to be updated.
Note that the `update` timeout is used separately for `ebs_volume_size`, `instance_type`, `number_of_broker_nodes`, `configuration_info`, `kafka_version` and monitoring and logging update timeouts.
* `delete` - (Default `120 minutes`) How long to wait for the MSK Cluster to be deleted.

## Import

MSK clusters can be imported using the cluster `arn`, e.g.,

```
$ terraform import aws_msk_cluster.example arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3
```
