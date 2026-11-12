---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_protect_configuration_association"
description: |-
  Manages an AWS End User Messaging SMS Protect Configuration Association.
---

# Resource: aws_pinpointsmsvoicev2_protect_configuration_association

Manages an AWS End User Messaging SMS Protect Configuration Association. The association applies a protect configuration's country rules to messages sent through a configuration set.

## Example Usage

```terraform
resource "aws_pinpointsmsvoicev2_configuration_set" "example" {
  name = "example"
}

resource "aws_pinpointsmsvoicev2_protect_configuration" "example" {
  sms_country_rules = {
    RU = "BLOCK"
  }
}

resource "aws_pinpointsmsvoicev2_protect_configuration_association" "example" {
  configuration_set_name   = aws_pinpointsmsvoicev2_configuration_set.example.name
  protect_configuration_id = aws_pinpointsmsvoicev2_protect_configuration.example.id
}
```

## Argument Reference

The following arguments are required:

* `configuration_set_name` - (Required, Forces new resource) Name of the configuration set.
* `protect_configuration_id` - (Required) ID of the protect configuration.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

~> **Note:** A configuration set can have only one protect configuration. Creating this resource for a configuration set that already has a protect configuration replaces it.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_protect_configuration_association.example
  identity = {
    configuration_set_name = "example"
  }
}

resource "aws_pinpointsmsvoicev2_protect_configuration_association" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `configuration_set_name` - (String) Name of the configuration set.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS End User Messaging SMS Protect Configuration Association using the `configuration_set_name`. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_protect_configuration_association.example
  id = "example"
}
```

Using `terraform import`, import an AWS End User Messaging SMS Protect Configuration Association using the `configuration_set_name`. For example:

```console
% terraform import aws_pinpointsmsvoicev2_protect_configuration_association.example example
```
