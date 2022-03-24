---
subcategory: "Kafka Connect (MSK Connect)"
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

* `arn` - The Amazon Resource Name (ARN) of the connector.
* `description` - A summary description of the connector.
* `version` - The current version of the connector.
