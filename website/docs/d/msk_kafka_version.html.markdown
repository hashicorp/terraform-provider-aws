---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_kafka_version"
description: |-
  Get information on a Amazon MSK Kafka Version
---

# Data Source: aws_msk_cluster

Get information on a Amazon MSK Kafka Version

## Example Usage

```terraform
data "aws_msk_kafka_version" "preferred" {
  preferred_versions = ["2.4.1.1", "2.4.1", "2.2.1"]
}

data "aws_msk_kafka_version" "example" {
  version = "2.8.0"
}
```

## Argument Reference

This data source supports the following arguments:

* `preferred_versions` - (Optional) Ordered list of preferred Kafka versions. The first match in this list will be returned. Either `preferred_versions` or `version` must be set.
* `version` - (Optional) Version of MSK Kafka. For example 2.4.1.1 or "2.2.1" etc. Either `preferred_versions` or `version` must be set.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `status` - Status of the MSK Kafka version eg. `ACTIVE` or `DEPRECATED`.
