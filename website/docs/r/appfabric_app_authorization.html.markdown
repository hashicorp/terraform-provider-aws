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

The following arguments are required:

* `app` - (Required) Name of the application. For valid values, see the [CreateAppAuthorization API reference](https://docs.aws.amazon.com/appfabric/latest/api/API_CreateAppAuthorization.html).
* `app_bundle_arn` - (Required) Amazon Resource Name (ARN) of the app bundle to use for the request.
* `auth_type` - (Required) Authorization type for the app authorization. Valid values are `oauth2` and `apiKey`.
* `credential` - (Required) Credentials for the application, such as an API key or OAuth2 client ID and secret. Specify credentials that match the authorization type for your request. For example, if the authorization type for your request is OAuth2 (`oauth2`), then you should provide only the OAuth2 credentials. See [`credential` Block](#credential-block) for details.
* `tenant` - (Required) Information about an application tenant, such as the application display name and identifier. See [`tenant` Block](#tenant-block) for details.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `credential` Block

The `credential` configuration block supports the following arguments:

* `api_key_credential` - (Optional) API key credential information. See [`api_key_credential` Block](#credential-api_key_credential-block) for details.
* `oauth2_credential` - (Optional) OAuth2 client credential information. See [`oauth2_credential` Block](#credential-oauth2_credential-block) for details.

### `credential.api_key_credential` Block

The `api_key_credential` configuration block supports the following arguments:

* `api_key` - (Required) API key.

### `credential.oauth2_credential` Block

The `oauth2_credential` configuration block supports the following arguments:

* `client_id` - (Required) Client ID of the client application.
* `client_secret` - (Required) Client secret of the client application.

### `tenant` Block

The `tenant` configuration block supports the following arguments:

* `tenant_display_name` - (Required) Display name of the tenant.
* `tenant_identifier` - (Required) ID of the application tenant.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the App Authorization.
* `auth_url` - Application URL for the OAuth flow.
* `created_at` - Timestamp of when the app authorization was created.
* `persona` - User persona of the app authorization.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Timestamp of when the app authorization was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)
