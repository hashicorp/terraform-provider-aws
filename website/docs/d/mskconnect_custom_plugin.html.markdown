---
subcategory: "Managed Streaming for Kafka Connect"
layout: "aws"
page_title: "AWS: aws_mskconnect_custom_plugin"
description: |-
  Get information on an Amazon MSK Connect custom plugin.
---

# Data Source: aws_mskconnect_custom_plugin

Get information on an Amazon MSK Connect custom plugin.

## Example Usage

```terraform
data "aws_mskconnect_custom_plugin" "example" {
  name = "example-debezium-1"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the custom plugin.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - the ARN of the custom plugin.
* `description` - a summary description of the custom plugin.
* `latest_revision` - an ID of the latest successfully created revision of the custom plugin.
* `state` - the state of the custom plugin.
