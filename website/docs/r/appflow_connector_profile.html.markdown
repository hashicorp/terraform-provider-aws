---
subcategory: "AppFlow"
layout: "aws"
page_title: "AWS: aws_appflow_connector_profile"
description: |-
  Provides an AppFlow connector profile resource.
---

# Resource: aws_appflow_connector_profile

Creates an AWS AppFlow connector profile.

For information about AppFlow flows, see the
[Amazon AppFlow API Reference][1]. For specific information about creating
AppFlow connector profile, see the [CreateConnectorProfile][2] page in the Amazon
AppFlow API Reference.

## Example Usage

### Amplitude

```terraform
resource "aws_appflow_connector_profile" "amplitude" {
  name            = "amplitude-connector"
  connection_mode = "Public"
  connector_type  = "Amplitude"

  connector_profile_config {
    connector_profile_credentials {
      amplitude {
        api_key    = "0123456789abcdef0123456789abcdef"
        secret_key = "0123456789abcdef0123456789abcdef"
      }
    }
    connector_profile_properties {
      amplitude {
      }
    }
  }
}
```

### Zendesk

```terraform
data "aws_region" "current" {}

resource "aws_appflow_connector_profile" "zendesk" {
  name            = "zendesk-connector"
  connection_mode = "Public"
  connector_type  = "Zendesk"

  connector_profile_config {
    connector_profile_credentials {
      zendesk {
        client_id     = "zendesk-client-id"
        client_secret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
        access_token  = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
        oauth_request {
          redirect_uri = "https://${data.aws_region.current.name}.console.aws.amazon.com/appflow/oauth"
        }
      }
    }
    connector_profile_properties {
      zendesk {
        instance_url = "https://subdomain.zendesk.com"
      }
    }
  }
}
```


## Argument Reference

The AppFlow connector profile argument layout is a complex structure.

### Top-Level Arguments

* `name ` (Required) - The name of the connector profile. The name is unique for each ConnectorProfile in your AWS account.

* `connection_mode` (Required) - Indicates the connection mode and specifies whether it is public or private. Private flows use AWS PrivateLink to route data over AWS infrastructure without exposing it to the public internet. One of: `Public`, `Private`.

* `connector_type` (Required) - The type of connector. One of: `Amplitude`, `CustomConnector`, `CustomerProfiles`, `Datadog`, `Dynatrace`, `EventBridge`, `Googleanalytics`, `Honeycode`, `Infornexus`, `LookoutMetrics`, `Marketo`, `Redshift`, `S3`, `Salesforce`, `SAPOData`, `Servicenow`, `Singular`, `Slack`, `Snowflake`, `Trendmicro`, `Upsolver`, `Veeva`, `Zendesk`.

* `connector_profile_config` (Required) - Defines the connector-specific [configuration and credentials](#connector-profile-config-arguments).

* `connector_label` (Optional) - The label of the connector. The label is unique for each ConnectorRegistration in your AWS account. Only needed if calling for `CustomConnector` connector type.

* `kms_arn` (Optional) - The ARN (Amazon Resource Name) of the Key Management Service (KMS) key you provide for encryption. This is required if you do not want to use the Amazon AppFlow-managed KMS key. If you don't provide anything here, Amazon AppFlow uses the Amazon AppFlow-managed KMS key.

#### Connector Profile Config Arguments

* `connector_profile_credentials` (Required) - The connector-specific [credentials](#connector-profile-credentials-arguments) required by each connector.

* `connector_profile_properties` (Required) - The connector-specific [properties](#connector-profile-properties-arguments) of the profile configuration.


##### Connector Profile Credentials Arguments

* `amplitude` (Optional) - The connector-specific [credentials](#amplitude-connector-profile-credentials-arguments) required when using Amplitude.

* `custom_connector` (Optional) - The connector-specific profile [credentials](#custom-connector-profile-credentials-arguments) required when using the custom connector.

* `datadog` (Optional) - The connector-specific [credentials](#datadog-connector-profile-credentials-arguments) required when using Datadog.

* `dynatrace` (Optional) - The connector-specific [credentials](#dynatrace-connector-profile-credentials-arguments) required when using Dynatrace.

* `google_analytics` (Optional) - The connector-specific [credentials](#google-analytics-connector-profile-credentials-arguments) required when using Google Analytics.

* `honeycode` (Optional) - The connector-specific [credentials](#honeycode-connector-profile-credentials-arguments) required when using Amazon Honeycode.

* `infor_nexus` (Optional) - The connector-specific [credentials](#infor-nexus-connector-profile-credentials-arguments) required when using Infor Nexus.

* `marketo` (Optional) - The connector-specific [credentials](#marketo-connector-profile-credentials-arguments) required when using Marketo.

* `redshift` (Optional) - The connector-specific [credentials](#redshift-connector-profile-credentials-arguments) required when using Amazon Redshift.

* `salesforce` (Optional) - The connector-specific [credentials](#salesforce-connector-profile-credentials-arguments) required when using Salesforce.

* `service_now` (Optional) - The connector-specific [credentials](#servicenow-connector-profile-credentials-arguments) required when using ServiceNow.

* `singular` (Optional) - The connector-specific [credentials](#singular-connector-profile-credentials-arguments) required when using Singular.

* `slack` (Optional) - The connector-specific [credentials](#slack-connector-profile-credentials-arguments) required when using Slack.

* `snowflake` (Optional) - The connector-specific [credentials](#snowflake-connector-profile-credentials-arguments) required when using Snowflake.

* `trendmicro` (Optional) - The connector-specific [credentials](#trendmicro-connector-profile-credentials-arguments) required when using Trend Micro.

* `veeva` (Optional) - The connector-specific [credentials](#veeva-connector-profile-credentials-arguments) required when using Veeva.

* `zendesk` (Optional) - The connector-specific [credentials](#zendesk-connector-profile-credentials-arguments) required when using Zendesk.

##### Amplitude Connector Profile Credentials Arguments

* `api_key` (Required) - A unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.

* `secret_key` (Required) - The Secret Access Key portion of the credentials.

##### Custom Connector Profile Credentials Arguments

* `api_key` (Optional) - The API keys required for the authentication of the user.
    * `api_key` (Required) - The API key required for API key authentication.
    * `api_secret_key` (Optional) - The API secret key required for API key authentication.

* `authentication_type` (Required) - The authentication type that the custom connector uses for authenticating while creating a connector profile. One of: `APIKEY`, `BASIC`, `CUSTOM`, `OAUTH2`.

* `basic` (Optional) - The basic credentials that are required for the authentication of the user.
    * `password` (Required) - The password to use to connect to a resource.
    * `username` (Required) - The username to use to connect to a resource.

* `custom` (Optional) - If the connector uses the custom authentication mechanism, this holds the required credentials.
    * `credentials_map` (Optional) - A map that holds custom authentication credentials.
    * `custom_authentication_type` (Required) - The custom authentication type that the connector uses.

* `oauth2` (Optional) - The OAuth 2.0 credentials required for the authentication of the user.
    * `access_token` (Optional) - The access token used to access the connector on your behalf.
    * `client_id` (Optional) - The identifier for the desired client.
    * `client_secret` (Optional) - The client secret used by the OAuth client to authenticate to the authorization server.
    * `oauth_request` (Optional) - Used by select connectors for which the OAuth workflow is supported.
        * `auth_code` (Optional) - The code provided by the connector when it has been authenticated via the connected app.
        * `redirect_uri` (Optional) - The URL to which the authentication server redirects the browser after authorization has been granted.
    * `refresh_token` (Optional) - The refresh token used to refresh an expired access token.

##### Datadog Connector Profile Credentials Arguments

* `api_key` (Required) - A unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.

* `application_key` (Required) - Application keys, in conjunction with your API key, give you full access to Datadog’s programmatic API. Application keys are associated with the user account that created them. The application key is used to log all requests made to the API.

##### Dynatrace Connector Profile Credentials Arguments

* `api_token` (Required) - The API tokens used by Dynatrace API to authenticate various API calls.

##### Google Analytics Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Google Analytics resources.

* `client_id` (Required) - The identifier for the desired client.

* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.

* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint.

* `refresh_token` (Optional) - The credentials used to acquire new access tokens. This is required only for OAuth2 access tokens, and is not required for OAuth1 access tokens.

##### Honeycode Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Amazon Honeycode resources.

* `oauth_request` (Optional) - Used by select connectors for which the OAuth workflow is supported, such as Salesforce, Google Analytics, Marketo, Zendesk, and Slack.

* `refresh_token` (Optional) - The credentials used to acquire new access tokens.

##### Infor Nexus Connector Profile Credentials Arguments

* `access_key_id` (Required) - The Access Key portion of the credentials.

* `datakey` (Required) - The encryption keys used to encrypt data.

* `secret_access_key` (Required) - The secret key used to sign requests.

* `user_id` (Required) - The identifier for the user.

##### Marketo Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Marketo resources.

* `client_id` (Required) - The identifier for the desired client.

* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.

* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint.

##### Redshift Connector Profile Credentials Arguments

* `password` (Required) - The password that corresponds to the user name.

* `username` (Required) - The name of the user.

##### Salesforce Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Salesforce resources.

* `client_credentials_arn` (Optional) - The secret manager ARN, which contains the client ID and client secret of the connected app.

* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint.

* `refresh_token` (Optional) - The credentials used to acquire new access tokens.

##### ServiceNow Connector Profile Credentials Arguments

* `password` (Required) - The password that corresponds to the user name.

* `username` (Required) - The name of the user.

##### Singular Connector Profile Credentials Arguments

* `api_key` (Required) - A unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.

##### Slack Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Slack resources.

* `client_id` (Required) - The identifier for the client.

* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.

* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint.

##### Snowflake Connector Profile Credentials Arguments

* `password` (Required) - The password that corresponds to the user name.

* `username` (Required) - The name of the user.

##### Trendmicro Connector Profile Credentials Arguments

* `api_secret_key` (Required) - The Secret Access Key portion of the credentials.

##### Veeva Connector Profile Credentials Arguments

* `password` (Required) - The password that corresponds to the user name.

* `username` (Required) - The name of the user.

##### Zendesk Connector Profile Credentials Arguments

* `access_token` (Optional) - The credentials used to access protected Zendesk resources.

* `client_id` (Required) - The identifier for the desired client.

* `client_secret` (Required) - The client secret used by the OAuth client to authenticate to the authorization server.

* `oauth_request` (Optional) - The OAuth requirement needed to request security tokens from the connector endpoint.

##### Connector Profile Properties Arguments

* `custom_connector` (Optional) - The [properties](#custom-connector-profile-properties-arguments) required by the custom connector.

* `datadog` (Optional) - The connector-specific [properties](#datadog-connector-profile-properties-arguments) required by Datadog.

* `dynatrace` (Optional) - The connector-specific [properties](#dynatrace-connector-profile-properties-arguments) required by Dynatrace.

* `infor_nexus` (Optional) - The connector-specific [properties](#infor-nexus-connector-profile-properties-arguments) required by Infor Nexus.

* `marketo` (Optional) - The connector-specific [properties](#marketo-connector-profile-properties-arguments) required by Marketo.

* `redshift` (Optional) - The connector-specific [properties](#redshift-connector-profile-properties-arguments) required by Amazon Redshift.

* `salesforce` (Optional) - The connector-specific [properties](#salesforce-connector-profile-properties-arguments) required by Salesforce.

* `service_now` (Optional) - The connector-specific [properties](#servicenow-connector-profile-properties-arguments) required by ServiceNow.

* `slack` (Optional) - The connector-specific [properties](#slack-connector-profile-properties-arguments) required by Slack.

* `snowflake` (Optional) - The connector-specific [properties](#snowflake-connector-profile-properties-arguments) required by Snowflake.

* `veeva` (Optional) - The connector-specific [properties](#veeva-connector-profile-properties-arguments) required by Veeva.

* `zendesk` (Optional) - The connector-specific [properties](#zendesk-connector-profile-properties-arguments) required by Zendesk.

##### Custom Connector Profile Properties Arguments

* `oauth2_properties` (Optional) - The OAuth 2.0 properties required for OAuth 2.0 authentication.
    * `oauth2_grant_type` (Required) - The OAuth 2.0 grant type used by connector for OAuth 2.0 authentication. One of: `AUTHORIZATION_CODE`, `CLIENT_CREDENTIALS`.
    * `token_url` (Required) - The token URL required for OAuth 2.0 authentication.
    * `token_url_custom_properties` (Optional) - Associates your token URL with a map of properties that you define. Use this parameter to provide any additional details that the connector requires to authenticate your request.

* `profile_properties` (Optional) - A map of properties that are required to create a profile for the custom connector.

##### Datadog Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Datadog resource.

##### Dynatrace Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Dynatrace resource.

##### Infor Nexus Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Infor Nexus resource.

##### Marketo Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Marketo resource.

##### Redshift Connector Profile Properties Arguments

* `bucket_name` (Required) - A name for the associated Amazon S3 bucket.

* `bucket_prefix` (Optional) - The object key for the destination bucket in which Amazon AppFlow places the files.

* `database_url` (Required) - The JDBC URL of the Amazon Redshift cluster.

* `role_arn` (Required) - The Amazon Resource Name (ARN) of the IAM role.

##### Salesforce Connector Profile Properties Arguments

* `instance_url` (Optional) - The location of the Salesforce resource.

* `is_sandbox_environment` (Optional) - Indicates whether the connector profile applies to a sandbox or production environment.

##### ServiceNow Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the ServiceNow resource.

##### Slack Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Slack resource.

##### Snowflake Connector Profile Properties Arguments

* `account_name` (Optional) - The name of the account.

* `bucket_name` (Required) - The name of the Amazon S3 bucket associated with Snowflake.

* `bucket_prefix` (Optional) - The bucket path that refers to the Amazon S3 bucket associated with Snowflake.

* `private_link_service_name` (Optional) - The Snowflake Private Link service name to be used for private data transfers.

* `region` (Optional) - The AWS Region of the Snowflake account.

* `stage` (Required) - The name of the Amazon S3 stage that was created while setting up an Amazon S3 stage in the Snowflake account. This is written in the following format: `<Database>.<Schema>.<Stage Name>`.

* `warehouse` (Required) - The name of the Snowflake warehouse.

##### Veeva Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Veeva resource.

##### Zendesk Connector Profile Properties Arguments

* `instance_url` (Required) - The location of the Zendesk resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the connector profile.

* `credentials_arn` - The Amazon Resource Name (ARN) of the connector profile credentials.

## Import

AppFlow Connector Profile can be imported using the connector profile name, e.g.

```
$ terraform import aws_appflow_connector_profile.profile connector-profile-name
```

[1]: https://docs.aws.amazon.com/appflow/1.0/APIReference/Welcome.html
[2]: https://docs.aws.amazon.com/appflow/1.0/APIReference/API_CreateConnectorProfile.html

