---
subcategory: "Managed Streaming for Kafka Connect"
layout: "aws"
page_title: "AWS: aws_mskconnect_worker_configuration"
description: |-
  Get information on an Amazon MSK Connect worker configuration.
---

# Data Source: aws_mskconnect_worker_configuration

Get information on an Amazon MSK Connect Worker Configuration.

## Example Usage

```terraform
data "aws_mskconnect_worker_configuration" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the worker configuration.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - the ARN of the worker configuration.
* `description` - a summary description of the worker configuration.
* `latest_revision` - an ID of the latest successfully created revision of the worker configuration.
* `properties_file_content` - contents of connect-distributed.properties file.
