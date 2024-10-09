---
subcategory: "Managed Streaming for Kafka Connect"
layout: "aws"
page_title: "AWS: aws_mskconnect_connector"
description: |-
  Provides an Amazon MSK Connect Connector resource.
---

# Resource: aws_mskconnect_connector

Provides an Amazon MSK Connect Connector resource.

## Example Usage

### Basic configuration

```terraform
resource "aws_mskconnect_connector" "example" {
  name = "example"

  kafkaconnect_version = "2.7.1"

  capacity {
    autoscaling {
      mcu_count        = 1
      min_worker_count = 1
      max_worker_count = 2

      scale_in_policy {
        cpu_utilization_percentage = 20
      }

      scale_out_policy {
        cpu_utilization_percentage = 80
      }
    }
  }

  connector_configuration = {
    "connector.class" = "com.github.jcustenborder.kafka.connect.simulator.SimulatorSinkConnector"
    "tasks.max"       = "1"
    "topics"          = "example"
  }

  kafka_cluster {
    apache_kafka_cluster {
      bootstrap_servers = aws_msk_cluster.example.bootstrap_brokers_tls

      vpc {
        security_groups = [aws_security_group.example.id]
        subnets         = [aws_subnet.example1.id, aws_subnet.example2.id, aws_subnet.example3.id]
      }
    }
  }

  kafka_cluster_client_authentication {
    authentication_type = "NONE"
  }

  kafka_cluster_encryption_in_transit {
    encryption_type = "TLS"
  }

  plugin {
    custom_plugin {
      arn      = aws_mskconnect_custom_plugin.example.arn
      revision = aws_mskconnect_custom_plugin.example.latest_revision
    }
  }

  service_execution_role_arn = aws_iam_role.example.arn
}
```

## Argument Reference

The following arguments are required:

* `capacity` - (Required) Information about the capacity allocated to the connector. See [`capacity` Block](#capacity-block) for details.
* `connector_configuration` - (Required) A map of keys to values that represent the configuration for the connector.
* `kafka_cluster` - (Required) Specifies which Apache Kafka cluster to connect to. See [`kafka_cluster` Block](#kafka_cluster-block) for details.
* `kafka_cluster_client_authentication` - (Required) Details of the client authentication used by the Apache Kafka cluster. See [`kafka_cluster_client_authentication` Block](#kafka_cluster_client_authentication-block) for details.
* `kafka_cluster_encryption_in_transit` - (Required) Details of encryption in transit to the Apache Kafka cluster. See [`kafka_cluster_encryption_in_transit` Block](#kafka_cluster_encryption_in_transit-block) for details.
* `kafkaconnect_version` - (Required) The version of Kafka Connect. It has to be compatible with both the Apache Kafka cluster's version and the plugins.
* `name` - (Required) The name of the connector.
* `plugin` - (Required) Specifies which plugins to use for the connector. See [`plugin` Block](#plugin-block) for details.
* `service_execution_role_arn` - (Required) The Amazon Resource Name (ARN) of the IAM role used by the connector to access the Amazon Web Services resources that it needs. The types of resources depends on the logic of the connector. For example, a connector that has Amazon S3 as a destination must have permissions that allow it to write to the S3 destination bucket.

The following arguments are optional:

* `description` - (Optional) A summary description of the connector.
* `log_delivery` - (Optional) Details about log delivery. See [`log_delivery` Block](#log_delivery-block) for details.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `worker_configuration` - (Optional) Specifies which worker configuration to use with the connector. See [`worker_configuration` Block](#worker_configuration-block) for details.

### `capacity` Block

The `capacity` configuration block supports the following arguments:

* `autoscaling` - (Optional) Information about the auto scaling parameters for the connector. See [`autoscaling` Block](#autoscaling-block) for details.
* `provisioned_capacity` - (Optional) Details about a fixed capacity allocated to a connector. See [`provisioned_capacity` Block](#provisioned_capacity-block) for details.

### `autoscaling` Block

The `autoscaling` configuration block supports the following arguments:

* `max_worker_count` - (Required) The maximum number of workers allocated to the connector.
* `mcu_count` - (Optional) The number of microcontroller units (MCUs) allocated to each connector worker. Valid values: `1`, `2`, `4`, `8`. The default value is `1`.
* `min_worker_count` - (Required) The minimum number of workers allocated to the connector.
* `scale_in_policy` - (Optional) The scale-in policy for the connector. See [`scale_in_policy` Block](#scale_in_policy-block) for details.
* `scale_out_policy` - (Optional) The scale-out policy for the connector. See [`scale_out_policy` Block](#scale_out_policy-block) for details.

### `scale_in_policy` Block

The `scale_in_policy` configuration block supports the following arguments:

* `cpu_utilization_percentage` - (Required) Specifies the CPU utilization percentage threshold at which you want connector scale in to be triggered.

### `scale_out_policy` Block

The `scale_out_policy` configuration block supports the following arguments:

* `cpu_utilization_percentage` - (Required) The CPU utilization percentage threshold at which you want connector scale out to be triggered.

### `provisioned_capacity` Block

The `provisioned_capacity` configuration block supports the following arguments:

* `mcu_count` - (Optional) The number of microcontroller units (MCUs) allocated to each connector worker. Valid values: `1`, `2`, `4`, `8`. The default value is `1`.
* `worker_count` - (Required) The number of workers that are allocated to the connector.

### `kafka_cluster` Block

The `kafka_cluster` configuration block supports the following arguments:

* `apache_kafka_cluster` - (Required) The Apache Kafka cluster to which the connector is connected. See [`apache_kafka_cluster` Block](#apache_kafka_cluster-block) for details.

### `apache_kafka_cluster` Block

The `apache_kafka_cluster` configuration block supports the following arguments:

* `bootstrap_servers` - (Required) The bootstrap servers of the cluster.
* `vpc` - (Required) Details of an Amazon VPC which has network connectivity to the Apache Kafka cluster. See [`vpc` Block](#vpc-block) for details.

### `vpc` Block

The `vpc` configuration block supports the following arguments:

* `security_groups` - (Required) The security groups for the connector.
* `subnets` - (Required) The subnets for the connector.

### `kafka_cluster_client_authentication` Block

The `kafka_cluster_client_authentication` configuration block supports the following arguments:

* `authentication_type` - (Optional) The type of client authentication used to connect to the Apache Kafka cluster. Valid values: `IAM`, `NONE`. A value of `NONE` means that no client authentication is used. The default value is `NONE`.

### `kafka_cluster_encryption_in_transit` Block

The `kafka_cluster_encryption_in_transit` configuration block supports the following arguments:

* `encryption_type` - (Optional) The type of encryption in transit to the Apache Kafka cluster. Valid values: `PLAINTEXT`, `TLS`. The default values is `PLAINTEXT`.

### `log_delivery` Block

The `log_delivery` configuration block supports the following arguments:

* `worker_log_delivery` - (Required) The workers can send worker logs to different destination types. This configuration specifies the details of these destinations. See [`worker_log_delivery` Block](#worker_log_delivery-block) for details.

### `worker_log_delivery` Block

The `worker_log_delivery` configuration block supports the following arguments:

* `cloudwatch_logs` - (Optional) Details about delivering logs to Amazon CloudWatch Logs. See [`cloudwatch_logs` Block](#cloudwatch_logs-block) for details.
* `firehose` - (Optional) Details about delivering logs to Amazon Kinesis Data Firehose. See [`firehose` Block](#firehose-block) for details.
* `s3` - (Optional) Details about delivering logs to Amazon S3. See [`s3` Block](#s3-block) for deetails.

### `cloudwatch_logs` Block

The `cloudwatch_logs` configuration block supports the following arguments:

* `enabled` - (Optional) Whether log delivery to Amazon CloudWatch Logs is enabled.
* `log_group` - (Required) The name of the CloudWatch log group that is the destination for log delivery.

### `firehose` Block

The `firehose` configuration block supports the following arguments:

* `delivery_stream` - (Optional) The name of the Kinesis Data Firehose delivery stream that is the destination for log delivery.
* `enabled` - (Required) Specifies whether connector logs get delivered to Amazon Kinesis Data Firehose.

### `s3` Block

The `s3` configuration block supports the following arguments:

* `bucket` - (Optional) The name of the S3 bucket that is the destination for log delivery.
* `enabled` - (Required) Specifies whether connector logs get sent to the specified Amazon S3 destination.
* `prefix` - (Optional) The S3 prefix that is the destination for log delivery.

### `plugin` Block

The `plugin` configuration block supports the following argumens:

* `custom_plugin` - (Required) Details about a custom plugin. See [`custom_plugin` Block](#custom_plugin-block) for details.

### `custom_plugin` Block

The `custom_plugin` configuration block supports the following arguments:

* `arn` - (Required) The Amazon Resource Name (ARN) of the custom plugin.
* `revision` - (Required) The revision of the custom plugin.

### `worker_configuration` Block

The `worker_configuration` configuration block supports the following arguments:

* `arn` - (Required) The Amazon Resource Name (ARN) of the worker configuration.
* `revision` - (Required) The revision of the worker configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the connector.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version` - The current version of the connector.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `update` - (Default `20m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MSK Connect Connector using the connector's `arn`. For example:

```terraform
import {
  to = aws_mskconnect_connector.example
  id = "arn:aws:kafkaconnect:eu-central-1:123456789012:connector/example/264edee4-17a3-412e-bd76-6681cfc93805-3"
}
```

Using `terraform import`, import MSK Connect Connector using the connector's `arn`. For example:

```console
% terraform import aws_mskconnect_connector.example 'arn:aws:kafkaconnect:eu-central-1:123456789012:connector/example/264edee4-17a3-412e-bd76-6681cfc93805-3'
```
