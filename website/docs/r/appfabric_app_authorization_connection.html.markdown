---
subcategory: "AppFabric"
layout: "aws"
page_title: "AWS: aws_appfabric_app_authorization_connection"
description: |-
  Terraform resource for managing an AWS AppFabric App Authorization Connection.
---

# Resource: aws_appfabric_app_authorization_connection

Terraform resource for managing an AWS AppFabric App Authorization Connection.

## Example Usage

### Basic Usage

```terraform
resource "aws_appfabric_app_authorization_connection" "example" {
  app_authorization_arn = aws_appfabric_app_authorization.test.arn
  app_bundle_arn        = aws_appfabric_app_bundle.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `app_bundle_arn` - (Required) The Amazon Resource Name (ARN) of the app bundle to use for the request.
* `app_authorization_arn` - (Required) The Amazon Resource Name (ARN) or Universal Unique Identifier (UUID) of the app authorization to use for the request.
* `auth_request` - (Optional) Contains OAuth2 authorization information.This is required if the app authorization for the request is configured with an OAuth2 (oauth2) authorization type.

Auth Request support the following:

* `code` - (Required) The authorization code returned by the application after permission is granted in the application OAuth page (after clicking on the AuthURL)..
* `redirect_uri` - (Optional) The redirect URL that is specified in the AuthURL and the application client.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `app` - The name of the application.
* `tenant` - Contains information about an application tenant, such as the application display name and identifier.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
