---
subcategory: "CloudWatch Observability Access Manager"
layout: "aws"
page_title: "AWS: aws_oam_sink"
description: |-
  Terraform resource for managing an AWS CloudWatch Observability Access Manager Sink.
---

# Resource: aws_oam_sink

Terraform resource for managing an AWS CloudWatch Observability Access Manager Sink.

## Example Usage

### Basic Usage

```terraform
resource "aws_oam_sink" "example" {
  name = "ExampleSink"

  tags = {
    Env = "prod"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name for the sink.

The following arguments are optional:

* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Sink.
* `id` - ARN of the Sink.
* `sink_id` - ID string that AWS generated as part of the sink ARN.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `1m`)
* `update` - (Default `1m`)
* `delete` - (Default `1m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Observability Access Manager Sink using the `arn`. For example:

```terraform
import {
  to = aws_oam_sink.example
  id = "arn:aws:oam:us-west-2:123456789012:sink/sink-id"
}
```

Using `terraform import`, import CloudWatch Observability Access Manager Sink using the `arn`. For example:

```console
% terraform import aws_oam_sink.example arn:aws:oam:us-west-2:123456789012:sink/sink-id
```
