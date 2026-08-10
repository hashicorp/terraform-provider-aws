---
subcategory: "AppFlow"
layout: "aws"
page_title: "AWS: aws_appflow_connector_profile"
description: |-
  Provides an AppFlow Connector Profile resource.
---

# Resource: aws_appflow_connector_profile

Provides an AppFlow connector profile resource.

For information about AppFlow flows, see the [Amazon AppFlow API Reference](https://docs.aws.amazon.com/appflow/1.0/APIReference/Welcome.html).
For specific information about creating an AppFlow connector profile, see the
[CreateConnectorProfile](https://docs.aws.amazon.com/appflow/1.0/APIReference/API_CreateConnectorProfile.html) page in the Amazon AppFlow API Reference.

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
  bucket = "example-bucket"
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

This resource supports the following arguments:

* `connection_mode` - (Required) Connection mode and specifies whether it is public or private. Private flows use AWS PrivateLink to route data over AWS infrastructure without exposing it to the public internet. One of: `Public`, `Private`.
* `connector_label` - (Optional) Label of the connector. The label is unique for each `ConnectorRegistration` in your AWS account. Only needed if calling for the `CustomConnector` connector type.
* `connector_profile_config` - (Required) Connector-specific configuration and credentials. See [`connector_profile_config` Block](#connector_profile_config-block) for details.
* `connector_type` - (Required) Type of connector. One of: `Amplitude`, `CustomConnector`, `CustomerProfiles`, `Datadog`, `Dynatrace`, `EventBridge`, `Googleanalytics`, `Honeycode`, `Infornexus`, `LookoutMetrics`, `Marketo`, `Redshift`, `S3`, `Salesforce`, `SAPOData`, `Servicenow`, `Singular`, `Slack`, `Snowflake`, `Trendmicro`, `Upsolver`, `Veeva`, `Zendesk`.
* `kms_arn` - (Optional) ARN of the Key Management Service (KMS) key you provide for encryption. This is required if you do not want to use the Amazon AppFlow-managed KMS key. If you don't provide anything here, Amazon AppFlow uses the Amazon AppFlow-managed KMS key.
* `name` - (Required) Name of the connector profile. The name is unique for each `ConnectorProfile` in your AWS account.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `connector_profile_config` Block

* `connector_profile_credentials` - (Required) Connector-specific credentials required by each connector. See [`connector_profile_credentials` Block](#connector_profile_credentials-block) for details.
* `connector_profile_properties` - (Required) Connector-specific properties of the profile configuration. See [`connector_profile_properties` Block](#connector_profile_properties-block) for details.

### `connector_profile_credentials` Block

* `amplitude` - (Optional) Connector-specific credentials required when using Amplitude. See [`connector_profile_config.connector_profile_credentials.amplitude` Block](#connector_profile_configconnector_profile_credentialsamplitude-block) for details.
* `custom_connector` - (Optional) Connector-specific profile credentials required when using the custom connector. See [`connector_profile_config.connector_profile_credentials.custom_connector` Block](#connector_profile_configconnector_profile_credentialscustom_connector-block) for details.
* `datadog` - (Optional) Connector-specific credentials required when using Datadog. See [`connector_profile_config.connector_profile_credentials.datadog` Block](#connector_profile_configconnector_profile_credentialsdatadog-block) for details.
* `dynatrace` - (Optional) Connector-specific credentials required when using Dynatrace. See [`connector_profile_config.connector_profile_credentials.dynatrace` Block](#connector_profile_configconnector_profile_credentialsdynatrace-block) for details.
* `google_analytics` - (Optional) Connector-specific credentials required when using Google Analytics. See [`connector_profile_config.connector_profile_credentials.google_analytics` Block](#connector_profile_configconnector_profile_credentialsgoogle_analytics-block) for details.
* `honeycode` - (Optional) Connector-specific credentials required when using Amazon Honeycode. See [`connector_profile_config.connector_profile_credentials.honeycode` Block](#connector_profile_configconnector_profile_credentialshoneycode-block) for details.
* `infor_nexus` - (Optional) Connector-specific credentials required when using Infor Nexus. See [`connector_profile_config.connector_profile_credentials.infor_nexus` Block](#connector_profile_configconnector_profile_credentialsinfor_nexus-block) for details.
* `marketo` - (Optional) Connector-specific credentials required when using Marketo. See [`connector_profile_config.connector_profile_credentials.marketo` Block](#connector_profile_configconnector_profile_credentialsmarketo-block) for details.
* `redshift` - (Optional) Connector-specific credentials required when using Amazon Redshift. See [`connector_profile_config.connector_profile_credentials.redshift` Block](#connector_profile_configconnector_profile_credentialsredshift-block) for details.
* `salesforce` - (Optional) Connector-specific credentials required when using Salesforce. See [`connector_profile_config.connector_profile_credentials.salesforce` Block](#connector_profile_configconnector_profile_credentialssalesforce-block) for details.
* `sapo_data` - (Optional) Connector-specific credentials required when using SAPOData. See [`connector_profile_config.connector_profile_credentials.sapo_data` Block](#connector_profile_configconnector_profile_credentialssapo_data-block) for details.
* `service_now` - (Optional) Connector-specific credentials required when using ServiceNow. See [`connector_profile_config.connector_profile_credentials.service_now` Block](#connector_profile_configconnector_profile_credentialsservice_now-block) for details.
* `singular` - (Optional) Connector-specific credentials required when using Singular. See [`connector_profile_config.connector_profile_credentials.singular` Block](#connector_profile_configconnector_profile_credentialssingular-block) for details.
* `slack` - (Optional) Connector-specific credentials required when using Slack. See [`connector_profile_config.connector_profile_credentials.slack` Block](#connector_profile_configconnector_profile_credentialsslack-block) for details.
* `snowflake` - (Optional) Connector-specific credentials required when using Snowflake. See [`connector_profile_config.connector_profile_credentials.snowflake` Block](#connector_profile_configconnector_profile_credentialssnowflake-block) for details.
* `trendmicro` - (Optional) Connector-specific credentials required when using Trend Micro. See [`connector_profile_config.connector_profile_credentials.trendmicro` Block](#connector_profile_configconnector_profile_credentialstrendmicro-block) for details.
* `veeva` - (Optional) Connector-specific credentials required when using Veeva. See [`connector_profile_config.connector_profile_credentials.veeva` Block](#connector_profile_configconnector_profile_credentialsveeva-block) for details.
* `zendesk` - (Optional) Connector-specific credentials required when using Zendesk. See [`connector_profile_config.connector_profile_credentials.zendesk` Block](#connector_profile_configconnector_profile_credentialszendesk-block) for details.

### `connector_profile_config.connector_profile_credentials.amplitude` Block

* `api_key` - (Required) Unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.
* `secret_key` - (Required) Secret Access Key portion of the credentials.

### `connector_profile_config.connector_profile_credentials.custom_connector` Block

* `api_key` - (Optional) API keys required for the authentication of the user. See [`connector_profile_config.connector_profile_credentials.custom_connector.api_key` Block](#connector_profile_configconnector_profile_credentialscustom_connectorapi_key-block) for details.
* `authentication_type` - (Required) Authentication type that the custom connector uses for authenticating while creating a connector profile. One of: `APIKEY`, `BASIC`, `CUSTOM`, `OAUTH2`.
* `basic` - (Optional) Basic credentials that are required for the authentication of the user. See [`connector_profile_config.connector_profile_credentials.custom_connector.basic` Block](#connector_profile_configconnector_profile_credentialscustom_connectorbasic-block) for details.
* `custom` - (Optional) Credentials required when the connector uses the custom authentication mechanism. See [`connector_profile_config.connector_profile_credentials.custom_connector.custom` Block](#connector_profile_configconnector_profile_credentialscustom_connectorcustom-block) for details.
* `oauth2` - (Optional) OAuth 2.0 credentials required for the authentication of the user. See [`connector_profile_config.connector_profile_credentials.custom_connector.oauth2` Block](#connector_profile_configconnector_profile_credentialscustom_connectoroauth2-block) for details.

### `connector_profile_config.connector_profile_credentials.custom_connector.api_key` Block

* `api_key` - (Required) API key required for API key authentication.
* `api_secret_key` - (Optional) API secret key required for API key authentication.

### `connector_profile_config.connector_profile_credentials.custom_connector.basic` Block

* `password` - (Required) Password to use to connect to a resource.
* `username` - (Required) Username to use to connect to a resource.

### `connector_profile_config.connector_profile_credentials.custom_connector.custom` Block

* `credentials_map` - (Optional) Map that holds custom authentication credentials.
* `custom_authentication_type` - (Required) Custom authentication type that the connector uses.

### `connector_profile_config.connector_profile_credentials.custom_connector.oauth2` Block

* `access_token` - (Optional) Access token used to access the connector on your behalf.
* `client_id` - (Optional) Identifier for the desired client.
* `client_secret` - (Optional) Client secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) OAuth requirement needed to request security tokens from the connector endpoint. See [`connector_profile_config.connector_profile_credentials.custom_connector.oauth2.oauth_request` Block](#connector_profile_configconnector_profile_credentialscustom_connectoroauth2oauth_request-block) for details.
* `refresh_token` - (Optional) Refresh token used to refresh an expired access token.

### `connector_profile_config.connector_profile_credentials.custom_connector.oauth2.oauth_request` Block

* `auth_code` - (Optional) Code provided by the connector when it has been authenticated via the connected app.
* `redirect_uri` - (Optional) URL to which the authentication server redirects the browser after authorization has been granted.

### `connector_profile_config.connector_profile_credentials.datadog` Block

* `api_key` - (Required) Unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.
* `application_key` - (Required) Application key, used in conjunction with your API key, that gives you full access to Datadog's programmatic API. Application keys are associated with the user account that created them and are used to log all requests made to the API.

### `connector_profile_config.connector_profile_credentials.dynatrace` Block

* `api_token` - (Required) API token used by the Dynatrace API to authenticate various API calls.

### `connector_profile_config.connector_profile_credentials.google_analytics` Block

* `access_token` - (Optional) Credentials used to access protected Google Analytics resources.
* `client_id` - (Required) Identifier for the desired client.
* `client_secret` - (Required) Client secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) OAuth requirement needed to request security tokens from the connector endpoint. See [`connector_profile_config.connector_profile_credentials.google_analytics.oauth_request` Block](#connector_profile_configconnector_profile_credentialsgoogle_analyticsoauth_request-block) for details.
* `refresh_token` - (Optional) Credentials used to acquire new access tokens. This is required only for OAuth2 access tokens, and is not required for OAuth1 access tokens.

### `connector_profile_config.connector_profile_credentials.google_analytics.oauth_request` Block

* `auth_code` - (Optional) Code provided by the connector when it has been authenticated via the connected app.
* `redirect_uri` - (Optional) URL to which the authentication server redirects the browser after authorization has been granted.

### `connector_profile_config.connector_profile_credentials.honeycode` Block

* `access_token` - (Optional) Credentials used to access protected Amazon Honeycode resources.
* `oauth_request` - (Optional) OAuth requirement needed to request security tokens from the connector endpoint. See [`connector_profile_config.connector_profile_credentials.honeycode.oauth_request` Block](#connector_profile_configconnector_profile_credentialshoneycodeoauth_request-block) for details.
* `refresh_token` - (Optional) Credentials used to acquire new access tokens.

### `connector_profile_config.connector_profile_credentials.honeycode.oauth_request` Block

* `auth_code` - (Optional) Code provided by the connector when it has been authenticated via the connected app.
* `redirect_uri` - (Optional) URL to which the authentication server redirects the browser after authorization has been granted.

### `connector_profile_config.connector_profile_credentials.infor_nexus` Block

* `access_key_id` - (Required) Access Key portion of the credentials.
* `datakey` - (Required) Encryption keys used to encrypt data.
* `secret_access_key` - (Required) Secret key used to sign requests.
* `user_id` - (Required) Identifier for the user.

### `connector_profile_config.connector_profile_credentials.marketo` Block

* `access_token` - (Optional) Credentials used to access protected Marketo resources.
* `client_id` - (Required) Identifier for the desired client.
* `client_secret` - (Required) Client secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) OAuth requirement needed to request security tokens from the connector endpoint. See [`connector_profile_config.connector_profile_credentials.marketo.oauth_request` Block](#connector_profile_configconnector_profile_credentialsmarketooauth_request-block) for details.

### `connector_profile_config.connector_profile_credentials.marketo.oauth_request` Block

* `auth_code` - (Optional) Code provided by the connector when it has been authenticated via the connected app.
* `redirect_uri` - (Optional) URL to which the authentication server redirects the browser after authorization has been granted.

### `connector_profile_config.connector_profile_credentials.redshift` Block

* `password` - (Required) Password that corresponds to the user name.
* `username` - (Required) Name of the user.

### `connector_profile_config.connector_profile_credentials.salesforce` Block

* `access_token` - (Optional) Credentials used to access protected Salesforce resources.
* `client_credentials_arn` - (Optional) Secret manager ARN, which contains the client ID and client secret of the connected app.
* `jwt_token` - (Optional) JSON web token (JWT) that authorizes access to Salesforce records.
* `oauth2_grant_type` - (Optional) OAuth 2.0 grant type used when requesting an access token from Salesforce. Valid values are `CLIENT_CREDENTIALS`, `AUTHORIZATION_CODE`, and `JWT_BEARER`.
* `oauth_request` - (Optional) OAuth requirement needed to request security tokens from the connector endpoint. See [`connector_profile_config.connector_profile_credentials.salesforce.oauth_request` Block](#connector_profile_configconnector_profile_credentialssalesforceoauth_request-block) for details.
* `refresh_token` - (Optional) Credentials used to acquire new access tokens.

### `connector_profile_config.connector_profile_credentials.salesforce.oauth_request` Block

* `auth_code` - (Optional) Code provided by the connector when it has been authenticated via the connected app.
* `redirect_uri` - (Optional) URL to which the authentication server redirects the browser after authorization has been granted.

### `connector_profile_config.connector_profile_credentials.sapo_data` Block

* `basic_auth_credentials` - (Optional) SAPOData basic authentication credentials. See [`connector_profile_config.connector_profile_credentials.sapo_data.basic_auth_credentials` Block](#connector_profile_configconnector_profile_credentialssapo_databasic_auth_credentials-block) for details.
* `oauth_credentials` - (Optional) SAPOData OAuth type authentication credentials. See [`connector_profile_config.connector_profile_credentials.sapo_data.oauth_credentials` Block](#connector_profile_configconnector_profile_credentialssapo_dataoauth_credentials-block) for details.

### `connector_profile_config.connector_profile_credentials.sapo_data.basic_auth_credentials` Block

* `password` - (Required) Password to use to connect to a resource.
* `username` - (Required) Username to use to connect to a resource.

### `connector_profile_config.connector_profile_credentials.sapo_data.oauth_credentials` Block

* `access_token` - (Optional) Access token used to access protected SAPOData resources.
* `client_id` - (Required) Identifier for the desired client.
* `client_secret` - (Required) Client secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) OAuth requirement needed to request security tokens from the connector endpoint. See [`connector_profile_config.connector_profile_credentials.sapo_data.oauth_credentials.oauth_request` Block](#connector_profile_configconnector_profile_credentialssapo_dataoauth_credentialsoauth_request-block) for details.
* `refresh_token` - (Optional) Refresh token used to refresh an expired access token.

### `connector_profile_config.connector_profile_credentials.sapo_data.oauth_credentials.oauth_request` Block

* `auth_code` - (Optional) Code provided by the connector when it has been authenticated via the connected app.
* `redirect_uri` - (Optional) URL to which the authentication server redirects the browser after authorization has been granted.

### `connector_profile_config.connector_profile_credentials.service_now` Block

* `password` - (Required) Password that corresponds to the user name.
* `username` - (Required) Name of the user.

### `connector_profile_config.connector_profile_credentials.singular` Block

* `api_key` - (Required) Unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.

### `connector_profile_config.connector_profile_credentials.slack` Block

* `access_token` - (Optional) Credentials used to access protected Slack resources.
* `client_id` - (Required) Identifier for the client.
* `client_secret` - (Required) Client secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) OAuth requirement needed to request security tokens from the connector endpoint. See [`connector_profile_config.connector_profile_credentials.slack.oauth_request` Block](#connector_profile_configconnector_profile_credentialsslackoauth_request-block) for details.

### `connector_profile_config.connector_profile_credentials.slack.oauth_request` Block

* `auth_code` - (Optional) Code provided by the connector when it has been authenticated via the connected app.
* `redirect_uri` - (Optional) URL to which the authentication server redirects the browser after authorization has been granted.

### `connector_profile_config.connector_profile_credentials.snowflake` Block

* `password` - (Required) Password that corresponds to the user name.
* `username` - (Required) Name of the user.

### `connector_profile_config.connector_profile_credentials.trendmicro` Block

* `api_secret_key` - (Required) Secret Access Key portion of the credentials.

### `connector_profile_config.connector_profile_credentials.veeva` Block

* `password` - (Required) Password that corresponds to the user name.
* `username` - (Required) Name of the user.

### `connector_profile_config.connector_profile_credentials.zendesk` Block

* `access_token` - (Optional) Credentials used to access protected Zendesk resources.
* `client_id` - (Required) Identifier for the desired client.
* `client_secret` - (Required) Client secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) OAuth requirement needed to request security tokens from the connector endpoint. See [`connector_profile_config.connector_profile_credentials.zendesk.oauth_request` Block](#connector_profile_configconnector_profile_credentialszendeskoauth_request-block) for details.

### `connector_profile_config.connector_profile_credentials.zendesk.oauth_request` Block

* `auth_code` - (Optional) Code provided by the connector when it has been authenticated via the connected app.
* `redirect_uri` - (Optional) URL to which the authentication server redirects the browser after authorization has been granted.

### `connector_profile_properties` Block

* `custom_connector` - (Optional) Connector-specific profile properties required when using the custom connector. See [`connector_profile_config.connector_profile_properties.custom_connector` Block](#connector_profile_configconnector_profile_propertiescustom_connector-block) for details.
* `datadog` - (Optional) Connector-specific properties required when using Datadog. See [`connector_profile_config.connector_profile_properties.datadog` Block](#connector_profile_configconnector_profile_propertiesdatadog-block) for details.
* `dynatrace` - (Optional) Connector-specific properties required when using Dynatrace. See [`connector_profile_config.connector_profile_properties.dynatrace` Block](#connector_profile_configconnector_profile_propertiesdynatrace-block) for details.
* `infor_nexus` - (Optional) Connector-specific properties required when using Infor Nexus. See [`connector_profile_config.connector_profile_properties.infor_nexus` Block](#connector_profile_configconnector_profile_propertiesinfor_nexus-block) for details.
* `marketo` - (Optional) Connector-specific properties required when using Marketo. See [`connector_profile_config.connector_profile_properties.marketo` Block](#connector_profile_configconnector_profile_propertiesmarketo-block) for details.
* `redshift` - (Optional) Connector-specific properties required when using Amazon Redshift. See [`connector_profile_config.connector_profile_properties.redshift` Block](#connector_profile_configconnector_profile_propertiesredshift-block) for details.
* `salesforce` - (Optional) Connector-specific properties required when using Salesforce. See [`connector_profile_config.connector_profile_properties.salesforce` Block](#connector_profile_configconnector_profile_propertiessalesforce-block) for details.
* `sapo_data` - (Optional) Connector-specific properties required when using SAPOData. See [`connector_profile_config.connector_profile_properties.sapo_data` Block](#connector_profile_configconnector_profile_propertiessapo_data-block) for details.
* `service_now` - (Optional) Connector-specific properties required when using ServiceNow. See [`connector_profile_config.connector_profile_properties.service_now` Block](#connector_profile_configconnector_profile_propertiesservice_now-block) for details.
* `slack` - (Optional) Connector-specific properties required when using Slack. See [`connector_profile_config.connector_profile_properties.slack` Block](#connector_profile_configconnector_profile_propertiesslack-block) for details.
* `snowflake` - (Optional) Connector-specific properties required when using Snowflake. See [`connector_profile_config.connector_profile_properties.snowflake` Block](#connector_profile_configconnector_profile_propertiessnowflake-block) for details.
* `veeva` - (Optional) Connector-specific properties required when using Veeva. See [`connector_profile_config.connector_profile_properties.veeva` Block](#connector_profile_configconnector_profile_propertiesveeva-block) for details.
* `zendesk` - (Optional) Connector-specific properties required when using Zendesk. See [`connector_profile_config.connector_profile_properties.zendesk` Block](#connector_profile_configconnector_profile_propertieszendesk-block) for details.

### `connector_profile_config.connector_profile_properties.custom_connector` Block

* `oauth2_properties` - (Optional) OAuth 2.0 properties required for OAuth 2.0 authentication. See [`connector_profile_config.connector_profile_properties.custom_connector.oauth2_properties` Block](#connector_profile_configconnector_profile_propertiescustom_connectoroauth2_properties-block) for details.
* `profile_properties` - (Optional) Map of properties that are required to create a profile for the custom connector.

### `connector_profile_config.connector_profile_properties.custom_connector.oauth2_properties` Block

* `oauth2_grant_type` - (Required) OAuth 2.0 grant type used by the connector for OAuth 2.0 authentication. One of: `AUTHORIZATION_CODE`, `CLIENT_CREDENTIALS`.
* `token_url` - (Required) Token URL required for OAuth 2.0 authentication.
* `token_url_custom_properties` - (Optional) Map of properties associated with your token URL. Use this parameter to provide any additional details that the connector requires to authenticate your request.

### `connector_profile_config.connector_profile_properties.datadog` Block

* `instance_url` - (Required) Location of the Datadog resource.

### `connector_profile_config.connector_profile_properties.dynatrace` Block

* `instance_url` - (Required) Location of the Dynatrace resource.

### `connector_profile_config.connector_profile_properties.infor_nexus` Block

* `instance_url` - (Required) Location of the Infor Nexus resource.

### `connector_profile_config.connector_profile_properties.marketo` Block

* `instance_url` - (Required) Location of the Marketo resource.

### `connector_profile_config.connector_profile_properties.redshift` Block

* `bucket_name` - (Required) Name for the associated Amazon S3 bucket.
* `bucket_prefix` - (Optional) Object key for the destination bucket in which Amazon AppFlow places the files.
* `cluster_identifier` - (Optional) Unique ID that's assigned to an Amazon Redshift cluster.
* `data_api_role_arn` - (Optional) ARN of the IAM role that permits AppFlow to access the database through Data API.
* `database_name` - (Optional) Name of an Amazon Redshift database.
* `database_url` - (Required) JDBC URL of the Amazon Redshift cluster.
* `role_arn` - (Required) ARN of the IAM role.

### `connector_profile_config.connector_profile_properties.salesforce` Block

* `instance_url` - (Optional) Location of the Salesforce resource.
* `is_sandbox_environment` - (Optional) Whether the connector profile applies to a sandbox or production environment.
* `use_privatelink_for_metadata_and_authorization` - (Optional) Whether Amazon AppFlow uses the private network to send metadata and authorization calls to Salesforce. Amazon AppFlow sends private calls through AWS PrivateLink. These calls travel through AWS infrastructure without being exposed to the public internet.

### `connector_profile_config.connector_profile_properties.sapo_data` Block

* `application_host_url` - (Required) Location of the SAPOData resource.
* `application_service_path` - (Required) Application path to catalog service.
* `client_number` - (Required) Client number for the client creating the connection.
* `logon_language` - (Optional) Logon language of the SAPOData instance.
* `oauth_properties` - (Optional) SAPOData OAuth properties required for OAuth type authentication. See [`connector_profile_config.connector_profile_properties.sapo_data.oauth_properties` Block](#connector_profile_configconnector_profile_propertiessapo_dataoauth_properties-block) for details.
* `port_number` - (Required) Port number of the SAPOData instance.
* `private_link_service_name` - (Optional) SAPOData Private Link service name to be used for private data transfers.

### `connector_profile_config.connector_profile_properties.sapo_data.oauth_properties` Block

* `auth_code_url` - (Required) Authorization code URL required to redirect to the SAP Login Page to fetch the authorization code for OAuth type authentication.
* `oauth_scopes` - (Required) OAuth scopes required for OAuth type authentication.
* `token_url` - (Required) Token URL required to fetch access and refresh tokens using the authorization code, and to refresh an expired access token using the refresh token.

### `connector_profile_config.connector_profile_properties.service_now` Block

* `instance_url` - (Required) Location of the ServiceNow resource.

### `connector_profile_config.connector_profile_properties.slack` Block

* `instance_url` - (Required) Location of the Slack resource.

### `connector_profile_config.connector_profile_properties.snowflake` Block

* `account_name` - (Optional) Name of the account.
* `bucket_name` - (Required) Name of the Amazon S3 bucket associated with Snowflake.
* `bucket_prefix` - (Optional) Bucket path that refers to the Amazon S3 bucket associated with Snowflake.
* `private_link_service_name` - (Optional) Snowflake Private Link service name to be used for private data transfers.
* `region` - (Optional) AWS Region of the Snowflake account.
* `stage` - (Required) Name of the Amazon S3 stage that was created while setting up an Amazon S3 stage in the Snowflake account. This is written in the following format: `<Database>.<Schema>.<Stage Name>`.
* `warehouse` - (Required) Name of the Snowflake warehouse.

### `connector_profile_config.connector_profile_properties.veeva` Block

* `instance_url` - (Required) Location of the Veeva resource.

### `connector_profile_config.connector_profile_properties.zendesk` Block

* `instance_url` - (Required) Location of the Zendesk resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connector profile.
* `credentials_arn` - ARN of the connector profile credentials.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_appflow_connector_profile.example
  identity = {
    name = "example_profile"
  }
}

resource "aws_appflow_connector_profile" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the Appflow connector profile.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppFlow Connector Profile using the connector profile `name`. For example:

```terraform
import {
  to = aws_appflow_connector_profile.example
  id = "example-profile"
}
```

Using `terraform import`, import AppFlow Connector Profile using the connector profile `name`. For example:

```console
% terraform import aws_appflow_connector_profile.example example-profile
```
