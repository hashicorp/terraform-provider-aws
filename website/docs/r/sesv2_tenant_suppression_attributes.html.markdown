---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_tenant_suppression_attributes"
description: |-
  Manages AWS SESv2 (Simple Email V2) tenant-level suppression attributes.
---

# Resource: aws_sesv2_tenant_suppression_attributes

Manages AWS SESv2 (Simple Email V2) tenant-level suppression attributes.

Use this resource to configure the suppression list preferences for a [tenant](sesv2_tenant.html.markdown). A tenant can either maintain its own suppression list (`TENANT` scope) or use the account-level suppression list (`ACCOUNT` scope). This is the tenant-scoped equivalent of [`aws_sesv2_account_suppression_attributes`](sesv2_account_suppression_attributes.html.markdown).

## Example Usage

```terraform
resource "aws_sesv2_tenant" "example" {
  tenant_name = "example-tenant"
}

resource "aws_sesv2_tenant_suppression_attributes" "example" {
  tenant_name        = aws_sesv2_tenant.example.tenant_name
  suppression_scope  = "TENANT"
  suppressed_reasons = ["BOUNCE", "COMPLAINT"]
}
```

## Argument Reference

The following arguments are required:

* `tenant_name` - (Required) Name of the tenant to configure suppression list preferences for. Changing the tenant name forces creation of a new resource.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `suppressed_reasons` - (Optional) Reasons that email addresses are automatically added to the suppression list for the tenant. Only meaningful when `suppression_scope` is `TENANT`; when the scope is `ACCOUNT` the tenant uses the account-level suppression list. Valid values: `COMPLAINT`, `BOUNCE`.
* `suppression_scope` - (Optional) Suppression scope for the tenant. Specify `TENANT` to use the tenant's own suppression list, or `ACCOUNT` to use the account-level suppression list. If not specified, the tenant defaults to `ACCOUNT` scope.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_sesv2_tenant_suppression_attributes.example
  identity = {
    tenant_name = "example-tenant"
  }
}
```

### Identity Schema

#### Required

* `tenant_name` (String) Name of the tenant.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import tenant-level suppression attributes using the `tenant_name`. For example:

```terraform
import {
  to = aws_sesv2_tenant_suppression_attributes.example
  id = "example-tenant"
}
```

Using `terraform import`, import tenant-level suppression attributes using the `tenant_name`. For example:

```console
% terraform import aws_sesv2_tenant_suppression_attributes.example example-tenant
```
