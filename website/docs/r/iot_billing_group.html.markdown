---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_billing_group"
description: |-
    Manages an AWS IoT Billing Group.
---

# Resource: aws_iot_billing_group

Manages an AWS IoT Billing Group.

## Example Usage

```terraform
resource "aws_iot_billing_group" "example" {
  name = "example"

  properties {
    description = "This is my billing group"
  }

  tags = {
    terraform = "true"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the Billing Group.
* `properties` - (Optional) The Billing Group properties. Defined below.
* `tags` - (Optional) Key-value mapping of resource tags

### properties Reference

* `description` - (Optional) A description of the Billing Group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Billing Group.
* `id` - The Billing Group ID.
* `version` - The current version of the Billing Group record in the registry.

## Import

IoT Billing Groups can be imported using the name, e.g.

```
$ terraform import aws_iot_billing_group.example example
```
