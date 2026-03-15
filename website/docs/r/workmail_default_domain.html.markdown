---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_default_domain"
description: |-
  Manages the default mail domain for an AWS WorkMail organization.
---

# Resource: aws_workmail_default_domain

Manages the default mail domain for an AWS WorkMail organization.

~> **NOTE:** This does not register a domain for workmail. This resource requires a verified domain name to be used as default domain for workmail organization.

## Example Usage

### Basic Usage

```terraform
resource "aws_workmail_organization" "example" {
  organization_alias = "example-org"
}

resource "aws_workmail_default_domain" "example" {
  organization_id = aws_workmail_organization.example.id
  domain_name     = aws_workmail_organization.example.default_mail_domain
}
```

## Argument Reference

This resource supports the following arguments:

* `domain_name` - (Required) Mail domain name to set as the default.
* `organization_id` - (Required) Identifier of the WorkMail organization. Changing this forces a new resource.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

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
