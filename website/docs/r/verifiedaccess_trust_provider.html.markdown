---
subcategory: "Verified Access"
layout: "aws"
page_title: "AWS: aws_verifiedaccess_trust_provider"
description: |-
  Terraform resource for managing a Verified Access Trust Provider.
---

# Resource: aws_verifiedaccess_trust_provider

Terraform resource for managing a Verified Access Trust Provider.

## Example Usage

```terraform
resource "aws_verifiedaccess_trust_provider" "example" {
  policy_reference_name    = "example"
  trust_provider_type      = "user"
  user_trust_provider_type = "iam-identity-center"
}
```

## Argument Reference

The following arguments are required:

* `policy_reference_name` - (Required) The identifier to be used when working with policy rules.
* `trust_provider_type` - (Required) The type of trust provider can be either user or device-based.

The following arguments are optional:

* `description` - (Optional) A description for the AWS Verified Access trust provider.
* `device_options` - (Optional) A block of options for device identity based trust providers.
* `device_trust_provider_type` (Optional) The type of device-based trust provider.
* `native_application_oidc_options` - (Optional) The OpenID Connect details for an Native Application OIDC, user-identity based trust provider.
* `oidc_options` - (Optional) The OpenID Connect details for an oidc-type, user-identity based trust provider.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_trust_provider_type` - (Optional) The type of user-based trust provider.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the AWS Verified Access trust provider.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Transfer Workflows using the `id`. For example:

```terraform
import {
  to = aws_verifiedaccess_trust_provider.example
  id = "vatp-8012925589"
}
```

Using `terraform import`, import Transfer Workflows using the  `id`. For example:

```console
% terraform import aws_verifiedaccess_trust_provider.example vatp-8012925589
```
