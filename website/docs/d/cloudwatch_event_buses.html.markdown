---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_buses"
description: |-
  Terraform data source for managing an AWS EventBridge (Cloudwatch) Event Buses.
---

# Data Source: aws_cloudwatch_event_buses

Terraform data source for managing an AWS EventBridge Event Buses.

## Example Usage

### Basic Usage

```terraform
data "aws_cloudwatch_event_buses" "example" {
  name_prefix = "test"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name_prefix` - (Optional) Specifying this limits the results to only those event buses with names that start with the specified prefix.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `event_buses` - This list of event buses.

### `event_buses` Attribute Reference

* `arn` - The ARN of the event bus.
* `creation_time` - The time the event bus was created.
* `description` - The event bus description.
* `last_modified_time` - The time the event bus was last modified.
* `name` - The name of the event bus.
* `policy` - The permissions policy of the event bus, describing which other AWS accounts can write events to this event bus.
