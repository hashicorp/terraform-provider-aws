---
subcategory: "MQ"
layout: "aws"
page_title: "AWS: aws_mq_engine_versions"
description: |-
  Terraform data source for managing an AWS MQ Engine Versions.
---

# Data Source: aws_mq_engine_versions

Terraform data source for managing an AWS MQ Engine Versions.

## Example Usage

### Basic Usage

```terraform
data "aws_mq_engine_versions" "example" {
  filters {
    engine_type = "ACTIVEMQ"
  }
}
```

## Argument Reference

* `filters` - Filters the results of the request. See [Filters](#filters).

### filter

The following filters are optional.

* `engine_type` - (Optional) The database engine to return version details for.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `broker_engine_types` - A list of available engine types and versions. See [Engine Types](#engine-types).

### engine-types

* `engine_type` - The broker's engine type.
* `engine_versions` - The list of engine versions.
