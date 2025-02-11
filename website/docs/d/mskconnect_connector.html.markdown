---
subcategory: "Managed Streaming for Kafka Connect"
layout: "aws"
page_title: "AWS: aws_mskconnect_connector"
description: |-
  Get information on an Amazon MSK Connect Connector.
---

# Data Source: aws_mskconnect_connector

Get information on an Amazon MSK Connect Connector.

## Example Usage

```terraform
data "aws_mskconnect_connector" "example" {
  name = "example-mskconnector"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the connector.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connector.
* `description` - Summary description of the connector.
* `tags` - A map of tags assigned to the resource.
* `version` - Current version of the connector.
