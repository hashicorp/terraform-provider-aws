---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_configuration"
description: |-
  Get information on an Amazon MSK Configuration
---

# Data Source: aws_msk_configuration

Get information on an Amazon MSK Configuration.

## Example Usage

```terraform
data "aws_msk_configuration" "example" {
  name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the configuration.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the configuration.
* `description` - Description of the configuration.
* `kafka_versions` - List of Apache Kafka versions which can use this configuration.
* `latest_revision` - Latest revision of the configuration.
* `server_properties` - Contents of the server.properties file.
