---
subcategory: "MQ"
layout: "aws"
page_title: "AWS: aws_mq_broker_instance_type_offerings"
description: |-
  Provides details about available MQ broker instance type offerings.
---

# Data Source: aws_mq_broker_instance_type_offerings

Provides details about available MQ broker instance type offerings. Use this data source to discover supported instance types, storage types, and deployment modes for Amazon MQ brokers.

## Example Usage

```terraform
# Get all instance type offerings
data "aws_mq_broker_instance_type_offerings" "all" {}

# Filter by engine type
data "aws_mq_broker_instance_type_offerings" "activemq" {
  engine_type = "ACTIVEMQ"
}

# Filter by storage type
data "aws_mq_broker_instance_type_offerings" "ebs" {
  storage_type = "EBS"
}

# Filter by instance type
data "aws_mq_broker_instance_type_offerings" "m5" {
  host_instance_type = "mq.m5.large"
}

# Filter by multiple criteria
data "aws_mq_broker_instance_type_offerings" "filtered" {
  engine_type        = "ACTIVEMQ"
  storage_type       = "EBS"
  host_instance_type = "mq.m5.large"
}
```

## Argument Reference

The following arguments are optional:

* `engine_type` - (Optional) Filter response by engine type.
* `host_instance_type` - (Optional) Filter response by host instance type.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `storage_type` - (Optional) Filter response by storage type.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `broker_instance_options` - List of broker instance options. See [Broker Instance Options](#broker-instance-options) below.

### Broker Instance Options

* `availability_zones` - List of available Availability Zones. See [Availability Zones](#availability-zones) below.
* `engine_type` - Broker's engine type.
* `host_instance_type` - Broker's instance type.
* `storage_type` - Broker's storage type.
* `supported_deployment_modes` - List of supported deployment modes.
* `supported_engine_versions` - List of supported engine versions.

### Availability Zones

* `name` - Name of the Availability Zone.
