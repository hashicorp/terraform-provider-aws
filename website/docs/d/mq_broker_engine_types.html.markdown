---
subcategory: "MQ"
layout: "aws"
page_title: "AWS: aws_mq_broker_engine_types"
description: |-
  Retrieve information about available broker engines.
---

# Data Source: aws_mq_broker_engine_types

Retrieve information about available broker engines.

## Example Usage

### Basic Usage

```terraform
data "aws_mq_broker_engine_types" "example" {
  engine_type = "ACTIVEMQ"
}
```

## Argument Reference

* `engine_type` - (Optional) The MQ engine type to return version details for.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `broker_engine_types` - A list of available engine types and versions. See [Engine Types](#engine-types).

### engine-types

* `engine_type` - The broker's engine type.
* `engine_versions` - The list of engine versions.
