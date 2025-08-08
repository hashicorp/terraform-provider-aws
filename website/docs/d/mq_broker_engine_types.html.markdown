---
subcategory: "MQ"
layout: "aws"
page_title: "AWS: aws_mq_broker_engine_types"
description: |-
  Provides details about available MQ broker engine types.
---

# Data Source: aws_mq_broker_engine_types

Provides details about available MQ broker engine types. Use this data source to retrieve supported engine types and their versions for Amazon MQ brokers.

## Example Usage

```terraform
data "aws_mq_broker_engine_types" "example" {
  engine_type = "ACTIVEMQ"
}
```

## Argument Reference

This data source supports the following arguments:

* `engine_type` - (Optional) MQ engine type to return version details for.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `broker_engine_types` - List of available engine types and versions. See [Engine Types](#engine-types).

### Engine Types

* `engine_type` - Broker's engine type.
* `engine_versions` - List of engine versions. See [Engine Versions](#engine-versions).

### Engine Versions

* `name` - Name of the engine version.
