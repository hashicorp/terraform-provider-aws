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

This data source supports the following arguments:

* `name` - (Required) The AppIntegrations Event Integration name.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the AppIntegrations Event Integration.
* `description` - The description of the Event Integration.
* `eventbridge_bus` - The EventBridge bus.
* `event_filter` - A block that defines the configuration information for the event filter. The Event Filter block is documented below.
* `id` - The identifier of the Event Integration which is the name of the Event Integration.
* `tags` - Metadata that you can assign to help organize the report plans you create.

### Event Filter Attributes

`event_filter` has the following attributes:

* `source` - The source of the events.
