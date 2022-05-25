---
subcategory: "PinpointSMSVoiceV2"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_opt_out_list"
description: |-
  Terraform resource for managing an AWS Pinpoint SMS Voice V2 Opt Out List.
---

# Resource: aws_pinpointsmsvoicev2_opt_out_list

Terraform resource for managing an AWS Pinpoint SMS Voice V2 Opt Out List.

## Example Usage

### Basic Usage

```terraform
resource "aws_pinpointsmsvoicev2_opt_out_list" "example" {
  name = "example-opt-out-list"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the opt-out list.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the opt-out list.

## Import

Pinpoint SMS Voice V2 opt-out list can be imported using the `name`, e.g.,

```
$ terraform import aws_pinpointsmsvoicev2_opt_out_list.example example-opt-out-list
```
