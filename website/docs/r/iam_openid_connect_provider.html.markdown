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

### With A Thumbprint

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
* `thumbprint_list` - (Optional) List of server certificate thumbprints for the OpenID Connect (OIDC) identity provider's server certificate(s). For certain OIDC identity providers (_e.g._, Auth0, GitHub, GitLab, Google, or those using an Amazon S3 bucket to host a JSON Web Key Set [JWKS] endpoint), AWS uses a library of trusted root certificate authorities (CAs) instead of the thumbprint for validation. In these cases, the specified thumbprint list is retained in the configuration but not used for verification. If no thumbprint list is provided and the IdP is not in this group, IAM retrieves and uses the top intermediate CA thumbprint of the OIDC IdP server certificate.
* `tags` - (Optional) Map of resource tags for the IAM OIDC provider. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN assigned by AWS for this provider.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

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
