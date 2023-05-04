---
subcategory: "AppIntegrations"
layout: "aws"
page_title: "AWS: aws_appintegrations_event_integration"
description: |-
  Provides details about an Amazon AppIntegrations Event Integration
---

# Data Source: aws_appintegrations_event_integration

Use this data source to get information on an existing AppIntegrations Event Integration.

## Example Usage

```terraform
data "aws_appintegrations_event_integration" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The AppIntegrations Event Integration name.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `arn` - The ARN of the AppIntegrations Event Integration.
* `description` - The description of the Event Integration.
* `eventbridge_bus` - The EventBridge bus.
* `event_filter` - A block that defines the configuration information for the event filter. The Event Filter block is documented below.
* `id` - The identifier of the Event Integration which is the name of the Event Integration.
* `tags` - Metadata that you can assign to help organize the report plans you create.

### Event Filter Attributes

For **event_filter** the following attributes are supported:

* `source` - The source of the events.
