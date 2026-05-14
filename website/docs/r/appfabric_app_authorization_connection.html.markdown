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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `app_authorization_arn` - (Required) ARN of the app authorization to use for the request.
* `app_bundle_arn` - (Required) ARN of the app bundle to use for the request.
* `auth_request` - (Optional) OAuth2 authorization information. Required when the app authorization is configured with OAuth2. See [`auth_request` Block](#auth_request-block) below.

### `auth_request` Block

The `auth_request` block supports the following arguments:

* `code` - (Required) Authorization code returned by the application after permission is granted in the application OAuth page.
* `redirect_uri` - (Required) Redirect URL specified in the AuthURL and the application client.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `app` - Name of the application.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `tenant` - Information about the application tenant. See [`tenant` Block](#tenant-block) below.

### `tenant` Block

* `tenant_display_name` - Display name of the tenant.
* `tenant_identifier` - ID of the application tenant.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
