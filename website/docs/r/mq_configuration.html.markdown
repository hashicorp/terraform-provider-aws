---
subcategory: "MQ"
layout: "aws"
page_title: "AWS: aws_mq_configuration"
description: "Manages an Amazon MQ configuration"
---

# Resource: aws_mq_configuration

Manages an Amazon MQ configuration. Use this resource to create and manage broker configurations for ActiveMQ and RabbitMQ brokers.

## Example Usage

### ActiveMQ

```terraform
resource "aws_mq_configuration" "example" {
  description    = "Example Configuration"
  name           = "example"
  engine_type    = "ActiveMQ"
  engine_version = "5.17.6"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>
DATA
}
```

### RabbitMQ

```terraform
resource "aws_mq_configuration" "example" {
  description    = "Example Configuration"
  name           = "example"
  engine_type    = "RabbitMQ"
  engine_version = "3.11.20"

  data = <<DATA
# Default RabbitMQ delivery acknowledgement timeout is 30 minutes in milliseconds
consumer_timeout = 1800000
DATA
}
```

## Argument Reference

The following arguments are required:

* `data` - (Required) Broker configuration in XML format for ActiveMQ or Cuttlefish format for RabbitMQ. See [AWS documentation](https://docs.aws.amazon.com/amazon-mq/latest/developer-guide/amazon-mq-broker-configuration-parameters.html) for supported parameters and format of the XML.
* `engine_type` - (Required) Type of broker engine. Valid values are `ActiveMQ` and `RabbitMQ`.
* `engine_version` - (Required) Version of the broker engine.
* `name` - (Required) Name of the configuration.

The following arguments are optional:

* `authentication_strategy` - (Optional) Authentication strategy associated with the configuration. Valid values are `simple` and `ldap`. `ldap` is not supported for RabbitMQ engine type.
* `description` - (Optional) Description of the configuration.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the configuration.
* `id` - Unique ID that Amazon MQ generates for the configuration.
* `latest_revision` - Latest revision of the configuration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MQ Configurations using the configuration ID. For example:

```terraform
import {
  to = aws_mq_configuration.example
  id = "c-0187d1eb-88c8-475a-9b79-16ef5a10c94f"
}
```

Using `terraform import`, import MQ Configurations using the configuration ID. For example:

```console
% terraform import aws_mq_configuration.example c-0187d1eb-88c8-475a-9b79-16ef5a10c94f
```
