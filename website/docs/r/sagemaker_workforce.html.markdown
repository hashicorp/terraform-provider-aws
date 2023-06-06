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

The following arguments are supported:

* `workforce_name` - (Required) The name of the Workforce (must be unique).
* `cognito_config` - (Optional) Use this parameter to configure an Amazon Cognito private workforce. A single Cognito workforce is created using and corresponds to a single Amazon Cognito user pool. Conflicts with `oidc_config`. see [Cognito Config](#cognito-config) details below.
* `oidc_config` - (Optional) Use this parameter to configure a private workforce using your own OIDC Identity Provider. Conflicts with `cognito_config`. see [OIDC Config](#oidc-config) details below.
* `source_ip_config` - (Optional) A list of IP address ranges Used to create an allow list of IP addresses for a private workforce. By default, a workforce isn't restricted to specific IP addresses. see [Source Ip Config](#source-ip-config) details below.
* `workforce_vpc_config` - (Optional) configure a workforce using VPC. see [Workforce VPC Config](#workforce-vpc-config) details below.

### Cognito Config

* `client_id` - (Required) The client ID for your Amazon Cognito user pool.
* `user_pool` - (Required) ID for your Amazon Cognito user pool.

### Oidc Config

* `authorization_endpoint` - (Required) The OIDC IdP authorization endpoint used to configure your private workforce.
* `client_id` - (Required) The OIDC IdP client ID used to configure your private workforce.
* `client_secret` - (Required) The OIDC IdP client secret used to configure your private workforce.
* `issuer` - (Required) The OIDC IdP issuer used to configure your private workforce.
* `jwks_uri` - (Required) The OIDC IdP JSON Web Key Set (Jwks) URI used to configure your private workforce.
* `logout_endpoint` - (Required) The OIDC IdP logout endpoint used to configure your private workforce.
* `token_endpoint` - (Required) The OIDC IdP token endpoint used to configure your private workforce.
* `user_info_endpoint` - (Required) The OIDC IdP user information endpoint used to configure your private workforce.

### Source Ip Config

* `cidrs` - (Required) A list of up to 10 CIDR values.

### Workforce VPC Config

* `security_group_ids` - (Optional) The VPC security group IDs. The security groups must be for the same VPC as specified in the subnet.
* `subnets` - (Optional) The ID of the subnets in the VPC that you want to connect.
* `vpc_id` - (Optional) The ID of the VPC that the workforce uses for communication.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Workforce.
* `id` - The name of the Workforce.
* `subdomain` - The subdomain for your OIDC Identity Provider.
* `workforce_vpc_config.0.vpc_endpoint_id` - The IDs for the VPC service endpoints of your VPC workforce.

## Import

SageMaker Workforces can be imported using the `workforce_name`, e.g.,

```
$ terraform import aws_sagemaker_workforce.example example
```
