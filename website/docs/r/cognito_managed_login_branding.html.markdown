---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_managed_login_branding"
description: |-
  Terraform resource for managing an AWS Cognito IDP (Identity Provider) Managed Login Branding.
---
# Resource: aws_cognito_managed_login_branding

Terraform resource for managing an AWS Cognito IDP (Identity Provider) Managed Login Branding.

## Example Usage

### Use cognito provided values

```terraform
resource "aws_cognito_managed_login_branding" "example" {
  client_id                   = aws_cognito_user_pool_client.example.id
  user_pool_id                = aws_cognito_user_pool.example.id
  use_cognito_provided_values = true

  asset {
    bytes      = filebase64("login_branding_asset.svg")
    category   = "PAGE_HEADER_BACKGROUND"
    color_mode = "DARK"
    extension  = "SVG"
  }
}
```

### Use your own settings

```terraform
resource "aws_cognito_managed_login_branding" "example" {
  client_id    = aws_cognito_user_pool_client.example.id
  user_pool_id = aws_cognito_user_pool.example.id

  asset {
    bytes      = filebase64("login_branding_asset.svg")
    category   = "PAGE_HEADER_BACKGROUND"
    color_mode = "DARK"
    extension  = "SVG"
  }

  settings = jsonencode({
    // place you setting in JSON
  })
}
```

## Argument Reference

The following arguments are required:

* `client_id` – (Required) App client associated with the branding style.
* `user_pool_id` – (Required) ID of the user pool containing the app client.

The following arguments are optional:

* `asset` – (Optional, ForceNew if changed) Block for image files such as backgrounds, logos, and icons. See [asset](#asset) for details.
* `settings` – (Optional) Style settings as a JSON string. Exactly one of `settings` or `use_cognito_provided_values` must be set.
* `use_cognito_provided_values` – (Optional) Boolean value that determines whether to apply the default branding style options. Exactly one of `settings` or `use_cognito_provided_values` must be set.

### asset

* `bytes` – (Optional) Base64-encoded image data.
* `category` – (Required) Image category. See [AWS Documentation](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_AssetType.html) for valid values.
* `color_mode` – (Required) Display mode for the asset. Valid values: `LIGHT`, `DARK`, `DYNAMIC`.
* `extension` – (Required) Image file type. Valid values: `ICO`, `JPEG`, `PNG`, `SVG`, `WEBP`.
* `resource_id` – (Optional) ID of the asset.

## Attribute Reference

This resource exports the following attributes:

* `creation_date` – Date and time when the resource was created.
* `id` – Composite ID formed by joining `managed_login_branding_id`, `user_pool_id`, and `client_id` with `|`.
* `last_modified_date` – Date and time when the resource was last modified.
* `managed_login_branding_id` – (Attribute) ID of the managed login branding style.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Cognito IDP Managed Login Branding resource. Use the composite ID, which consists of the managed login branding ID, user pool ID, and client ID, separated by pipes (`|`). Example:

```hcl
import {
  to = aws_cognito_managed_login_branding.example
  id = "1ee76372-aa1d-4030-9b97-b0e86ced64bf|us-east-1_TsleZoGGE|7fpogu27daf3gg86ncehv8bh6h"
}
```

Alternatively, use the `terraform import` CLI command:

```console
terraform import aws_cognito_managed_login_branding.example 1ee76372-aa1d-4030-9b97-b0e86ced64bf|us-east-1_TsleZoGGE|7fpogu27daf3gg86ncehv8bh6h
```
