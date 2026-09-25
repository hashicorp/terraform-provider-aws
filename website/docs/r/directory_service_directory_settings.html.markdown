---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_ds_directory_service_settings"
description: |-
  Manages an AWS Directory Service Settings.
---

# Resource: aws_ds_directory_service_settings

Manages an AWS Directory Service Settings.

## Example Usage

### Basic Usage

```terraform
resource "aws_ds_directory_service_settings" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Settings.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ds_directory_service_settings.example
  identity = {
  }
}

resource "aws_ds_directory_service_settings" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

#### Optional
* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Directory Service Settings using the `example_id_arg`. For example:

```terraform
import {
  to = aws_ds_directory_service_settings.example
  id = "directory_service_settings-id-12345678"
}
```

Using `terraform import`, import Directory Service Settings using the `example_id_arg`. For example:

```console
% terraform import aws_ds_directory_service_settings.example directory_service_settings-id-12345678
```
