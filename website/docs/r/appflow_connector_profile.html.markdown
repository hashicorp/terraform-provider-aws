---
subcategory: "AppFlow"
layout: "aws"
page_title: "AWS: aws_appflow_connector_profile"
description: |-
  Provides an AppFlow Connector Profile resource.
---

# Resource: aws_appflow_connector_profile

Provides an AppFlow connector profile resource.

For information about AppFlow flows, see the [Amazon AppFlow API Reference][1].
For specific information about creating an AppFlow connector profile, see the
[CreateConnectorProfile][2] page in the Amazon AppFlow API Reference.

## Example Usage

```terraform
data "aws_iam_policy" "example" {
  name = "AmazonRedshiftAllCommandsFullAccess"
}

resource "aws_iam_role" "example" {
  name = "example_role"

  managed_policy_arns = [data.aws_iam_policy.test.arn]

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_s3_bucket" "example" {
  bucket = "example_bucket"
}

resource "aws_redshift_cluster" "example" {
  cluster_identifier = "example_cluster"
  database_name      = "example_db"
  master_username    = "exampleuser"
  master_password    = "examplePassword123!"
  node_type          = "dc1.large"
  cluster_type       = "single-node"
}

resource "aws_appflow_connector_profile" "example" {
  name            = "example_profile"
  connector_type  = "Redshift"
  connection_mode = "Public"

  connector_profile_config {

    connector_profile_credentials {
      redshift {
        password = aws_redshift_cluster.example.master_password
        username = aws_redshift_cluster.example.master_username
      }
    }

    connector_profile_properties {
      redshift {
        bucket_name  = aws_s3_bucket.example.name
        database_url = "jdbc:redshift://${aws_redshift_cluster.example.endpoint}/${aws_redshift_cluster.example.database_name}"
        role_arn     = aws_iam_role.example.arn
      }
    }
  }
}
```

## Argument Reference

The AppFlow connector profile argument layout is a complex structure. The following top-level arguments are supports:

* `name ` (Required) - Name of the connector profile. The name is unique for each `ConnectorProfile` in your AWS account.
* `connection_mode` (Required) - Indicates the connection mode and specifies whether it is public or private. Private flows use AWS PrivateLink to route data over AWS infrastructure without exposing it to the public internet. One of: `Public`, `Private`.
* `connector_label` (Optional) - The label of the connector. The label is unique for each ConnectorRegistration in your AWS account. Only needed if calling for `CustomConnector` connector type.
* `connector_profile_config` (Required) - Defines the connector-specific configuration and credentials. See [Connector Profile Config](#connector-profile-config) for more details.
* `connector_type` (Required) - The type of connector. One of: `Amplitude`, `CustomConnector`, `CustomerProfiles`, `Datadog`, `Dynatrace`, `EventBridge`, `Googleanalytics`, `Honeycode`, `Infornexus`, `LookoutMetrics`, `Marketo`, `Redshift`, `S3`, `Salesforce`, `SAPOData`, `Servicenow`, `Singular`, `Slack`, `Snowflake`, `Trendmicro`, `Upsolver`, `Veeva`, `Zendesk`.
* `kms_arn` (Optional) - ARN (Amazon Resource Name) of the Key Management Service (KMS) key you provide for encryption. This is required if you do not want to use the Amazon AppFlow-managed KMS key. If you don't provide anything here, Amazon AppFlow uses the Amazon AppFlow-managed KMS key.

### Connector Profile Config

* `connector_profile_credentials` (Required) - The connector-specific credentials required by each connector. See [Connector Profile Credentials](#connector-profile-credentials) for more details.
* `connector_profile_properties` (Required) - The connector-specific properties of the profile configuration. See [Connector Profile Properties](#connector-profile-properties) for more details.

### Connector Profile Credentials

* `amplitude` (Optional) - The connector-specific credentials required when using Amplitude. See [Amplitude Connector Profile Credentials](#amplitude-connector-profile-credentials) for more details.
* `custom_connector` (Optional) - The connector-specific profile credentials required when using the custom connector. See [Custom Connector Profile Credentials](#custom-connector-profile-credentials) for more details.
* `datadog` (Optional) - Connector-specific credentials required when using Datadog. See [Datadog Connector Profile Credentials](#datadog-connector-profile-credentials) for more details.
* `dynatrace` (Optional) - The connector-specific credentials required when using Dynatrace. See [Dynatrace Connector Profile Credentials](#dynatrace-connector-profile-credentials) for more details.
* `google_analytics` (Optional) - The connector-specific credentials required when using Google Analytics. See [Google Analytics Connector Profile Credentials](#google-analytics-connector-profile-credentials) for more details.
* `honeycode` (Optional) - The connector-specific credentials required when using Amazon Honeycode. See [Honeycode Connector Profile Credentials](#honeycode-connector-profile-credentials) for more details.
* `infor_nexus` (Optional) - The connector-specific credentials required when using Infor Nexus. See [Infor Nexus Connector Profile Credentials](#infor-nexus-connector-profile-credentials) for more details.
* `marketo` (Optional) - Connector-specific credentials required when using Marketo. See [Marketo Connector Profile Credentials](#marketo-connector-profile-credentials) for more details.
* `redshift` (Optional) - Connector-specific credentials required when using Amazon Redshift. See [Redshift Connector Profile Credentials](#redshift-connector-profile-credentials) for more details.
* `salesforce` (Optional) - The connector-specific credentials required when using Salesforce. See [Salesforce Connector Profile Credentials](#salesforce-connector-profile-credentials) for more details.
* `sapo_data` (Optional) - The connector-specific credentials required when using SAPOData. See [SAPOData Connector Profile Credentials](#sapodata-connector-profile-credentials) for more details.
* `service_now` (Optional) - The connector-specific credentials required when using ServiceNow. See [ServiceNow Connector Profile Credentials](#servicenow-connector-profile-credentials) for more details.
* `singular` (Optional) - Connector-specific credentials required when using Singular. See [Singular Connector Profile Credentials](#singular-connector-profile-credentials) for more details.
* `slack` (Optional) - Connector-specific credentials required when using Slack. See [Slack Connector Profile Credentials](#amplitude-connector-profile-credentials) for more details.
* `snowflake` (Optional) - The connector-specific credentials required when using Snowflake. See [Snowflake Connector Profile Credentials](#snowflake-connector-profile-credentials) for more details.
* `trendmicro` (Optional) - The connector-specific credentials required when using Trend Micro. See [Trend Micro Connector Profile Credentials](#trendmicro-connector-profile-credentials) for more details.
* `veeva` (Optional) - Connector-specific credentials required when using Veeva. See [Veeva Connector Profile Credentials](#veeva-connector-profile-credentials) for more details.
* `zendesk` (Optional) - Connector-specific credentials required when using Zendesk. See [Zendesk Connector Profile Credentials](#zendesk-connector-profile-credentials) for more details.

#### Amplitude Connector Profile Credentials

* `api_key` (Required) - Unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.
* `secret_key` (Required) - The Secret Access Key portion of the credentials.

#### Custom Connector Profile Credentials

* `api_key` (Optional) - API keys required for the authentication of the user.
    * `api_key` (Required) - The API key required for API key authentication.
    * `api_secret_key` (Optional) - The API secret key required for API key authentication.
* `authentication_type` (Required) - The authentication type that the custom connector uses for authenticating while creating a connector profile. One of: `APIKEY`, `BASIC`, `CUSTOM`, `OAUTH2`.
* `basic` (Optional) - Basic credentials that are required for the authentication of the user.
    * `password` (Required) - The password to use to connect to a resource.
    * `username` (Required) - The username to use to connect to a resource.
* `custom` (Optional) - If the connector uses the custom authentication mechanism, this holds the required credentials.
    * `credentials_map` (Optional) - A map that holds custom authentication credentials.
    * `custom_authentication_type` (Required) - The custom authentication type that the connector uses.
* `oauth2` (Optional) - OAuth 2.0 credentials required for the authentication of the user.
    * `access_token` (Optional) - The access token used to access the connector on your behalf.
    * `client_id` (Optional) - The identifier for the desired client.
    * `client_secret` (Optional) - The client secret used by the OAuth client to authenticate to the authorization server.
    * `oauth_request` (Optional) - Used by select connectors for which the OAuth workflow is supported. See [OAuth Request](#oauth-request) for more details.
    * `refresh_token` (Optional) - The refresh token used to refresh an expired access token.

#### Datadog Connector Profile Credentials

* `api_key` (Required) - Unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.
* `application_key` (Required) - Application keys, in conjunction with your API key, give you full access to Datadog’s programmatic API. Application keys are associated with the user account that created them. The application key is used to log all requests made to the API.

#### Dynatrace Connector Profile Credentials

* `api_token` (Required) - The API tokens used by Dynatrace API to authenticate various API calls.

#### Google Analytics Connector Profile Credentials

* `access_token` (Optional) - The credentials used to access protected Google Analytics resources.
* `client_id` (Required) - The identifier for the desired client.
* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint. See [OAuth Request](#oauth-request) for more details.
* `refresh_token` (Optional) - The credentials used to acquire new access tokens. This is required only for OAuth2 access tokens, and is not required for OAuth1 access tokens.

#### Honeycode Connector Profile Credentials

* `access_token` (Optional) - The credentials used to access protected Amazon Honeycode resources.
* `oauth_request` (Optional) - Used by select connectors for which the OAuth workflow is supported, such as Salesforce, Google Analytics, Marketo, Zendesk, and Slack. See [OAuth Request](#oauth-request) for more details.
* `refresh_token` (Optional) - The credentials used to acquire new access tokens.

#### Infor Nexus Connector Profile Credentials

* `access_key_id` (Required) - The Access Key portion of the credentials.
* `datakey` (Required) - Encryption keys used to encrypt data.
* `secret_access_key` (Required) - The secret key used to sign requests.
* `user_id` (Required) - Identifier for the user.

#### Marketo Connector Profile Credentials

* `access_token` (Optional) - The credentials used to access protected Marketo resources.
* `client_id` (Required) - The identifier for the desired client.
* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint. See [OAuth Request](#oauth-request) for more details.

#### Redshift Connector Profile Credentials

* `password` (Required) - Password that corresponds to the user name.
* `username` (Required) - Name of the user.

#### Salesforce Connector Profile Credentials

* `access_token` (Optional) - The credentials used to access protected Salesforce resources.
* `client_credentials_arn` (Optional) - The secret manager ARN, which contains the client ID and client secret of the connected app.
* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint. See [OAuth Request](#oauth-request) for more details.
* `refresh_token` (Optional) - The credentials used to acquire new access tokens.

#### SAPOData Connector Profile Credentials

* `basic_auth_credentials` (Optional) - The SAPOData basic authentication credentials.
    * `password` (Required) - The password to use to connect to a resource.
    * `username` (Required) - The username to use to connect to a resource.
* `oauth_credentials` (Optional) - The SAPOData OAuth type authentication credentials.
    * `access_token` (Optional) - The access token used to access protected SAPOData resources.
    * `client_id` (Required) - The identifier for the desired client.
    * `client_secret` (Required) -  The client secret used by the OAuth client to authenticate to the authorization server.
    * `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint. See [OAuth Request](#oauth-request) for more details.
    * `refresh_token` (Optional) - The refresh token used to refresh expired access token.

#### ServiceNow Connector Profile Credentials

* `password` (Required) - Password that corresponds to the user name.
* `username` (Required) - Name of the user.

#### Singular Connector Profile Credentials

* `api_key` (Required) - Unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.

#### Slack Connector Profile Credentials

* `access_token` (Optional) - The credentials used to access protected Slack resources.
* `client_id` (Required) - The identifier for the client.
* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint. See [OAuth Request](#oauth-request) for more details.

#### Snowflake Connector Profile Credentials

* `password` (Required) - Password that corresponds to the user name.
* `username` (Required) - Name of the user.

#### Trendmicro Connector Profile Credentials

* `api_secret_key` (Required) - The Secret Access Key portion of the credentials.

#### Veeva Connector Profile Credentials

* `password` (Required) - Password that corresponds to the user name.
* `username` (Required) - Name of the user.

#### Zendesk Connector Profile Credentials

* `access_token` (Optional) - The credentials used to access protected Zendesk resources.
* `client_id` (Required) - The identifier for the desired client.
* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint. See [OAuth Request](#oauth-request) for more details.

##### OAuth Request

* `auth_code` (Optional) - The code provided by the connector when it has been authenticated via the connected app.
* `redirect_uri` (Optional) - The URL to which the authentication server redirects the browser after authorization has been granted.

### Connector Profile Properties

* `custom_connector` (Optional) - The connector-specific profile properties required when using the custom connector. See [Custom Connector Profile Properties](#custom-connector-profile-properties) for more details.
* `datadog` (Optional) - Connector-specific properties required when using Datadog. See [Generic Connector Profile Properties](#generic-connector-profile-properties) for more details.
* `dynatrace` (Optional) - The connector-specific properties required when using Dynatrace. See [Generic Connector Profile Properties](#generic-connector-profile-properties) for more details.
* `infor_nexus` (Optional) - The connector-specific properties required when using Infor Nexus. See [Generic Connector Profile Properties](#generic-connector-profile-properties) for more details.
* `marketo` (Optional) - Connector-specific properties required when using Marketo. See [Generic Connector Profile Properties](#generic-connector-profile-properties) for more details.
* `redshift` (Optional) - Connector-specific properties required when using Amazon Redshift. See [Redshift Connector Profile Properties](#redshift-connector-profile-properties) for more details.
* `salesforce` (Optional) - The connector-specific properties required when using Salesforce. See [Salesforce Connector Profile Properties](#salesforce-connector-profile-properties) for more details.
* `sapo_data` (Optional) - The connector-specific properties required when using SAPOData. See [SAPOData Connector Profile Properties](#sapodata-connector-profile-properties) for more details.
* `service_now` (Optional) - The connector-specific properties required when using ServiceNow. See [Generic Connector Profile Properties](#generic-connector-profile-properties) for more details.
* `slack` (Optional) - Connector-specific properties required when using Slack. See [Generic Connector Profile Properties](#generic-connector-profile-properties) for more details.
* `snowflake` (Optional) - The connector-specific properties required when using Snowflake. See [Snowflake Connector Profile Properties](#snowflake-connector-profile-properties) for more details.
* `veeva` (Optional) - Connector-specific properties required when using Veeva. See [Generic Connector Profile Properties](#generic-connector-profile-properties) for more details.
* `zendesk` (Optional) - Connector-specific properties required when using Zendesk. See [Generic Connector Profile Properties](#generic-connector-profile-properties) for more details.

#### Custom Connector Profile Properties

* `oauth2_properties` (Optional) - The OAuth 2.0 properties required for OAuth 2.0 authentication.
    * `oauth2_grant_type` (Required) - The OAuth 2.0 grant type used by connector for OAuth 2.0 authentication. One of: `AUTHORIZATION_CODE`, `CLIENT_CREDENTIALS`.
    * `token_url` (Required) - The token URL required for OAuth 2.0 authentication.
    * `token_url_custom_properties` (Optional) - Associates your token URL with a map of properties that you define. Use this parameter to provide any additional details that the connector requires to authenticate your request.
* `profile_properties` (Optional) - A map of properties that are required to create a profile for the custom connector.

#### Generic Connector Profile Properties

Datadog, Dynatrace, Infor Nexus, Marketo, ServiceNow, Slack, Veeva, and Zendesk all support the following attributes:

* `instance_url` (Required) - The location of the Datadog resource.

#### Redshift Connector Profile Properties

* `bucket_name` (Required) - A name for the associated Amazon S3 bucket.
* `bucket_prefix` (Optional) - The object key for the destination bucket in which Amazon AppFlow places the files.
* `cluster_identifier` (Optional) - The unique ID that's assigned to an Amazon Redshift cluster.
* `database_name` (Optional) - The name of an Amazon Redshift database.
* `database_url` (Required) - The JDBC URL of the Amazon Redshift cluster.
* `data_api_role_arn` (Optional) - ARN of the IAM role that permits AppFlow to access the database through Data API.
* `role_arn` (Required) - ARN of the IAM role.

#### Salesforce Connector Profile Properties

* `instance_url` (Optional) - The location of the Salesforce resource.
* `is_sandbox_environment` (Optional) - Indicates whether the connector profile applies to a sandbox or production environment.

#### SAPOData Connector Profile Properties

* `application_host_url` (Required) - The location of the SAPOData resource.
* `application_service_path` (Required) - The application path to catalog service.
* `client_number` (Required) - The client number for the client creating the connection.
* `logon_language` (Optional) - The logon language of SAPOData instance.
* `oauth_properties` (Optional) - The SAPOData OAuth properties required for OAuth type authentication.
    * `auth_code_url` (Required) - The authorization code url required to redirect to SAP Login Page to fetch authorization code for OAuth type authentication.
    * `oauth_scopes` (Required) - The OAuth scopes required for OAuth type authentication.
    * `token_url` (Required) - The token url required to fetch access/refresh tokens using authorization code and also to refresh expired access token using refresh token.
* `port_number` (Required) - The port number of the SAPOData instance.
* `private_link_service_name` (Optional) - The SAPOData Private Link service name to be used for private data transfers.

#### Snowflake Connector Profile Properties

* `account_name` (Optional) - The name of the account.
* `bucket_name` (Required) - The name of the Amazon S3 bucket associated with Snowflake.
* `bucket_prefix` (Optional) - The bucket path that refers to the Amazon S3 bucket associated with Snowflake.
* `private_link_service_name` (Optional) - The Snowflake Private Link service name to be used for private data transfers.
* `region` (Optional) - AWS Region of the Snowflake account.
* `stage` (Required) - Name of the Amazon S3 stage that was created while setting up an Amazon S3 stage in the Snowflake account. This is written in the following format: `<Database>.<Schema>.<Stage Name>`.
* `warehouse` (Required) - The name of the Snowflake warehouse.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connector profile.
* `credentials_arn` - ARN of the connector profile credentials.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppFlow Connector Profile using the connector profile `arn`. For example:

```terraform
import {
  to = aws_appflow_connector_profile.profile
  id = "arn:aws:appflow:us-west-2:123456789012:connectorprofile/example-profile"
}
```

Using `terraform import`, import AppFlow Connector Profile using the connector profile `arn`. For example:

```console
% terraform import aws_appflow_connector_profile.profile arn:aws:appflow:us-west-2:123456789012:connectorprofile/example-profile
```

[1]: https://docs.aws.amazon.com/appflow/1.0/APIReference/Welcome.html
[2]: https://docs.aws.amazon.com/appflow/1.0/APIReference/API_CreateConnectorProfile.html
