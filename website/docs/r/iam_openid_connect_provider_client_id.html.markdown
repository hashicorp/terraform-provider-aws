---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_openid_connect_provider_client_id"
description: |-
  Manages an AWS IAM (Identity & Access Management) Open ID Connect Provider Client ID list.
---

<!---
Documentation guidelines:
- Begin resource descriptions with "Manages..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Resource: aws_iam_openid_connect_provider_client_id

Adds a Client ID to an AWS IAM (Identity & Access Management) Open ID Connect Provider.

~> **NOTE:** The usage of this resource conflicts with the `aws_iam_openid_connect_provider` resource `client_id_list` parameter.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_openid_connect_provider" "example" {
  url = "https://accounts.google.com"
}

resource "aws_iam_openid_connect_provider_client_id" "example" {
  client_id                   = "266362248691-342342xasdasdasda-apps.googleusercontent.com"
  openid_connect_provider_arn = aws_iam_openid_connect_provider.example.arn
}
```

## Argument Reference

The following arguments are required:

- `client_id` - (Required) Client ID to add to the provider's list.

- `openid_connect_provider_arn` - (Required) ARN of the IAM OpenID Connect provider.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_iam_openid_connect_provider_client_id.example
  identity = {
    openid_connect_provider_arn = "arn:aws:iam::11111111111:oidc-provider/app.eu.terraform.io"
    client_id                   = "266362248691-342342xasdasdasda-apps.googleusercontent.com"
  }
}

resource "aws_iam_openid_connect_provider_client_id" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `openid_connect_provider_arn` - ARN of the Open ID Connect Provider.
- `client_id` - Client ID argument of the Open ID Connect Provider Client ID list.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM (Identity & Access Management) Open ID Connect Provider Client ID using the `example_id_arg`. For example:

```terraform
import {
  to = aws_iam_openid_connect_provider_client_id.example
  id = "arn:aws:iam::11111111111:oidc-provider/app.eu.terraform.io|266362248691-342342xasdasdasda-apps.googleusercontent.com"
}
```

Using `terraform import`, import IAM (Identity & Access Management) Open ID Connect Provider Client ID using the `example_id_arg`. For example:

```console
% terraform import aws_iam_openid_connect_provider_client_id.example arn:aws:iam::11111111111:oidc-provider/app.eu.terraform.io|266362248691-342342xasdasdasda-apps.googleusercontent.com
```
