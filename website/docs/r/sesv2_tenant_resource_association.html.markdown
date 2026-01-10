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
  tenant_name  = "example-tenant"
  resource_arn = "arn:aws:ses:us-east-1:123456789012:configuration-set/example"
}
```

## Argument Reference

The following arguments are required:

* `tenant_name` - (Required) Name of SES Tenant.
* `resource_arn` - (Required) ARN of the SES resource to associate with the tenant.

The following arguments are optional:
* `region` - (Optional) AWS region for SESv2 operations. If not specified, the default provider region is used.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier for the tenant resource association.  The ID is a composite value in the form of `tenant_name|resource_arn`

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
  id = "example-tenant|arn:aws:ses:us-east-1:123456789012:configuration-set/example"
}
```

Using `terraform import`, import SESv2 (Simple Email V2) Tenant Resource Association using the `example_id_arg`. For example:

```console
% terraform import aws_sesv2_tenant_resource_association.example example-tenant|arn:aws:ses:us-east-1:123456789012:configuration-set/example

```
