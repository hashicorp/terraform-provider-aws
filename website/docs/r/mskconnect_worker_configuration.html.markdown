---
subcategory: "Kafka Connect (MSK Connect)"
layout: "aws"
page_title: "AWS: aws_mskconnect_worker_configuration"
description: |-
  Provides an Amazon MSK Connect worker configuration resource.
---

# Resource: aws_mskconnect_worker_configuration

Provides an Amazon MSK Connect Worker Configuration Resource.

## Example Usage

### Basic configuration

```terraform
resource "aws_mskconnect_worker_configuration" "example" {
  name                    = "example"
  properties_file_content = <<EOT
key.converter=org.apache.kafka.connect.storage.StringConverter
value.converter=org.apache.kafka.connect.storage.StringConverter
EOT
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the worker configuration.
* `properties_file_content` - (Required) Contents of connect-distributed.properties file. The value can be either base64 encoded or in raw format.

The following arguments are optional:

* `description` - (Optional) A summary description of the worker configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - the Amazon Resource Name (ARN) of the worker configuration.
* `latest_revision` - an ID of the latest successfully created revision of the worker configuration.

## Import

MSK Connect Worker Configuration can be imported using the plugin's `arn`, e.g.,

```
$ terraform import aws_mskconnect_worker_configuration.example 'arn:aws:kafkaconnect:eu-central-1:123456789012:worker-configuration/example/8848493b-7fcc-478c-a646-4a52634e3378-4'
```
