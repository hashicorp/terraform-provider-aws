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

This resource supports the following arguments:

* `description` - (Optional) Description of the Event Integration.
* `eventbridge_bus` - (Required) EventBridge bus.
* `event_filter` - (Required) Block that defines the configuration information for the event filter. The Event Filter block is documented below.
* `name` - (Required) Name of the Event Integration.
* `tags` - (Optional) Tags to apply to the Event Integration. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

A `event_filter` block supports the following arguments:

* `source` - (Required) Source of the events.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Event Integration.
* `id` - Identifier of the Event Integration which is the name of the Event Integration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon AppIntegrations Event Integrations using the `name`. For example:

```terraform
import {
  to = aws_appintegrations_event_integration.example
  id = "example-name"
}
```

Using `terraform import`, import Amazon AppIntegrations Event Integrations using the `name`. For example:

```console
% terraform import aws_appintegrations_event_integration.example example-name
```
