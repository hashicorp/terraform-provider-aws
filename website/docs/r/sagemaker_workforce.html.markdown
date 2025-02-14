---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_workforce"
description: |-
  Provides a SageMaker Workforce resource.
---

# Resource: aws_sagemaker_workforce

Provides a SageMaker Workforce resource.

## Example Usage

### Cognito Usage

```terraform
resource "aws_sagemaker_workforce" "example" {
  workforce_name = "example"

  cognito_config {
    client_id = aws_cognito_user_pool_client.example.id
    user_pool = aws_cognito_user_pool_domain.example.user_pool_id
  }
}

resource "aws_cognito_user_pool" "example" {
  name = "example"
}

resource "aws_cognito_user_pool_client" "example" {
  name            = "example"
  generate_secret = true
  user_pool_id    = aws_cognito_user_pool.example.id
}

resource "aws_cognito_user_pool_domain" "example" {
  domain       = "example"
  user_pool_id = aws_cognito_user_pool.example.id
}
```

### Oidc Usage

```terraform
resource "aws_sagemaker_workforce" "example" {
  workforce_name = "example"

  oidc_config {
    authorization_endpoint = "https://example.com"
    client_id              = "example"
    client_secret          = "example"
    issuer                 = "https://example.com"
    jwks_uri               = "https://example.com"
    logout_endpoint        = "https://example.com"
    token_endpoint         = "https://example.com"
    user_info_endpoint     = "https://example.com"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `workforce_name` - (Required) The name of the Workforce (must be unique).
* `cognito_config` - (Optional) Use this parameter to configure an Amazon Cognito private workforce. A single Cognito workforce is created using and corresponds to a single Amazon Cognito user pool. Conflicts with `oidc_config`. see [Cognito Config](#cognito-config) details below.
* `oidc_config` - (Optional) Use this parameter to configure a private workforce using your own OIDC Identity Provider. Conflicts with `cognito_config`. see [OIDC Config](#oidc-config) details below.
* `source_ip_config` - (Optional) A list of IP address ranges Used to create an allow list of IP addresses for a private workforce. By default, a workforce isn't restricted to specific IP addresses. see [Source Ip Config](#source-ip-config) details below.
* `workforce_vpc_config` - (Optional) configure a workforce using VPC. see [Workforce VPC Config](#workforce-vpc-config) details below.

### Cognito Config

* `client_id` - (Required) The client ID for your Amazon Cognito user pool.
* `user_pool` - (Required) ID for your Amazon Cognito user pool.

### Oidc Config

* `authentication_request_extra_params` - (Optional) A string to string map of identifiers specific to the custom identity provider (IdP) being used.
* `authorization_endpoint` - (Required) The OIDC IdP authorization endpoint used to configure your private workforce.
* `client_id` - (Required) The OIDC IdP client ID used to configure your private workforce.
* `client_secret` - (Required) The OIDC IdP client secret used to configure your private workforce.
* `issuer` - (Required) The OIDC IdP issuer used to configure your private workforce.
* `jwks_uri` - (Required) The OIDC IdP JSON Web Key Set (Jwks) URI used to configure your private workforce.
* `logout_endpoint` - (Required) The OIDC IdP logout endpoint used to configure your private workforce.
* `scope` - (Optional) An array of string identifiers used to refer to the specific pieces of user data or claims that the client application wants to access.
* `token_endpoint` - (Required) The OIDC IdP token endpoint used to configure your private workforce.
* `user_info_endpoint` - (Required) The OIDC IdP user information endpoint used to configure your private workforce.

### Source Ip Config

* `cidrs` - (Required) A list of up to 10 CIDR values.

### Workforce VPC Config

* `security_group_ids` - (Optional) The VPC security group IDs. The security groups must be for the same VPC as specified in the subnet.
* `subnets` - (Optional) The ID of the subnets in the VPC that you want to connect.
* `vpc_id` - (Optional) The ID of the VPC that the workforce uses for communication.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Workforce.
* `id` - The name of the Workforce.
* `subdomain` - The subdomain for your OIDC Identity Provider.
* `workforce_vpc_config.0.vpc_endpoint_id` - The IDs for the VPC service endpoints of your VPC workforce.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker Workforces using the `workforce_name`. For example:

```terraform
import {
  to = aws_sagemaker_workforce.example
  id = "example"
}
```

Using `terraform import`, import SageMaker Workforces using the `workforce_name`. For example:

```console
% terraform import aws_sagemaker_workforce.example example
```
