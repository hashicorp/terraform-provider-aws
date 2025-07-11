---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognitoidp_log_delivery_configuration"
description: |-
  Manages an AWS Cognito IDP (Identity Provider) Log Delivery Configuration.
---

# Resource: aws_cognitoidp_log_delivery_configuration

Manages an AWS Cognito IDP (Identity Provider) Log Delivery Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_cognitoidp_log_delivery_configuration" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Log Delivery Configuration.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cognito IDP (Identity Provider) Log Delivery Configuration using the `example_id_arg`. For example:

```terraform
import {
  to = aws_cognitoidp_log_delivery_configuration.example
  id = "log_delivery_configuration-id-12345678"
}
```

Using `terraform import`, import Cognito IDP (Identity Provider) Log Delivery Configuration using the `example_id_arg`. For example:

```console
% terraform import aws_cognitoidp_log_delivery_configuration.example log_delivery_configuration-id-12345678
```
