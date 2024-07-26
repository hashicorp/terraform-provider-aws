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

## Argument Reference

The following arguments are required:

* `architecture` – (Optional) The CPU architecture of an application. Valid values are `ARM64` or `X86_64`. Default value is `X86_64`.
* `auto_start_configuration` – (Optional) The configuration for an application to automatically start on job submission.
* `auto_stop_configuration` – (Optional) The configuration for an application to automatically stop after a certain amount of time being idle.
* `image_configuration` – (Optional) The image configuration applied to all worker types.
* `initial_capacity` – (Optional) The capacity to initialize when the application is created.
* `interactive_configuration` – (Optional) Enables the interactive use cases to use when running an application.
* `maximum_capacity` – (Optional) The maximum capacity to allocate when the application is created. This is cumulative across all workers at any given point in time, not just when an application is created. No new resources will be created once any one of the defined limits is hit.
* `name` – (Required) The name of the application.
* `network_configuration` – (Optional) The network configuration for customer VPC connectivity.
* `release_label` – (Required) The EMR release version associated with the application.
* `type` – (Required) The type of application you want to start, such as `spark` or `hive`.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### auto_start_configuration Arguments

* `enabled` - (Optional) Enables the application to automatically start on job submission. Defaults to `true`.

### auto_stop_configuration Arguments

* `enabled` - (Optional) Enables the application to automatically stop after a certain amount of time being idle. Defaults to `true`.
* `idle_timeout_minutes` - (Optional) The amount of idle time in minutes after which your application will automatically stop. Defaults to `15` minutes.

### initial_capacity Arguments

* `initial_capacity_config` - (Optional) The initial capacity configuration per worker.
* `initial_capacity_type` - (Required) The worker type for an analytics framework. For Spark applications, the key can either be set to `Driver` or `Executor`. For Hive applications, it can be set to `HiveDriver` or `TezTask`.

### maximum_capacity Arguments

* `cpu` - (Required) The maximum allowed CPU for an application.
* `disk` - (Optional) The maximum allowed disk for an application.
* `memory` - (Required) The maximum allowed resources for an application.

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

Using `terraform import`, import EMR Severless applications using the `id`. For example:

```console
% terraform import aws_emrserverless_application.example id
```
