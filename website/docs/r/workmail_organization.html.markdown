---
subcategory: "WorkMail"
layout: "aws"
page_title: "AWS: aws_workmail_organization"
description: |-
  Terraform resource for managing an AWS WorkMail Organization.
---

# Resource: aws_workmail_organization

Terraform resource for managing an AWS WorkMail Organization.

## Example Usage

### Basic Usage

```terraform
resource "aws_workmail_organization" "example" {
  alias = "example"
}
```

## Argument Reference

The following arguments are required:

* `alias` - (Required) Organization alias.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Organization.
* `default_mail_domain` - Default mail domain associated with the organization.
* `directory_id` - Identifier for the directory associated with an WorkMail organization.
* `directory_type` - Type of directory associated with the WorkMail organization.
* `id` - ID of the Organization.
* `state` - State of the Organization.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `read` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

WorkMail Organization can be imported using the `id`, e.g.,

```
$ terraform import aws_workmail_organization.example m-21468bd081e647678f6a1263fed49ba6
```
