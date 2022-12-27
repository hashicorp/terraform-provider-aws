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

The following arguments are supported:

* `name` - (Required) Name of the connector.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the connector.
* `description` - Summary description of the connector.
* `version` - Current version of the connector.
