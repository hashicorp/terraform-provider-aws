---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_capacity_provider"
description: |-
  Manages an AWS Lambda Capacity Provider.
---

# Resource: aws_lambda_capacity_provider

Manages an AWS Lambda Capacity Provider.

## Example Usage

### Basic Usage

```terraform
resource "aws_lambda_capacity_provider" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Capacity Provider.
* `example_attribute` - Brief description of the attribute.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Capacity Provider using the `example_id_arg`. For example:

```terraform
import {
  to = aws_lambda_capacity_provider.example
  id = "capacity_provider-id-12345678"
}
```

Using `terraform import`, import Lambda Capacity Provider using the `example_id_arg`. For example:

```console
% terraform import aws_lambda_capacity_provider.example capacity_provider-id-12345678
```
