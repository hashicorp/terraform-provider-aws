---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_configuration"
description: |-
  Terraform resource for managing an Amazon Managed Streaming for Kafka configuration
---

# Resource: aws_msk_configuration

Manages an Amazon Managed Streaming for Kafka configuration. More information can be found on the [MSK Developer Guide](https://docs.aws.amazon.com/msk/latest/developerguide/msk-configuration.html).

## Example Usage

```terraform
resource "aws_msk_configuration" "example" {
  kafka_versions = ["2.1.0"]
  name           = "example"

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
delete.topic.enable = true
PROPERTIES
}
```

## Argument Reference

This resource supports the following arguments:

* `server_properties` - (Required) Contents of the server.properties file. Supported properties are documented in the [MSK Developer Guide](https://docs.aws.amazon.com/msk/latest/developerguide/msk-configuration-properties.html).
* `kafka_versions` - (Optional) List of Apache Kafka versions which can use this configuration.
* `name` - (Required) Name of the configuration.
* `description` - (Optional) Description of the configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the configuration.
* `latest_revision` - Latest revision of the configuration.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MSK configurations using the configuration ARN. For example:

```terraform
import {
  to = aws_msk_configuration.example
  id = "arn:aws:kafka:us-west-2:123456789012:configuration/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3"
}
```

Using `terraform import`, import MSK configurations using the configuration ARN. For example:

```console
% terraform import aws_msk_configuration.example arn:aws:kafka:us-west-2:123456789012:configuration/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3
```
