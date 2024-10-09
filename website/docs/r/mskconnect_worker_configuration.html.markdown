---
subcategory: "Managed Streaming for Kafka Connect"
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

* `name` - (Required, Forces new resource) The name of the worker configuration.
* `properties_file_content` - (Required, Forces new resource) Contents of connect-distributed.properties file. The value can be either base64 encoded or in raw format.

The following arguments are optional:

* `description` - (Optional, Forces new resource) A summary description of the worker configuration.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - the Amazon Resource Name (ARN) of the worker configuration.
* `latest_revision` - an ID of the latest successfully created revision of the worker configuration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MSK Connect Worker Configuration using the plugin's `arn`. For example:

```terraform
import {
  to = aws_mskconnect_worker_configuration.example
  id = "arn:aws:kafkaconnect:eu-central-1:123456789012:worker-configuration/example/8848493b-7fcc-478c-a646-4a52634e3378-4"
}
```

Using `terraform import`, import MSK Connect Worker Configuration using the plugin's `arn`. For example:

```console
% terraform import aws_mskconnect_worker_configuration.example 'arn:aws:kafkaconnect:eu-central-1:123456789012:worker-configuration/example/8848493b-7fcc-478c-a646-4a52634e3378-4'
```
