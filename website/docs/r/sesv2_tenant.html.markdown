---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_tenant"
description: |-
  Manages an AWS SESv2 (Simple Email V2) Tenant.
---

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

* `region` - Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - Map of tags to assign to the tenant.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `sending_status` – Current sending status of the tenant.
* `tags_all` – Map of tags assigned to the tenant, including provider default tags.
* `tenant_arn` - ARN of the Tenant.
* `tenant_id` - ID of the Tenant.

## Import

In Terraform v1.5.0 and later, use an import block to import an SESv2 tenant using the tenant name.

For example:

```terraform
import {
  to = aws_sesv2_tenant.example
  id = "example-tenant"
}
```

Using `terraform import`, import an SESv2 Tenant using the `tenant_name`. For example:

```console
% terraform import aws_sesv2_tenant.example example-tenant
```
