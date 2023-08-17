---
subcategory: "Cognito Identity"
layout: "aws"
page_title: "AWS: aws_cognito_identity_pool"
description: |-
  Terraform data source for managing an AWS Cognito Identity Pool.
---

# Data Source: aws_cognito_identity_pool

Terraform data source for managing an AWS Cognito Identity Pool.

## Example Usage

### Basic Usage

```terraform
data "aws_cognito_identity_pool" "example" {
  identity_pool_name = "test pool"
}
```

## Argument Reference

The following arguments are required:

* `identity_pool_name` - (Required)  The Cognito Identity Pool name.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - An identity pool ID, e.g. `us-west-2:1a234567-8901-234b-5cde-f6789g01h2i3`.
* `arn` - ARN of the Pool. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `allow_unauthenticated_identities` - Whether the identity pool supports unauthenticated logins or not.
* `allow_classic_flow` - Enables or disables the classic / basic authentication flow. Default is `false`.
* `developer_provider_name` - The "domain" by which Cognito will refer to your users. This name acts as a placeholder that allows your
backend and the Cognito service to communicate about the developer provider.
* `cognito_identity_providers` - An array of [Amazon Cognito Identity user pools](#cognito-identity-providers) and their client IDs.
* `openid_connect_provider_arns` - Set of OpendID Connect provider ARNs.
* `saml_provider_arns` - An array of Amazon Resource Names (ARNs) of the SAML provider for your identity.
* `supported_login_providers` - Key-Value pairs mapping provider names to provider app IDs.
* `tags` - A map of tags to assign to the Identity Pool. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
