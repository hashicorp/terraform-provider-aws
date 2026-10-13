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

This data source supports the following arguments:

* `id` - (Required) ID of the Input.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Input.
* `attached_channels` - Channels attached to Input.
* `destinations` - Destination settings for PUSH type inputs. See [`destinations` Block](#destinations-block) for more details.
* `input_class` - Input class.
* `input_devices` - Settings for the devices. See [`input_devices` Block](#input_devices-block) for more details.
* `input_partner_ids` - IDs for all Inputs which are partners of this one.
* `input_source_type` - Source type of the input.
* `media_connect_flows` - MediaConnect Flows. See [`media_connect_flows` Block](#media_connect_flows-block) for more details.
* `name` - Name of the input.
* `role_arn` - ARN of the role this input assumes during and after creation.
* `security_groups` - List of input security groups.
* `sources` - Source URLs for a PULL-type input. See [`sources` Block](#sources-block) for more details.
* `state` - State of the input.
* `tags` - Map of tags assigned to the Input.
* `type` - Type of the input.

### `destinations` Block

* `ip` - IP address.
* `port` - Port.
* `url` - URL.
* `vpc` - VPC settings. See [`destinations.vpc` Block](#destinationsvpc-block) for more details.

### `destinations.vpc` Block

* `availability_zone` - Availability Zone.
* `network_interface_id` - Network interface ID.

### `input_devices` Block

* `id` - ID of the device.

### `media_connect_flows` Block

* `flow_arn` - ARN of the MediaConnect Flow.

### `sources` Block

* `password_param` - Key used to extract the password from EC2 Parameter store.
* `url` - URL where the stream is pulled from.
* `username` - Username for the input source.
