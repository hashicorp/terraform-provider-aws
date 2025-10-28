---
subcategory: "EMR Serverless"
layout: "aws"
page_title: "AWS: aws_emrserverless_application"
description: |-
  Manages an EMR Serverless Application
---

# Resource: aws_emrserverless_application

Manages an EMR Serverless Application.

## Example Usage

### Basic Usage

```terraform
resource "aws_emrserverless_application" "example" {
  name          = "example"
  release_label = "emr-6.6.0"
  type          = "hive"
}
```

### Initial Capacity Usage

```terraform
resource "aws_emrserverless_application" "example" {
  name          = "example"
  release_label = "emr-6.6.0"
  type          = "hive"

  initial_capacity {
    initial_capacity_type = "HiveDriver"

    initial_capacity_config {
      worker_count = 1
      worker_configuration {
        cpu    = "2 vCPU"
        memory = "10 GB"
      }
    }
  }
}
```

### Maximum Capacity Usage

```terraform
resource "aws_emrserverless_application" "example" {
  name          = "example"
  release_label = "emr-6.6.0"
  type          = "hive"

  maximum_capacity {
    cpu    = "2 vCPU"
    memory = "10 GB"
  }
}
```

### Monitoring Configuration Usage

```terraform
resource "aws_emrserverless_application" "example" {
  name          = "example"
  release_label = "emr-7.1.0"
  type          = "spark"

  monitoring_configuration {
    cloudwatch_logging_configuration {
      enabled                = true
      log_group_name         = "/aws/emr-serverless/example"
      log_stream_name_prefix = "spark-logs"

      log_types {
        name   = "SPARK_DRIVER"
        values = ["STDOUT", "STDERR"]
      }

      log_types {
        name   = "SPARK_EXECUTOR"
        values = ["STDOUT"]
      }
    }

    managed_persistence_monitoring_configuration {
      enabled = true
    }

    prometheus_monitoring_configuration {
      remote_write_url = "https://prometheus-remote-write-endpoint.example.com/api/v1/write"
    }
  }
}
```

### Runtime Configuration Usage

```terraform
resource "aws_emrserverless_application" "example" {
  name          = "example"
  release_label = "emr-6.8.0"
  type          = "spark"
  runtime_configuration {
    classification = "spark-executor-log4j2"
    properties = {
      "rootLogger.level"                = "error"
      "logger.IdentifierForClass.name"  = "classpathForSettingLogger"
      "logger.IdentifierForClass.level" = "info"
    }
  }
  runtime_configuration {
    classification = "spark-defaults"
    properties = {
      "spark.executor.memory" = "1g"
      "spark.executor.cores"  = "1"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `runtime_configuration` - (Optional) A configuration specification to be used when provisioning an application. A configuration consists of a classification, properties, and optional nested configurations. A classification refers to an application-specific configuration file. Properties are the settings you want to change in that file.
* `architecture` - (Optional) The CPU architecture of an application. Valid values are `ARM64` or `X86_64`. Default value is `X86_64`.
* `auto_start_configuration` - (Optional) The configuration for an application to automatically start on job submission.
* `auto_stop_configuration` - (Optional) The configuration for an application to automatically stop after a certain amount of time being idle.
* `image_configuration` - (Optional) The image configuration applied to all worker types.
* `initial_capacity` - (Optional) The capacity to initialize when the application is created.
* `interactive_configuration` - (Optional) Enables the interactive use cases to use when running an application.
* `maximum_capacity` - (Optional) The maximum capacity to allocate when the application is created. This is cumulative across all workers at any given point in time, not just when an application is created. No new resources will be created once any one of the defined limits is hit.
* `monitoring_configuration` - (Optional) The configuration setting for monitoring.
* `name` - (Required) The name of the application.
* `network_configuration` - (Optional) The network configuration for customer VPC connectivity.
* `release_label` - (Required) The EMR release version associated with the application.
* `scheduler_configuration` - (Optional) Scheduler configuration for batch and streaming jobs running on this application. Supported with release labels `emr-7.0.0` and above. See [scheduler_configuration Arguments](#scheduler_configuration-arguments) below.
* `type` - (Required) The type of application you want to start, such as `spark` or `hive`.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### auto_start_configuration Arguments

* `enabled` - (Optional) Enables the application to automatically start on job submission. Defaults to `true`.

### auto_stop_configuration Arguments

* `enabled` - (Optional) Enables the application to automatically stop after a certain amount of time being idle. Defaults to `true`.
* `idle_timeout_minutes` - (Optional) The amount of idle time in minutes after which your application will automatically stop. Defaults to `15` minutes.

### runtime_configuration Arguments

* `classification` - (Required) The classification within a configuration.
* `properties` - (Optional) A set of properties specified within a configuration classification.

### initial_capacity Arguments

* `initial_capacity_config` - (Optional) The initial capacity configuration per worker.
* `initial_capacity_type` - (Required) The worker type for an analytics framework. For Spark applications, the key can either be set to `Driver` or `Executor`. For Hive applications, it can be set to `HiveDriver` or `TezTask`.

### maximum_capacity Arguments

* `cpu` - (Required) The maximum allowed CPU for an application.
* `disk` - (Optional) The maximum allowed disk for an application.
* `memory` - (Required) The maximum allowed resources for an application.

### monitoring_configuration Arguments

* `cloudwatch_logging_configuration` - (Optional) The Amazon CloudWatch configuration for monitoring logs.
* `s3_monitoring_configuration` - (Optional) The Amazon S3 configuration for monitoring log publishing.
* `managed_persistence_monitoring_configuration` - (Optional) The managed log persistence configuration for monitoring logs.
* `prometheus_monitoring_configuration` - (Optional) The Prometheus configuration for monitoring metrics.

#### cloudwatch_logging_configuration Arguments

* `enabled` - (Required) Enables CloudWatch logging.
* `log_group_name` - (Optional) The name of the log group in Amazon CloudWatch Logs where you want to publish your logs.
* `log_stream_name_prefix` - (Optional) Prefix for the CloudWatch log stream name.
* `log_types` - (Optional) The types of logs that you want to publish to CloudWatch. If you don't specify any log types, driver STDOUT and STDERR logs will be published to CloudWatch Logs by default. See [log_types](#log_types-arguments) for more details.
* `encryption_key_arn` - (Optional) The AWS Key Management Service (KMS) key ARN to encrypt the logs that you store in CloudWatch Logs.

##### log_types Arguments

* `name` - (Required) The worker type. Valid values are `SPARK_DRIVER`, `SPARK_EXECUTOR`, `HIVE_DRIVER`, and `TEZ_TASK`.
* `values` - (Required) The list of log types to publish. Valid values are `STDOUT`, `STDERR`, `HIVE_LOG`, `TEZ_AM`, and `SYSTEM_LOGS`.

#### s3_monitoring_configuration Arguments

* `log_uri` - (Optional) The Amazon S3 destination URI for log publishing.
* `encryption_key_arn` - (Optional) The KMS key ARN to encrypt the logs published to the given Amazon S3 destination.

#### managed_persistence_monitoring_configuration Arguments

* `enabled` - (Optional) Enables managed log persistence for monitoring logs.
* `encryption_key_arn` - (Optional) The KMS key ARN to encrypt the logs stored in managed persistence.

#### prometheus_monitoring_configuration Arguments

* `remote_write_url` - (Optional) The Prometheus remote write URL for sending metrics. Only supported in EMR 7.1.0 and later versions.

### network_configuration Arguments

* `security_group_ids` - (Optional) The array of security group Ids for customer VPC connectivity.
* `subnet_ids` - (Optional) The array of subnet Ids for customer VPC connectivity.

#### image_configuration Arguments

* `image_uri` - (Required) The image URI.

#### initial_capacity_config Arguments

* `worker_configuration` - (Optional) The resource configuration of the initial capacity configuration.
* `worker_count` - (Required) The number of workers in the initial capacity configuration.

### interactive_configuration Arguments

* `livy_endpoint_enabled` - (Optional) Enables an Apache Livy endpoint that you can connect to and run interactive jobs.
* `studio_enabled` - (Optional) Enables you to connect an application to Amazon EMR Studio to run interactive workloads in a notebook.

##### worker_configuration Arguments

* `cpu` - (Required) The CPU requirements for every worker instance of the worker type.
* `disk` - (Optional) The disk requirements for every worker instance of the worker type.
* `memory` - (Required) The memory requirements for every worker instance of the worker type.

### scheduler_configuration Arguments

When an empty `scheduler_configuration {}` block is specified, the feature is enabled with default settings.
To disable the feature after it has been enabled, remove the block from the configuration.

* `max_concurrent_runs` - (Optional) Maximum concurrent job runs on this application. Valid range is `1` to `1000`. Defaults to `15`.
* `queue_timeout_minutes` - (Optional) Maximum duration in minutes for the job in QUEUED state. Valid range is from `15` to `720`. Defaults to `360`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the cluster.
* `id` - The ID of the cluster.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EMR Severless applications using the `id`. For example:

```terraform
import {
  to = aws_emrserverless_application.example
  id = "id"
}
```

Using `terraform import`, import EMR Serverless applications using the `id`. For example:

```console
% terraform import aws_emrserverless_application.example id
```
