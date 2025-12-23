---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_tenant"
description: |-
  Manages an AWS SESv2 (Simple Email V2) Tenant.
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

# Resource: aws_sesv2_tenant

Manages an AWS SESv2 (Simple Email V2) Tenant.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_tenant" "example" {
  tenant_name = "example-tenant"

  tags = {
    Environment = "test"
  }
}
```

## Argument Reference

The following arguments are required:

* `tenant_name` - Name of the SESV2 tenant.  The name must be unique within the AWS account and Region.  Changing the tenant name forces creation of a new tenant.

The following arguments are optional:

* `tags` - Map of tags to assign to the tenant.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Tenant.
* `created_timestamp` – Timestamp when the tenant was created.
* `sending_status` – Current sending status of the tenant.
* `tags_all` – Map of tags assigned to the tenant, including provider default tags.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an import block to import an SESv2 tenant using the tenant name.

For example:

```terraform
import {
  to = aws_sesv2_tenant.example
  id = "example-tenant"
}
```

Using `terraform import`, import SESv2 (Simple Email V2) Tenant using the `tenant_name`. For example:

```console
% terraform import aws_sesv2_tenant.example example-tenant
```
