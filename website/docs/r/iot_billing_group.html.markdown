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

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the Billing Group.
* `properties` - (Optional) The Billing Group properties. Defined below.
* `tags` - (Optional) Key-value mapping of resource tags

### properties Reference

* `description` - (Optional) A description of the Billing Group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the Billing Group.
* `id` - The Billing Group ID.
* `version` - The current version of the Billing Group record in the registry.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IoT Billing Groups using the name. For example:

```terraform
import {
  to = aws_iot_billing_group.example
  id = "example"
}
```

Using `terraform import`, import IoT Billing Groups using the name. For example:

```console
% terraform import aws_iot_billing_group.example example
```
