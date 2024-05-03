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

* `url` - (Required) The URL of the identity provider. Corresponds to the _iss_ claim.
* `client_id_list` - (Required) A list of client IDs (also known as audiences). When a mobile or web app registers with an OpenID Connect provider, they establish a value that identifies the application. (This is the value that's sent as the client_id parameter on OAuth requests.)
* `thumbprint_list` - (Optional) A list of server certificate thumbprints for the OpenID Connect (OIDC) identity provider's server certificate(s).

  AWS secures communication with some OIDC identity providers (IdPs) through a library of trusted root certificate authorities (CAs) instead of using a certificate thumbprint to verify the IdP server certificate. In these cases, a specified thumbprint list remains in the configuration, but is not used for validation. These OIDC IdPs include Auth0, GitHub, GitLab, Google, and those that use an Amazon S3 bucket to host a JSON Web Key Set (JWKS) endpoint.

  If it is not specified, and the IdP is not included in the aforementioned group, IAM will retrieve and use the top intermediate certificate authority (CA) thumbprint of the OIDC identity provider server certificate.
* `tags` - (Optional) Map of resource tags for the IAM OIDC provider. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN assigned by AWS for this provider.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
