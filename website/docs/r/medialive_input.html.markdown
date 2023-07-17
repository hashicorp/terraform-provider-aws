---
subcategory: "Elemental MediaLive"
layout: "aws"
page_title: "AWS: aws_medialive_input"
description: |-
  Terraform resource for managing an AWS MediaLive Input.
---

# Resource: aws_medialive_input

Terraform resource for managing an AWS MediaLive Input.

## Example Usage

### Basic Usage

```terraform
resource "aws_medialive_input_security_group" "example" {
  whitelist_rules {
    cidr = "10.0.0.8/32"
  }

  tags = {
    ENVIRONMENT = "prod"
  }
}

resource "aws_medialive_input" "example" {
  name                  = "example-input"
  input_security_groups = [aws_medialive_input_security_group.example.id]
  type                  = "UDP_PUSH"

  tags = {
    ENVIRONMENT = "prod"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the input.
* `input_security_groups` - (Required) List of input security groups.
* `type` - (Required) The different types of inputs that AWS Elemental MediaLive supports.

The following arguments are optional:

* `destinations` - (Optional) Destination settings for PUSH type inputs. See [Destinations](#destinations) for more details.
* `input_devices` - (Optional) Settings for the devices. See [Input Devices](#input-devices) for more details.
* `media_connect_flows` - (Optional) A list of the MediaConnect Flows. See [Media Connect Flows](#media-connect-flows) for more details.
* `role_arn` - (Optional) The ARN of the role this input assumes during and after creation.
* `sources` - (Optional) The source URLs for a PULL-type input. See [Sources](#sources) for more details.
* `tags`- (Optional) A map of tags to assign to the Input. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc` - (Optional) Settings for a private VPC Input. See [VPC](#vpc) for more details.

### Destinations

* `stream_name` - A unique name for the location the RTMP stream is being pushed to.

### Input Devices

* `id` - The unique ID for the device.

### Media Connect Flows

* `flow_arn` - The ARN of the MediaConnect Flow

### Sources

* `password_param` - The key used to extract the password from EC2 Parameter store.
* `url` - The URL where the stream is pulled from.
* `username` - The username for the input source.

### VPC

* `subnet_ids` - A list of 2 VPC subnet IDs from the same VPC.
* `security_group_ids` - A list of up to 5 EC2 VPC security group IDs to attach to the Input.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Input.
* `attached_channels` - Channels attached to Input.
* `input_class` - The input class.
* `input_partner_ids` - A list of IDs for all Inputs which are partners of this one.
* `input_source_type` - Source type of the input.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

MediaLive Input can be imported using the `id`, e.g.,

```
$ terraform import aws_medialive_input.example 12345678
```
