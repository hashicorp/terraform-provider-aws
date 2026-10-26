---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_connection"
description: |-
  Provides details about an AWS Glue Connection.
---

# Data Source: aws_glue_connection

Provides details about an AWS Glue Connection.

## Example Usage

```terraform
data "aws_glue_connection" "example" {
  id = "123456789123:connection"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Concatenation of the catalog ID and connection name. For example, if your account ID is `123456789123` and the connection name is `conn` then the ID is `123456789123:conn`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Glue Connection.
* `athena_properties` - Map of connection properties specific to the Athena compute environment.
* `authentication_configuration` - Configuration block for authentication options. See [`authentication_configuration` Block](#authentication_configuration-block) for details.
* `catalog_id` - Catalog ID of the Glue Connection.
* `connection_properties` - Map of connection properties.
* `connection_type` - Type of Glue Connection.
* `description` - Description of the connection.
* `match_criteria` - List of criteria that can be used in selecting this connection.
* `name` - Name of the Glue Connection.
* `physical_connection_requirements` - Map of physical connection requirements, such as VPC and SecurityGroup. See [`physical_connection_requirements` Block](#physical_connection_requirements-block) for details.
* `tags` - Tags assigned to the resource.

### `authentication_configuration` Block

* `authentication_type` - Type of authentication used for the connection.
* `basic_authentication_credentials` - Basic authentication credentials. See [`basic_authentication_credentials` Block](#basic_authentication_credentials-block) for details.
* `custom_authentication_credentials` - Map of credentials used when the authentication type is custom authentication.
* `kms_key_arn` - ARN of the KMS key used to encrypt the connection.
* `oauth2_properties` - OAuth2 properties. See [`oauth2_properties` Block](#oauth2_properties-block) for details.
* `secret_arn` - ARN of the secret used for authentication.

### `basic_authentication_credentials` Block

* `password` - Password used for basic authentication.
* `username` - Username used for basic authentication.

### `oauth2_properties` Block

* `authorization_code_properties` - Authorization code properties. See [`authorization_code_properties` Block](#authorization_code_properties-block) for details.
* `oauth2_client_application` - OAuth2 client application. See [`oauth2_client_application` Block](#oauth2_client_application-block) for details.
* `oauth2_credentials` - OAuth2 credentials. See [`oauth2_credentials` Block](#oauth2_credentials-block) for details.
* `oauth2_grant_type` - OAuth2 grant type.
* `token_url` - URL of the provider's authentication server used to exchange an authorization code for an access token.
* `token_url_parameters_map` - Map of parameters to add to the token request.

### `authorization_code_properties` Block

* `authorization_code` - Authorization code used to obtain an access token.
* `redirect_uri` - Redirect URI used in the authorization code request.

### `oauth2_client_application` Block

* `aws_managed_client_application_reference` - Reference to the AWS managed client application.
* `user_managed_client_application_client_id` - Client ID of the user-managed client application.

### `oauth2_credentials` Block

* `access_token` - Access token used for OAuth2 authentication.
* `jwt_token` - JWT token used for OAuth2 authentication.
* `refresh_token` - Refresh token used for OAuth2 authentication.
* `user_managed_client_application_client_secret` - Client secret of the user-managed client application.

### `physical_connection_requirements` Block

* `availability_zone` - Availability Zone used by the connection.
* `security_group_id_list` - List of security group IDs used by the connection.
* `subnet_id` - Subnet ID used by the connection.
