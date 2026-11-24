---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_openid_connect_provider"
description: |-
  Provides an IAM OpenID Connect provider.
---

# Resource: aws_iam_openid_connect_provider

Provides an IAM OpenID Connect provider.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_openid_connect_provider" "default" {
  url = "https://accounts.google.com"

  client_id_list = [
    "266362248691-342342xasdasdasda-apps.googleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]
}
```

### Without A Thumbprint

```terraform
resource "aws_iam_openid_connect_provider" "default" {
  url = "https://accounts.google.com"

  client_id_list = [
    "266362248691-342342xasdasdasda-apps.googleusercontent.com",
  ]
}
```

## Argument Reference

This resource supports the following arguments:

* `url` - (Required) URL of the identity provider, corresponding to the `iss` claim.
* `client_id_list` - (Required) List of client IDs (audiences) that identify the application registered with the OpenID Connect provider. This is the value sent as the `client_id` parameter in OAuth requests.
* `context_provider` - (Optional) Configures how claims in the OIDC token become session context, enabling [trusted identity propagation](https://aws.amazon.com/blogs/security/know-the-identity-behind-the-role-authorize-and-audit-with-trusted-identity-propagation/). See [`context_provider`](#context_provider) below.
* `identity_center_instance_arn` - (Optional) ARN of the IAM Identity Center instance that validates the identity. Requires `context_provider`.
* `thumbprint_list` - (Optional) List of server certificate thumbprints for the OpenID Connect (OIDC) identity provider's server certificate(s). For certain OIDC identity providers (e.g., Auth0, GitHub, GitLab, Google, or those using an Amazon S3-hosted JWKS endpoint), AWS relies on its own library of trusted root certificate authorities (CAs) for validation instead of using any configured thumbprints. In these cases, any configured `thumbprint_list` is retained in the configuration but not used for verification. For other IdPs, if no `thumbprint_list` is provided, IAM automatically retrieves and uses the top intermediate CA thumbprint from the OIDC IdP server certificate. However, if a `thumbprint_list` is initially configured and later removed, Terraform does not prompt IAM to retrieve a thumbprint the same way. Instead, it continues using the original thumbprint list from the initial configuration. This differs from the behavior when creating an `aws_iam_openid_connect_provider` without a `thumbprint_list`.
* `tags` - (Optional) Map of resource tags for the IAM OIDC provider. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### context_provider

* `key_definition` - (Optional) Mappings from keys in the OIDC token to trusted context keys. Up to 10 blocks. See [`key_definition`](#key_definition) below.
* `subject_source_key` - (Required) Key in the OIDC token that supplies the subject.
* `subject_type_source_key` - (Optional) Key in the OIDC token that supplies the subject type. Conflicts with `subject_type_value`.
* `subject_type_value` - (Optional) Static value for the subject type, such as `user` or `workload`. Conflicts with `subject_type_source_key`.

### key_definition

* `source_key` - (Required) Key in the OIDC token to read the value from.
* `target_key` - (Optional) Trusted context key to map the value into.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN assigned by AWS for this provider.
* `identity_store_arn` - ARN of the identity store associated with the IAM Identity Center instance.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_iam_openid_connect_provider.example
  identity = {
    "arn" = "arn:aws:iam::123456789012:oidc-provider/example.com"
  }
}

resource "aws_iam_openid_connect_provider" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the IAM OpenID Connect provider.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM OpenID Connect Providers using the `arn`. For example:

```terraform
import {
  to = aws_iam_openid_connect_provider.default
  id = "arn:aws:iam::123456789012:oidc-provider/accounts.google.com"
}
```

Using `terraform import`, import IAM OpenID Connect Providers using the `arn`. For example:

```console
% terraform import aws_iam_openid_connect_provider.default arn:aws:iam::123456789012:oidc-provider/accounts.google.com
```
