---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_managed_login_branding"
description: |-
  Manages branding settings for a user pool style and associates it with an app client.
---

# Resource: aws_cognito_managed_login_branding

Manages branding settings for a user pool style and associates it with an app client.

## Example Usage

### Default Branding Style

```terraform
resource "aws_cognito_managed_login_branding" "client" {
  client_id    = aws_cognito_user_pool_client.example.id
  user_pool_id = aws_cognito_user_pool.example.id

  use_cognito_provided_values = true
}
```

### Custom Branding Style

```terraform
resource "aws_cognito_managed_login_branding" "client" {
  client_id    = aws_cognito_user_pool_client.example.id
  user_pool_id = aws_cognito_user_pool.example.id

  asset {
    bytes      = filebase64("login_branding_asset.svg")
    category   = "PAGE_HEADER_BACKGROUND"
    color_mode = "DARK"
    extension  = "SVG"
  }

  settings = jsonencode({
    # Your settings here.
  })
}
```

## Argument Reference

The following arguments are required:

* `client_id` - (Required) App client that the branding style is for.
* `user_pool_id` - (Required) User pool the client belongs to.

The following arguments are optional:

* `asset` - (Optional) Image files to apply to roles like backgrounds, logos, and icons. See [details below](#asset).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `settings` - (Optional) JSON document with the the settings to apply to the style.
* `use_cognito_provided_values` - (Optional) When `true`, applies the default branding style options.

### asset

* `bytes` - (Optional) Image file, in Base64-encoded binary.
* `category` - (Required) Category that the image corresponds to. See [AWS documentation](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_AssetType.html#CognitoUserPools-Type-AssetType-Category) for valid values.
* `color_mode` - (Required) Display-mode target of the asset. Valid values: `LIGHT`, `DARK`, `DYNAMIC`.
* `extensions` - (Required) File type of the image file. See [AWS documentation](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_AssetType.html#CognitoUserPools-Type-AssetType-Extension) for valid values.
* `resource_id` - (Optional) Asset ID.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `managed_login_branding_id` - ID of the managed login branding style.
* `settings_all` - Settings including Amazon Cognito defaults.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cognito branding settings using `user_pool_id` and `managed_login_branding_id` separated by `,`. For example:

```terraform
import {
  to = aws_cognito_managed_login_branding.example
  id = "us-west-2_rSss9Zltr,06c6ae7b-1e66-46d2-87a9-1203ea3307bd"
}
```

Using `terraform import`, import Cognito branding settings using `user_pool_id` and `managed_login_branding_id` separated by `,`. For example:

```console
% terraform import aws_cognito_managed_login_branding.example us-west-2_rSss9Zltr,06c6ae7b-1e66-46d2-87a9-1203ea3307bd
```
