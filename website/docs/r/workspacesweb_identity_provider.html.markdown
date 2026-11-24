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

* `identity_provider_details` - (Required) Identity provider details. The keys depend on `identity_provider_type`: `Google` and `LoginWithAmazon` use `client_id`, `client_secret`, and `authorize_scopes`. `Facebook` uses `client_id`, `client_secret`, `authorize_scopes`, and `api_version`. `SignInWithApple` uses `client_id`, `team_id`, `key_id`, `private_key`, and `authorize_scopes`. `OIDC` uses `client_id`, `client_secret`, `attributes_request_method`, `oidc_issuer`, and `authorize_scopes`, plus `authorize_url`, `token_url`, `attributes_url`, and `jwks_uri` if they aren't available from the discovery URL in `oidc_issuer`. `SAML` uses `MetadataFile` or `MetadataURL`, and optionally `IDPSignout` (boolean), `IDPInit` (boolean), `RequestSigningAlgorithm` (string, only `rsa-sha256`), and `EncryptedResponses` (boolean).
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
