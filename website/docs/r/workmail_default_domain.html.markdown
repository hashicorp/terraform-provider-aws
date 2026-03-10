---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_default_domain"
description: |-
  Manages the default mail domain for an AWS WorkMail organization.
---

# Resource: aws_workmail_default_domain

Manages the default mail domain for an AWS WorkMail organization.

~> On destroy, the default domain is reset to the auto-provisioned test domain. If the organization or test domain no longer exists, the resource is simply removed from state.

## Example Usage

### Basic Usage

```terraform
resource "aws_workmail_domain" "example" {
  organization_id = aws_workmail_organization.example.id
  domain_name     = "example.com"
}

resource "aws_workmail_default_domain" "example" {
  organization_id = aws_workmail_organization.example.id
  domain_name     = aws_workmail_domain.example.domain_name
}
```

## Argument Reference

This resource supports the following arguments:

* `domain_name` - (Required) Mail domain name to set as the default.
* `organization_id` - (Required) Identifier of the WorkMail organization. Changing this forces a new resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - WorkMail organization ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkMail Default Domain using the organization ID. For example:

```terraform
import {
  to = aws_workmail_default_domain.example
  id = "m-1234567890abcdef0"
}
```

Using `terraform import`, import WorkMail Default Domain using the organization ID. For example:

```console
% terraform import aws_workmail_default_domain.example "m-1234567890abcdef0"
```
