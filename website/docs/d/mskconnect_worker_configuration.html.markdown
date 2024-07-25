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

This data source supports the following arguments:

* `name` - (Required) Name of the worker configuration.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - the ARN of the worker configuration.
* `description` - a summary description of the worker configuration.
* `latest_revision` - an ID of the latest successfully created revision of the worker configuration.
* `properties_file_content` - contents of connect-distributed.properties file.
* `tags` - A map of tags assigned to the resource.
