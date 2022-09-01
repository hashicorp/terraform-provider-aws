---
subcategory: "AppIntegrations"
layout: "aws"
page_title: "AWS: aws_appintegrations_event_integration"
description: |-
  Provides details about a specific Amazon AppIntegrations Event Integration
---

# Resource: aws_appintegrations_event_integration

Provides an Amazon AppIntegrations Event Integration resource.

## Example Usage

```terraform
resource "aws_appintegrations_event_integration" "example" {
  name            = "example-name"
  description     = "Example Description"
  eventbridge_bus = "default"

  event_filter {
    source = "aws.partner/examplepartner.com"
  }

  tags = {
    "Name" = "Example Event Integration"
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) Description of the Event Integration.
* `eventbridge_bus` - (Required) EventBridge bus.
* `event_filter` - (Required) Block that defines the configuration information for the event filter. The Event Filter block is documented below.
* `name` - (Required) Name of the Event Integration.
* `tags` - (Optional) Tags to apply to the Event Integration. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

A `event_filter` block supports the following arguments:

* `source` - (Required) Source of the events.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Event Integration.
* `id` - Identifier of the Event Integration which is the name of the Event Integration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Amazon AppIntegrations Event Integrations can be imported using the `name` e.g.,

```
$ terraform import aws_appintegrations_event_integration.example example-name
```
