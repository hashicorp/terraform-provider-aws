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
* `type` - (Required) Type of input that AWS Elemental MediaLive supports.

The following arguments are optional:

* `destinations` - (Optional) Destination settings for PUSH type inputs. See [`destinations` Block](#destinations-block) for more details.
* `input_devices` - (Optional) Settings for the devices. See [`input_devices` Block](#input_devices-block) for more details.
* `input_security_groups` - (Optional) List of input security groups.
* `media_connect_flows` - (Optional) MediaConnect Flows. See [`media_connect_flows` Block](#media_connect_flows-block) for more details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role_arn` - (Optional) ARN of the role this input assumes during and after creation.
* `sources` - (Optional) Source URLs for a PULL-type input. See [`sources` Block](#sources-block) for more details.
* `tags` - (Optional) Map of tags to assign to the Input. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc` - (Optional) Settings for a private VPC Input. See [`vpc` Block](#vpc-block) for more details.

### `destinations` Block

* `stream_name` - (Required) Unique name for the location the RTMP stream is being pushed to.

### `input_devices` Block

* `id` - (Required) Unique ID for the device.

### `media_connect_flows` Block

* `flow_arn` - (Required) ARN of the MediaConnect Flow.

### `sources` Block

* `password_param` - (Required) Key used to extract the password from EC2 Parameter store.
* `url` - (Required) URL where the stream is pulled from.
* `username` - (Required) Username for the input source.

### `vpc` Block

* `security_group_ids` - (Optional) List of up to 5 EC2 VPC security group IDs to attach to the Input.
* `subnet_ids` - (Required) List of 2 VPC subnet IDs from the same VPC.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Input.
* `attached_channels` - Channels attached to Input.
* `input_class` - Input class.
* `input_partner_ids` - IDs for all Inputs which are partners of this one.
* `input_source_type` - Source type of the input.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_medialive_input.example
  identity = {
    id = "12345678"
  }
}

resource "aws_medialive_input" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) ID of the MediaLive Input.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MediaLive Input using the `id`. For example:

```terraform
import {
  to = aws_medialive_input.example
  id = "12345678"
}
```

Using `terraform import`, import MediaLive Input using the `id`. For example:

```console
% terraform import aws_medialive_input.example 12345678
```
