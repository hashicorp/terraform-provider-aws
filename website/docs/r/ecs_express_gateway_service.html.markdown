---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_express_gateway_service"
description: |-
  Manages an AWS ECS (Elastic Container) Express Gateway Service.
---
<!---
Documentation guidelines:
- Begin resource descriptions with "Manages..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Resource: aws_ecs_express_gateway_service

Manages an AWS ECS (Elastic Container) Express Gateway Service.

## Example Usage

### Basic Usage

```terraform
resource "aws_ecs_express_gateway_service" "example" {
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

* `arn` - ARN of the Express Gateway Service.
* `example_attribute` - Brief description of the attribute.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS (Elastic Container) Express Gateway Service using the `example_id_arg`. For example:

```terraform
import {
  to = aws_ecs_express_gateway_service.example
  id = "express_gateway_service-id-12345678"
}
```

Using `terraform import`, import ECS (Elastic Container) Express Gateway Service using the `example_id_arg`. For example:

```console
% terraform import aws_ecs_express_gateway_service.example express_gateway_service-id-12345678
```
