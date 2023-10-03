---
subcategory: "Verified Access"
layout: "aws"
page_title: "AWS: aws_verifiedaccess_trust_provider_attachment"
description: |-
  Terraform resource for managing a Verified Access Trust Provider Attachment.
---

# Resource: aws_verifiedaccess_trust_provider_attachment

Terraform resource for managing a Verified Access Trust Provider Attachment.

## Example Usage

```terraform
resource "aws_verifiedaccess_instance" "example" {}

resource "aws_verifiedaccess_trust_provider" "example" {
  device_trust_provider_type = "jamf"
  policy_reference_name      = "example"
  trust_provider_type        = "device"

  device_options {
    tenant_id = "example"
  }
}

resource "aws_verifiedaccess_trust_provider_attachment" "example" {
  instance_id       = aws_verifiedaccess_instance.example.id
  trust_provider_id = aws_verifiedaccess_trust_provider.example.id
}
```

## Argument Reference

The following arguments are required:

* `instance_id` - (Required) The ID of the Verified Access instance to attach the Trust Provider to.
* `trust_provider_id` - (Required) The ID of the Verified Access trust provider.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A combination of attributes, separated by a `/` to create a unique id: `instance_id`,`trust_provider_id`

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Verified Access Trust Provider Attachments using the `instance_id` and `trust_provider_id` separated by a forward slash (`/`). For example:

```terraform
import {
  to = aws_verifiedaccess_trust_provider_attachment.example
  id = "vai-1234567890abcdef0/vatp-8012925589"
}
```

Using `terraform import`, import Verified Access Trust Provider Attachments using the `instance_id` and `trust_provider_id` separated by a forward slash (`/`). For example:

```console
% terraform import aws_verifiedaccess_trust_provider_attachment.example vai-1234567890abcdef0/vatp-8012925589
```
