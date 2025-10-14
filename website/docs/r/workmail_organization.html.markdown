---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_organization"
description: |-
  Manages an AWS WorkMail Organization.
---

# Resource: aws_workmail_organization

Manages an AWS WorkMail Organization.

## Example Usage

### Basic Usage

```terraform
resource "aws_workmail_organization" "example" {
  alias = "example"
}
```

## Argument Reference

The following arguments are required:

* `alias` - (Required) A unique alias for the organisation.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Organization.

The resource has additional options, which are not supported yet.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkMail Organization using the `id`, or the organization id. For example:

```terraform
import {
  to = aws_workmail_organization.example
  id = "organization-id-12345678"
}
```

Using `terraform import`, import WorkMail Organization using the organization id. For example:

```console
% terraform import aws_workmail_organization.example organization-id-12345678
```
