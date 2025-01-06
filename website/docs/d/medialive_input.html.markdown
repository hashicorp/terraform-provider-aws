---
subcategory: "Elemental MediaLive"
layout: "aws"
page_title: "AWS: aws_medialive_input"
description: |-
  Terraform data source for managing an AWS Elemental MediaLive Input.
---

# Data Source: aws_medialive_input

Terraform data source for managing an AWS Elemental MediaLive Input.

## Example Usage

### Basic Usage

```terraform
data "aws_medialive_input" "example" {
  id = aws_medialive_input.example.id
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) The ID of the Input.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Input.
* `attached_channels` - Channels attached to Input.
* `destionations` - Destination settings for PUSH type inputs.
* `input_class` - The input class.
* `input_devices` - Settings for the devices.
* `input_partner_ids` - A list of IDs for all Inputs which are partners of this one.
* `input_source_type` - Source type of the input.
* `media_connect_flows` - A list of the MediaConnect Flows.
* `name` - Name of the input.
* `role_arn` - The ARN of the role this input assumes during and after creation.
* `security_groups` - List of input security groups.
* `sources` - The source URLs for a PULL-type input.
* `state` - The state of the input.
* `tags` - A map of tags assigned to the Input.
* `type` - The type of the input.
