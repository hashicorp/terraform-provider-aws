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

* `policy_reference_name` - (Required) Identifier to be used when working with policy rules.
* `trust_provider_type` - (Required) Type of trust provider can be either user or device-based.

The following arguments are optional:

* `description` - (Optional) Description for the AWS Verified Access trust provider.
* `device_options` - (Optional) Block of options for device identity based trust providers. [See below](#device_options-block).
* `device_trust_provider_type` - (Optional) Type of device-based trust provider.
* `native_application_oidc_options` - (Optional) OpenID Connect details for a Native Application OIDC, user-identity based trust provider. [See below](#native_application_oidc_options-block).
* `oidc_options` - (Optional) OpenID Connect details for an oidc-type, user-identity based trust provider. [See below](#oidc_options-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sse_specification` - (Optional) Block of options in use for server side encryption. [See below](#sse_specification-block).
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_trust_provider_type` - (Optional) Type of user-based trust provider.

### `device_options` Block

* `tenant_id` - (Optional) ID of the tenant application with the device-identity provider.

### `native_application_oidc_options` Block

* `authorization_endpoint` - (Optional) OIDC authorization endpoint.
* `client_id` - (Optional) OAuth 2.0 client identifier.
* `client_secret` - (Required) OAuth 2.0 client secret.
* `issuer` - (Optional) OIDC issuer identifier of the IdP.
* `public_signing_key_endpoint` - (Optional) OIDC public signing key endpoint.
* `scope` - (Optional) OpenID Connect (OIDC) scope specified.
* `token_endpoint` - (Optional) OIDC token endpoint.
* `user_info_endpoint` - (Optional) OIDC user info endpoint.

### `oidc_options` Block

* `authorization_endpoint` - (Optional) OIDC authorization endpoint.
* `client_id` - (Optional) OAuth 2.0 client identifier.
* `client_secret` - (Required) OAuth 2.0 client secret.
* `issuer` - (Optional) OIDC issuer identifier of the IdP.
* `scope` - (Optional) OpenID Connect (OIDC) scope specified.
* `token_endpoint` - (Optional) OIDC token endpoint.
* `user_info_endpoint` - (Optional) OIDC user info endpoint.

### `sse_specification` Block

* `customer_managed_key_enabled` - (Optional) Whether a customer managed key is in use.
* `kms_key_arn` - (Optional) ARN of the KMS key.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the AWS Verified Access trust provider.

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
