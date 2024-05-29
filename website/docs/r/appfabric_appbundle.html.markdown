---
subcategory: "AppFabric"
layout: "aws"
page_title: "AWS: aws_appfabric_appbundle"
description: |-
  Terraform resource for managing an AWS AppFabric AppBundle.
---

# Resource: aws_appfabric_appbundle

Terraform resource for managing an AWS AppFabric AppBundle.

## Example Usage

### Basic Usage

```terraform
resource "aws_appfabric_appbundle" "example" {
	customer_managed_key_identifier = "[KMS CMK ARN]"
	tags = {
    Environment = "test"
  }
}
```

## Argument Reference

The following arguments are required:

* There are no required arguments.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the AppBundle. 

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import
In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppFabric AppBundle using the `arn`. For example:

```terraform
import {
  to = aws_appfabric_appbundle.example
  id = "arn:aws:appfabric:[region]:[account]:appbundle/ee5587b4-5765-4288-a202-xxxxxxxxxx"
}
```

Using `terraform import`, import AppFabric AppBundle using the `arn`. For example:

```console
% terraform import aws_appfabric_appbundle.example arn:aws:appfabric:[region]:[account]:appbundle/ee5587b4-5765-4288-a202-xxxxxxxxxx
```