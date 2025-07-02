---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_thing"
description: |-
    Creates and manages an AWS IoT Thing.
---

# Resource: aws_iot_thing

Creates and manages an AWS IoT Thing.

## Example Usage

```terraform
resource "aws_iot_thing" "example" {
  name = "example"

  attributes = {
    First = "examplevalue"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the thing.
* `attributes` - (Optional) Map of attributes of the thing.
* `thing_type_name` - (Optional) The thing type name.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `default_client_id` - The default client ID.
* `version` - The current version of the thing record in the registry.
* `arn` - The ARN of the thing.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IOT Things using the name. For example:

```terraform
import {
  to = aws_iot_thing.example
  id = "example"
}
```

Using `terraform import`, import IOT Things using the name. For example:

```console
% terraform import aws_iot_thing.example example
```
