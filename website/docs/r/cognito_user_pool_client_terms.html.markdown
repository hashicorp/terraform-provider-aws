---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_client_terms"
description: |-
  Manages a terms document for an Amazon Cognito user pool app client.
---

# Resource: aws_cognito_user_pool_client_terms

Manages a terms of use or privacy policy document for an Amazon Cognito user pool app client.

Cognito displays the links on the managed login sign-up page when both document types have links configured for the same app client. Use a separate resource instance for each document type. For details on managed login and localization, see [Terms documents](https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-pools-managed-login.html#managed-login-terms-documents).

## Example Usage

```terraform
resource "aws_cognito_user_pool_client_terms" "example" {
  client_id    = "1example23456789"
  terms_name   = "terms-of-use"
  user_pool_id = "us-west-2_EXAMPLE"
  links = {
    "cognito:default" = "https://example.com/terms"
    "cognito:french"  = "https://example.com/fr/terms"
  }
}
```

## Argument Reference

The following arguments are required:

* `client_id` - (Required) ID of the app client in the specified user pool. Changing this value replaces the resource.
* `terms_name` - (Required) Document type. Allowed values are: `terms-of-use` and `privacy-policy`.
* `user_pool_id` - (Required) ID of the Cognito user pool. Changing this value replaces the resource.

The following arguments are optional:

* `links` - (Optional) Map of document URLs keyed by [Cognito language identifier](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateTerms.html#CognitoUserPools-CreateTerms-request-Links). Use `cognito:default` to provide a URL for languages without their own entry. Default value is an empty map.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `terms_id` - ID of the document assigned by Cognito.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute, for example:

```terraform
import {
  to = aws_cognito_user_pool_client_terms.example
  identity = {
    user_pool_id = "us-west-2_EXAMPLE"
    terms_id     = "a1b2c3d4-5678-4321-8765-abcdef123456"
  }
}
```

### Identity Schema

#### Required

* `terms_id` (String) ID of the terms document.
* `user_pool_id` (String) ID of the Cognito user pool.

#### Optional

* `account_id` (String) AWS account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `id` attribute. Use the user pool ID and terms document ID separated by a comma.

```terraform
import {
  to = aws_cognito_user_pool_client_terms.example
  id = "us-west-2_EXAMPLE,a1b2c3d4-5678-4321-8765-abcdef123456"
}
```

Using `terraform import`, import the resource using the user pool ID and terms document ID, separated by a comma. For example:

```console
% terraform import aws_cognito_user_pool_client_terms.example us-west-2_EXAMPLE,a1b2c3d4-5678-4321-8765-abcdef123456
```
