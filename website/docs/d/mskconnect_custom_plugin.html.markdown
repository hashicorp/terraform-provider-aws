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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the custom plugin.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - the ARN of the custom plugin.
* `description` - a summary description of the custom plugin.
* `latest_revision` - an ID of the latest successfully created revision of the custom plugin.
* `state` - the state of the custom plugin.
* `tags` - A map of tags assigned to the resource.
