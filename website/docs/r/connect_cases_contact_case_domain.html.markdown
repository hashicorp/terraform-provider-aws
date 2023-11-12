---
subcategory: "Connect Cases"
layout: "aws"
page_title: "AWS: aws_connectcases_domain":
description: |-
  Terraform resource for managing an Amazon Connect Cases Domain.
---

# Resource: aws_connectcases_domain

Terraform resource for managing an Amazon Connect Cases Domain.
See the [Create Domain](https://docs.aws.amazon.com/cases/latest/APIReference/API_CreateDomain.html) for more information.

## Example Usage

```terraform
resource "aws_connectcases_domain" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `name` - The name for your Cases domain. It must be unique for your AWS account.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `domain_arn` - The Amazon Resource Name (ARN) for the Cases domain.
* `domain_id` - The unique identifier of the Cases domain.
* `domain_status` - The status of the domain.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Connect Cases Domain using the resource `id`. For example:

```terraform
import {
  to = aws_connectcases_domain.example
  id = "e6f777be-22d0-4b40-b307-5d2720ef16b2"
}
```

Using `terraform import`, import Amazon Connect Cases Domain using the resource `id`. For example:

```console
% terraform import aws_connectcases_domain.example e6f777be-22d0-4b40-b307-5d2720ef16b2
```
