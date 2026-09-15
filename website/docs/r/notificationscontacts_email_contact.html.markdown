---
subcategory: "User Notifications Contacts"
layout: "aws"
page_title: "AWS: aws_notificationscontacts_email_contact"
description: |-
  Terraform resource for managing an AWS User Notifications Contacts Email Contact.
---

# Resource: aws_notificationscontacts_email_contact

Terraform resource for managing AWS User Notifications Contacts Email Contact.

## Example Usage

### Basic Usage

```terraform
resource "aws_notificationscontacts_email_contact" "example" {
  name          = "example-contact"
  email_address = "example@example.com"

  tags = {
    Environment = "Production"
  }
}
```

## Argument Reference

The following arguments are required:

* `email_address` - (Required) Email address for the contact. Must be between 6 and 254 characters and match an email
  pattern.
* `name` - (Required) Name of the email contact. Must be between 1 and 64 characters and can contain alphanumeric
  characters, underscores, tildes, periods, and hyphens.

The following arguments are optional:

* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [
  `default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block)
  present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Email Contact.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [
  `default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to
import User Notifications Contacts Email Contact using the `arn`. For example:

```terraform
import {
  to = aws_notificationscontacts_email_contact.example
  id = "arn:aws:notificationscontacts:us-west-2:123456789012:emailcontact:example-contact"
}
```

Using `terraform import`, import User Notifications Contacts Email Contact using the `arn`. For example:

```console
% terraform import aws_notificationscontacts_email_contact.example arn:aws:notificationscontacts:us-west-2:123456789012:emailcontact:example-contact
```
