---
subcategory: "Verified Access"
layout: "aws"
page_title: "AWS: aws_verifiedaccess_instance_trust_provider_attachment"
description: |-
  Terraform resource for managing a Verified Access Instance Trust Provider Attachment.
---

# Resource: aws_verifiedaccess_instance_trust_provider_attachment

Terraform resource for managing a Verified Access Instance Trust Provider Attachment.

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

resource "aws_verifiedaccess_instance_trust_provider_attachment" "example" {
  verifiedaccess_instance_id       = aws_verifiedaccess_instance.example.id
  verifiedaccess_trust_provider_id = aws_verifiedaccess_trust_provider.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `verifiedaccess_instance_id` - (Required) The ID of the Verified Access instance to attach the Trust Provider to.
* `verifiedaccess_trust_provider_id` - (Required) The ID of the Verified Access trust provider.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A combination of attributes, separated by a `/` to create a unique id: `verifiedaccess_instance_id`,`verifiedaccess_trust_provider_id`

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Verified Access Instance Trust Provider Attachments using the `verifiedaccess_instance_id` and `verifiedaccess_trust_provider_id` separated by a forward slash (`/`). For example:

```terraform
import {
  to = aws_verifiedaccess_instance_trust_provider_attachment.example
  id = "vai-1234567890abcdef0/vatp-8012925589"
}
```

Using `terraform import`, import Verified Access Instance Trust Provider Attachments using the `verifiedaccess_instance_id` and `verifiedaccess_trust_provider_id` separated by a forward slash (`/`). For example:

```console
% terraform import aws_verifiedaccess_instance_trust_provider_attachment.example vai-1234567890abcdef0/vatp-8012925589
```
