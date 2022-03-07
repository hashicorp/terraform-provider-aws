---
subcategory: "Kafka Connect (MSK Connect)"
layout: "aws"
page_title: "AWS: aws_mskconnect_connector"
description: |-
  Get information on an Amazon MSK Connect connector.
---

# Data Source: aws_mskconnect_connector

Get information on an Amazon MSK Connect connector.

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

* `arn` - the Amazon Resource Name (ARN) of the connector.
* `description` - a summary description of the connector.
* `version` - an ID of the latest successfully created version of the connector.
* `state` - the state of the connector.
