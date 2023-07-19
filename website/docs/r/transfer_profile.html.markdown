---
subcategory: "Transfer Family"
layout: "aws"
page_title: "AWS: aws_transfer_profile"
description: |-
  Provides a AWS Transfer AS2 Profile Resource
---

# Resource: aws_transfer_profile

Provides a AWS Transfer AS2 Profile resource.

## Example Usage

### Basic

```terraform
resource "aws_transfer_profile" "example" {
  as2_id          = "example"
  certificate_ids = [aws_transfer_certificate.example.certificate_id]
  usage           = "LOCAL"
}
```

## Argument Reference

This resource supports the following arguments:

* `as2_id` - (Required) The As2Id is the AS2 name as defined in the RFC 4130. For inbound ttransfers this is the AS2 From Header for the AS2 messages sent from the partner. For Outbound messages this is the AS2 To Header for the AS2 messages sent to the partner. his ID cannot include spaces.
* `certificate_ids` - (Optional) The list of certificate Ids from the imported certificate operation.
* `profile_type` - (Required) The profile type should be LOCAL or PARTNER.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `profile_id`  - The unique identifier for the AS2 profile

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Transfer AS2 Profile using the `profile_id`. For example:

```terraform
import {
  to = aws_transfer_profile.example
  id = "p-4221a88afd5f4362a"
}
```

Using `terraform import`, import Transfer AS2 Profile using the `profile_id`. For example:

```console
% terraform import aws_transfer_profile.example p-4221a88afd5f4362a
```
