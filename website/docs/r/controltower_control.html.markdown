---
subcategory: "Control Tower"
layout: "aws"
page_title: "AWS: aws_controltower_control"
description: |-
  Allows the application of pre-defined controls to organizational units.
---

# Resource: aws_controltower_control

Allows the application of pre-defined controls to organizational units. For more information on usage, please see the
[AWS Control Tower User Guide](https://docs.aws.amazon.com/controltower/latest/userguide/enable-guardrails.html).

## Example Usage

```terraform
data "aws_region" "current" {}

data "aws_organizations_organization" "example" {}

data "aws_organizations_organizational_units" "example" {
  parent_id = data.aws_organizations_organization.example.roots[0].id
}

resource "aws_controltower_control" "example" {
  control_identifier = "arn:aws:controltower:${data.aws_region.current.name}::control/AWS-GR_EC2_VOLUME_INUSE_CHECK"
  target_identifier = [
    for x in data.aws_organizations_organizational_units.example.children :
    x.arn if x.name == "Infrastructure"
  ][0]
}
```

## Argument Reference

The following arguments are supported:

* `control_identifier` - (Required) The ARN of the control. Only Strongly recommended and Elective controls are permitted, with the exception of the Region deny guardrail.
* `target_identifier` - (Required) The ARN of the organizational unit.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the organizational unit.

## Import

Control Tower Controls can be imported using their `organizational_unit_arn/control_identifier`, e.g.,

```
$ terraform import aws_controltower_control.example arn:aws:organizations::123456789101:ou/o-qqaejywet/ou-qg5o-ufbhdtv3,arn:aws:controltower:us-east-1::control/WTDSMKDKDNLE
```
