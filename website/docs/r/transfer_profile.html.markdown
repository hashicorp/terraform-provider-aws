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

* `as2_id` - (Required) AS2 name as defined in RFC 4130. For inbound transfers this is the AS2 From Header for the AS2 messages sent from the partner. For outbound messages this is the AS2 To Header for the AS2 messages sent to the partner. This ID cannot include spaces.
* `certificate_ids` - (Optional) List of certificate IDs from the imported certificate operation.
* `profile_type` - (Required) Profile type. Valid values are `LOCAL` or `PARTNER`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the profile.
* `profile_id` - Unique identifier for the AS2 profile.

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
