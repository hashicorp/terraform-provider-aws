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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `identity_pool_name` - (Required)  The Cognito Identity Pool name.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - An identity pool ID, e.g. `us-west-2:1a234567-8901-234b-5cde-f6789g01h2i3`.
* `arn` - ARN of the Pool.
* `allow_unauthenticated_identities` - Whether the identity pool supports unauthenticated logins or not.
* `allow_classic_flow` - Whether the classic / basic authentication flow is enabled.
* `developer_provider_name` - The "domain" by which Cognito will refer to your users.
* `cognito_identity_providers` - An array of Amazon Cognito Identity user pools and their client IDs.
* `openid_connect_provider_arns` - Set of OpendID Connect provider ARNs.
* `saml_provider_arns` - An array of Amazon Resource Names (ARNs) of the SAML provider for your identity.
* `supported_login_providers` - Key-Value pairs mapping provider names to provider app IDs.
* `tags` - A map of tags to assigned to the Identity Pool.
