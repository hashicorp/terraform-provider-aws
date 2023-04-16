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
resource "aws_verifiedaccess_instance" "example" {
  description = "example"
}

resource "aws_verifiedaccess_trust_provider" "example" {
  policy_reference_name    = "example"
  trust_provider_type      = "user"
  user_trust_provider_type = "iam-identity-center"
}

resource "aws_verifiedaccess_trust_provider_attachment" "example" {
  verified_access_instance_id       = aws_verifiedaccess_instance.example.id
  verified_access_trust_provider_id = aws_verifiedaccess_trust_provider.example.id
}
```

## Argument Reference

The following arguments are required:

* `verified_access_instance_id` - (Required) The ID of the Verified Access instance.
* `verified_access_trust_provider_id` - (Required) The ID of the Trust Provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Composite ID for the resource. The `verified_access_trust_provider_id` and `verified_access_instance_id` separated by a `/`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

Verified Access Trust Provider Attachments can be imported using the `verified_access_trust_provider_id` and `verified_access_instance_id` separated by a `/`:, e.g.,

```
$ terraform import aws_verifiedaccess_trust_provider_attachment.example vatp-8012925589/vai-9855292108
```
