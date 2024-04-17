---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_application"
description: |-
  Terraform resource for managing an AWS SSO Admin Application.
---
# Resource: aws_ssoadmin_application

Terraform resource for managing an AWS SSO Admin Application.

~> The `CreateApplication` API only supports custom OAuth 2.0 applications.
Creation of 3rd party SAML or OAuth 2.0 applications require setup to be done through the associated app service or AWS console.
See [this issue](https://github.com/hashicorp/terraform-provider-aws/issues/34813#issuecomment-1910380297) for additional context.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_application" "example" {
  name                     = "example"
  application_provider_arn = "arn:aws:sso::aws:applicationProvider/custom"
  instance_arn             = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}
```

### With Portal Options

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_application" "example" {
  name                     = "example"
  application_provider_arn = "arn:aws:sso::aws:applicationProvider/custom"
  instance_arn             = tolist(data.aws_ssoadmin_instances.example.arns)[0]

  portal_options {
    visibility = "ENABLED"
    sign_in_options {
      application_url = "http://example.com"
      origin          = "APPLICATION"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `application_provider_arn` - (Required) ARN of the application provider.
* `instance_arn` - (Required) ARN of the instance of IAM Identity Center.
* `name` - (Required) Name of the application.

The following arguments are optional:

* `client_token` - (Optional) A unique, case-sensitive ID that you provide to ensure the idempotency of the request. AWS generates a random value when not provided.
* `description` - (Optional) Description of the application.
* `portal_options` - (Optional) Options for the portal associated with an application. See [`portal_options`](#portal_options-argument-reference) below.
* `status` - (Optional) Status of the application. Valid values are `ENABLED` and `DISABLED`.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `portal_options` Argument Reference

* `sign_in_options` - (Optional) Sign-in options for the access portal. See [`sign_in_options`](#sign_in_options-argument-reference) below.
* `visibility` - (Optional) Indicates whether this application is visible in the access portal. Valid values are `ENABLED` and `DISABLED`.

### `sign_in_options` Argument Reference

* `application_url` - (Optional) URL that accepts authentication requests for an application.
* `origin` - (Required) Determines how IAM Identity Center navigates the user to the target application.
Valid values are `APPLICATION` and `IDENTITY_CENTER`.
If `APPLICATION` is set, IAM Identity Center redirects the customer to the configured `application_url`.
If `IDENTITY_CENTER` is set, IAM Identity Center uses SAML identity-provider initiated authentication to sign the customer directly into a SAML-based application.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `application_account` - AWS account ID.
* `application_arn` - ARN of the application.
* `id` - ARN of the application.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Admin Application using the `id`. For example:

```terraform
import {
  to = aws_ssoadmin_application.example
  id = "arn:aws:sso::012345678901:application/id-12345678"
}
```

Using `terraform import`, import SSO Admin Application using the `id`. For example:

```console
% terraform import aws_ssoadmin_application.example arn:aws:sso::012345678901:application/id-12345678
```
