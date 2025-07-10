---
subcategory: "IVS (Interactive Video)"
layout: "aws"
page_title: "AWS: aws_ivs_channel"
description: |-
  Terraform resource for managing an AWS IVS (Interactive Video) Channel.
---

# Resource: aws_ivs_channel

Terraform resource for managing an AWS IVS (Interactive Video) Channel.

## Example Usage

### Basic Usage

```terraform
resource "aws_ivs_channel" "example" {
  name = "channel-1"
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `authorized` - (Optional) If `true`, channel is private (enabled for playback authorization).
* `latency_mode` - (Optional) Channel latency mode. Valid values: `NORMAL`, `LOW`.
* `name` - (Optional) Channel name.
* `recording_configuration_arn` - (Optional) Recording configuration ARN.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) Channel type, which determines the allowable resolution and bitrate. Valid values: `STANDARD`, `BASIC`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Channel.
* `ingest_endpoint` - Channel ingest endpoint, part of the definition of an ingest server, used when setting up streaming software.
* `playback_url` - Channel playback URL.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IVS (Interactive Video) Channel using the ARN. For example:

```terraform
import {
  to = aws_ivs_channel.example
  id = "arn:aws:ivs:us-west-2:326937407773:channel/0Y1lcs4U7jk5"
}
```

Using `terraform import`, import IVS (Interactive Video) Channel using the ARN. For example:

```console
% terraform import aws_ivs_channel.example arn:aws:ivs:us-west-2:326937407773:channel/0Y1lcs4U7jk5
```
