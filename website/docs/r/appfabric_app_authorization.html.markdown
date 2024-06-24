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

* `app` - (Required) The name of the application for valid values see https://docs.aws.amazon.com/appfabric/latest/api/API_CreateAppAuthorization.html.
* `app_bundle_arn` - (Required) The Amazon Resource Name (ARN) of the app bundle to use for the request.
* `auth_type` - (Required) The authorization type for the app authorization valid values are oauth2 and apiKey.
* `credential` - (Required) Contains credentials for the application, such as an API key or OAuth2 client ID and secret.
Specify credentials that match the authorization type for your request. For example, if the authorization type for your request is OAuth2 (oauth2), then you should provide only the OAuth2 credentials.
* `tenant` - (Required) Contains information about an application tenant, such as the application display name and identifier.

Credential support the following:

* `api_key_credential` - (Optional) Contains API key credential information.
* `oauth2_credential` - (Optional) Contains OAuth2 client credential information.

API Key Credential support the following:

* `api_key` - (Required) Contains API key credential information.

oauth2 Credential support the following:

* `client_id` - (Required) The client ID of the client application.
* `client_secret` - (Required) The client secret of the client application.

Tenant support the following:

* `tenant_display_name` - (Required) The display name of the tenant.
* `tenant_identifier` - (Required) The ID of the application tenant.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the App Authorization. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `auth_url` - The application URL for the OAuth flow.
* `persona` - The user persona of the app authorization.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)
