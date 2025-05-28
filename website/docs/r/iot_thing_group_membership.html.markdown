---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_thing_group_membership"
description: |-
    Adds an IoT Thing to an IoT Thing Group.
---

# Resource: aws_iot_thing_group_membership

Adds an IoT Thing to an IoT Thing Group.

## Example Usage

```terraform
resource "aws_iot_thing_group_membership" "example" {
  thing_name       = "example-thing"
  thing_group_name = "example-group"

  override_dynamic_group = true
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `thing_name` - (Required) The name of the thing to add to a group.
* `thing_group_name` - (Required) The name of the group to which you are adding a thing.
* `override_dynamic_group` - (Optional) Override dynamic thing groups with static thing groups when 10-group limit is reached. If a thing belongs to 10 thing groups, and one or more of those groups are dynamic thing groups, adding a thing to a static group removes the thing from the last dynamic group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The membership ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IoT Thing Group Membership using the thing group name and thing name. For example:

```terraform
import {
  to = aws_iot_thing_group_membership.example
  id = "thing_group_name/thing_name"
}
```

Using `terraform import`, import IoT Thing Group Membership using the thing group name and thing name. For example:

```console
% terraform import aws_iot_thing_group_membership.example thing_group_name/thing_name
```
