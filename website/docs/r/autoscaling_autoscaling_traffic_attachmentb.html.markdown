---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_autoscaling_traffic_attachmentb"
description: |-
  Terraform resource for managing an AWS Auto Scaling Autoscaling Traffic Attachmentb.
---

# Resource: aws_autoscaling_autoscaling_traffic_attachmentb

Terraform resource for managing an AWS Auto Scaling Autoscaling Traffic Attachmentb.

## Example Usage

### Basic Usage

```terraform
resource "aws_autoscaling_autoscaling_traffic_attachmentb" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Autoscaling Traffic Attachmentb. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

Auto Scaling Autoscaling Traffic Attachmentb can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_autoscaling_autoscaling_traffic_attachmentb.example rft-8012925589
```
