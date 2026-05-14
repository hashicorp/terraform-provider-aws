---
subcategory: "AppFabric"
layout: "aws"
page_title: "AWS: aws_appfabric_app_authorization"
description: |-
  Terraform resource for managing an AWS AppFabric App Authorization.
---

# Resource: aws_appfabric_app_authorization

Terraform resource for managing an AWS AppFabric App Authorization.

## Example Usage

### Basic Usage

```terraform
resource "aws_appfabric_app_authorization" "example" {
  app            = "TERRAFORMCLOUD"
  app_bundle_arn = aws_appfabric_app_bundle.arn
  auth_type      = "apiKey"

  credential {
    api_key_credential {
      api_key = "exampleapikeytoken"
    }
  }
  tenant {
    tenant_display_name = "example"
    tenant_identifier   = "example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `app` - (Required) Name of the application. For valid values see the [CreateAppAuthorization API](https://docs.aws.amazon.com/appfabric/latest/api/API_CreateAppAuthorization.html).
* `app_bundle_arn` - (Required) ARN of the app bundle to use for the request.
* `auth_type` - (Required) Authorization type for the app authorization. Valid values are `oauth2` and `apiKey`.
* `credential` - (Required) Credentials for the application, such as an API key or OAuth2 client ID and secret. See [`credential` Block](#credential-block) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tenant` - (Required) Information about the application tenant. See [`tenant` Block](#tenant-block) below.

### `credential` Block

The `credential` block supports the following arguments:

* `api_key_credential` - (Optional) API key credential information. See [`api_key_credential` Block](#api_key_credential-block) below.
* `oauth2_credential` - (Optional) OAuth2 client credential information. See [`oauth2_credential` Block](#oauth2_credential-block) below.

### `api_key_credential` Block

The `api_key_credential` block supports the following arguments:

* `api_key` - (Required) API key to use for authentication.

### `oauth2_credential` Block

The `oauth2_credential` block supports the following arguments:

* `client_id` - (Required) Client ID of the client application.
* `client_secret` - (Required) Client secret of the client application.

### `tenant` Block

The `tenant` block supports the following arguments:

* `tenant_display_name` - (Required) Display name of the tenant.
* `tenant_identifier` - (Required) ID of the application tenant.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the App Authorization.
* `auth_url` - Application URL for the OAuth flow.
* `persona` - User persona of the app authorization.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)
