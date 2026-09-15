---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_identity_provider"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web Identity Provider.
---

# Resource: aws_workspacesweb_identity_provider

Terraform resource for managing an AWS WorkSpaces Web Identity Provider.

## Example Usage

### Basic Usage with SAML

```terraform
resource "aws_workspacesweb_portal" "example" {
  display_name = "example"
}

resource "aws_workspacesweb_identity_provider" "example" {
  identity_provider_name = "example-saml"
  identity_provider_type = "SAML"
  portal_arn             = aws_workspacesweb_portal.example.portal_arn

  identity_provider_details = {
    MetadataURL = "https://example.com/metadata"
  }
}
```

### OIDC Identity Provider

```terraform
resource "aws_workspacesweb_portal" "test" {
  display_name = "test"
}

resource "aws_workspacesweb_identity_provider" "test" {
  identity_provider_name = "test-updated"
  identity_provider_type = "OIDC"
  portal_arn             = aws_workspacesweb_portal.test.portal_arn

  identity_provider_details = {
    client_id                 = "test-client-id"
    client_secret             = "test-client-secret"
    oidc_issuer               = "https://accounts.google.com"
    attributes_request_method = "POST"
    authorize_scopes          = "openid, email"
  }
}
```

## Argument Reference

The following arguments are required:

* `identity_provider_details` - (Required) Identity provider details. The following list describes the provider detail keys for each identity provider type:
    * For Google and Login with Amazon:
        * `client_id`
        * `client_secret`
        * `authorize_scopes`
    * For Facebook:
        * `client_id`
        * `client_secret`
        * `authorize_scopes`
        * `api_version`
    * For Sign in with Apple:
        * `client_id`
        * `team_id`
        * `key_id`
        * `private_key`
        * `authorize_scopes`
    * For OIDC providers:
        * `client_id`
        * `client_secret`
        * `attributes_request_method`
        * `oidc_issuer`
        * `authorize_scopes`
        * `authorize_url` if not available from discovery URL specified by `oidc_issuer` key
        * `token_url` if not available from discovery URL specified by `oidc_issuer` key
        * `attributes_url` if not available from discovery URL specified by `oidc_issuer` key
        * `jwks_uri` if not available from discovery URL specified by `oidc_issuer` key
    * For SAML providers:
        * `MetadataFile` OR `MetadataURL`
        * `IDPSignout` (boolean) optional
        * `IDPInit` (boolean) optional
        * `RequestSigningAlgorithm` (string) optional - Only accepts rsa-sha256
        * `EncryptedResponses` (boolean) optional
* `identity_provider_name` - (Required) Identity provider name.
* `identity_provider_type` - (Required) Identity provider type. Valid values: `SAML`, `Facebook`, `Google`, `LoginWithAmazon`, `SignInWithApple`, `OIDC`.
* `portal_arn` - (Required) ARN of the web portal. Forces replacement if changed.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `identity_provider_arn` - ARN of the identity provider.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web Identity Provider using the `identity_provider_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_identity_provider.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:identityprovider/abcdef12345678/12345678-1234-1234-1234-123456789012"
}
```

Using `terraform import`, import WorkSpaces Web Identity Provider using the `identity_provider_arn`. For example:

```console
% terraform import aws_workspacesweb_identity_provider.example arn:aws:workspaces-web:us-west-2:123456789012:identityprovider/abcdef12345678/12345678-1234-1234-1234-123456789012
```
