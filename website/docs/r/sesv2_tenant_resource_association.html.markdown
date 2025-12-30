---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_tenant_resource_association"
description: |-
  Manages an AWS SESv2 (Simple Email V2) Tenant Resource Association.
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

# Resource: aws_sesv2_tenant_resource_association

Manages an AWS SESv2 (Simple Email V2) Tenant Resource Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_tenant_resource_association" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Tenant Resource Association.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SESv2 (Simple Email V2) Tenant Resource Association using the `example_id_arg`. For example:

```terraform
import {
  to = aws_sesv2_tenant_resource_association.example
  id = "tenant_resource_association-id-12345678"
}
```

Using `terraform import`, import SESv2 (Simple Email V2) Tenant Resource Association using the `example_id_arg`. For example:

```console
% terraform import aws_sesv2_tenant_resource_association.example tenant_resource_association-id-12345678
```
