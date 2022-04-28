---
subcategory: "MQ"
layout: "aws"
page_title: "AWS: aws_mq_broker_instance_type_offerings"
description: |-
  Provides a MQ Broker Instance Offerings data source.
---

# Data Source: aws_mq_broker_instance_type_offerings

Provides information about a MQ Broker Instance Offerings.

## Example Usage

```terraform
data "aws_mq_broker_instance_type_offerings" "empty" {}

data "aws_mq_broker_instance_type_offerings" "engine" {
  engine_type = "ACTIVEMQ"
}

data "aws_mq_broker_instance_type_offerings" "storage" {
  storage_type = "EBS"
}

data "aws_mq_broker_instance_type_offerings" "instance" {
  host_instance_type = "mq.m5.large"
}

data "aws_mq_broker_instance_type_offerings" "all" {
  host_instance_type = "mq.m5.large"
  storage_type       = "EBS"
  engine_type        = "ACTIVEMQ"
}
```

## Argument Reference

The following arguments are supported:

* `engine_type` - (Optional) Filter response by engine type.
* `host_instance_type` - (Optional) Filter response by host instance type.
* `storage_type` - (Optional) Filter response by storage type.

## Attributes Reference

* `broker_instance_options` -  Option for host instance type. See Broker Instance Options below.

### Broker Instance Options

* `availability_zones` - The list of available AZs. See Availability Zones. below
* `engine_type` - The broker's engine type.
* `host_instance_type` - The broker's instance type.
* `storage_type` - The broker's storage type.
* `supported_deployment_modes` - The list of supported deployment modes.
* `supported_engine_versions` - The list of supported engine versions.

### Availability Zones

* `name` - The name of the Availability Zone.
