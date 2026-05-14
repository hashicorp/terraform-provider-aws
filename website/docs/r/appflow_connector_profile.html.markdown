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

* `connection_mode` - (Required) Connection mode and specifies whether it is public or private. Private flows use AWS PrivateLink to route data over AWS infrastructure without exposing it to the public internet. Valid values: `Public`, `Private`.
* `connector_label` - (Optional) Label of the connector. Label is unique for each ConnectorRegistration in your AWS account. Only needed if calling for `CustomConnector` connector type.
* `connector_profile_config` - (Required) Connector-specific configuration and credentials. See [`connector_profile_config`](#connector_profile_config-block) below.
* `connector_type` - (Required) Type of connector. Valid values: `Amplitude`, `CustomConnector`, `CustomerProfiles`, `Datadog`, `Dynatrace`, `EventBridge`, `Googleanalytics`, `Honeycode`, `Infornexus`, `LookoutMetrics`, `Marketo`, `Redshift`, `S3`, `Salesforce`, `SAPOData`, `Servicenow`, `Singular`, `Slack`, `Snowflake`, `Trendmicro`, `Upsolver`, `Veeva`, `Zendesk`.
* `kms_arn` - (Optional) ARN of the Key Management Service (KMS) key you provide for encryption. If you don't provide anything here, Amazon AppFlow uses the Amazon AppFlow-managed KMS key.
* `name` - (Required) Name of the connector profile. The name is unique for each `ConnectorProfile` in your AWS account.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `connector_profile_config` Block

* `connector_profile_credentials` - (Required) Connector-specific credentials required by each connector. See [`connector_profile_credentials`](#connector_profile_credentials-block) below.
* `connector_profile_properties` - (Required) Connector-specific properties of the profile configuration. See [`connector_profile_properties`](#connector_profile_properties-block) below.

### `connector_profile_credentials` Block

* `amplitude` - (Optional) See [`amplitude`](#connector_profile_credentials-amplitude-block) below.
* `custom_connector` - (Optional) See [`custom_connector`](#custom_connector-block) below.
* `datadog` - (Optional) See [`datadog`](#connector_profile_credentials-datadog-block) below.
* `dynatrace` - (Optional) See [`dynatrace`](#connector_profile_credentials-dynatrace-block) below.
* `google_analytics` - (Optional) See [`google_analytics`](#google_analytics-block) below.
* `honeycode` - (Optional) See [`honeycode`](#honeycode-block) below.
* `infor_nexus` - (Optional) See [`infor_nexus`](#infor_nexus-block) below.
* `marketo` - (Optional) See [`marketo`](#connector_profile_credentials-marketo-block) below.
* `redshift` - (Optional) See [`redshift`](#connector_profile_credentials-redshift-block) below.
* `salesforce` - (Optional) See [`salesforce`](#connector_profile_credentials-salesforce-block) below.
* `sapo_data` - (Optional) See [`sapo_data`](#connector_profile_credentials-sapo_data-block) below.
* `service_now` - (Optional) See [`service_now`](#connector_profile_credentials-service_now-block) below.
* `singular` - (Optional) See [`singular`](#singular-block) below.
* `slack` - (Optional) See [`slack`](#connector_profile_credentials-slack-block) below.
* `snowflake` - (Optional) See [`snowflake`](#connector_profile_credentials-snowflake-block) below.
* `trendmicro` - (Optional) See [`trendmicro`](#trendmicro-block) below.
* `veeva` - (Optional) See [`veeva`](#veeva-block) below.
* `zendesk` - (Optional) See [`zendesk`](#connector_profile_credentials-zendesk-block) below.

### `connector_profile_credentials` `amplitude` Block

* `api_key` - (Required) Unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.
* `secret_key` - (Required) Secret Access Key portion of the credentials.

### `custom_connector` Block

* `authentication_type` - (Required) Uauthentication type that the custom connector uses. One of: `APIKEY`, `BASIC`, `CUSTOM`, `OAUTH2`.
* `api_key` - (Optional) API keys required for the authentication of the user. See [`api_key`](#custom_connector-api_key-block) below.
* `basic` - (Optional) Basic credentials for authentication. See [`basic`](#basic-block) below.
* `custom` - (Optional) Custom authentication credentials. See [`custom`](#custom-block) below.
* `oauth2` - (Optional) OAuth 2.0 credentials for authentication. See [`oauth2`](#oauth2-block) below.

### `custom_connector` `api_key` Block

* `api_key` - (Required) API key required for API key authentication.
* `api_secret_key` - (Optional) API secret key required for API key authentication.

### `basic` Block

* `password` - (Required) Upassword to use to connect to a resource.
* `username` - (Required) Uusername to use to connect to a resource.

### `custom` Block

* `credentials_map` - (Optional) Map that holds custom authentication credentials.
* `custom_authentication_type` - (Required) Custom authentication type that the connector uses.

### `oauth2` Block

* `access_token` - (Optional) Uaccess token used to access the connector on your behalf.
* `client_id` - (Optional) Uidentifier for the desired client.
* `client_secret` - (Optional) Uclient secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) See [`oauth_request`](#oauth_request-block) below.
* `refresh_token` - (Optional) Urefresh token used to refresh an expired access token.

### `connector_profile_credentials` `datadog` Block

* `api_key` - (Required) Unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.
* `application_key` - (Required) Application keys, in conjunction with your API key, give you full access to Datadog's programmatic API.

### `connector_profile_credentials` `dynatrace` Block

* `api_token` - (Required) API tokens used by Dynatrace API to authenticate various API calls.

### `google_analytics` Block

* `access_token` - (Optional) Ucredentials used to access protected Google Analytics resources.
* `client_id` - (Required) Uidentifier for the desired client.
* `client_secret` - (Required) Uclient secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) See [`oauth_request`](#oauth_request-block) below.
* `refresh_token` - (Optional) Ucredentials used to acquire new access tokens.

### `honeycode` Block

* `access_token` - (Optional) Ucredentials used to access protected Amazon Honeycode resources.
* `oauth_request` - (Optional) See [`oauth_request`](#oauth_request-block) below.
* `refresh_token` - (Optional) Ucredentials used to acquire new access tokens.

### `infor_nexus` Block

* `access_key_id` - (Required) Access Key portion of the credentials.
* `datakey` - (Required) Encryption keys used to encrypt data.
* `secret_access_key` - (Required) Secret key used to sign requests.
* `user_id` - (Required) Identifier for the user.

### `connector_profile_credentials` `marketo` Block

* `access_token` - (Optional) Ucredentials used to access protected Marketo resources.
* `client_id` - (Required) Uidentifier for the desired client.
* `client_secret` - (Required) Uclient secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) See [`oauth_request`](#oauth_request-block) below.

### `connector_profile_credentials` `redshift` Block

* `password` - (Required) Password that corresponds to the user name.
* `username` - (Required) Name of the user.

### `connector_profile_credentials` `salesforce` Block

* `access_token` - (Optional) Ucredentials used to access protected Salesforce resources.
* `client_credentials_arn` - (Optional) Usecret manager ARN, which contains the client ID and client secret of the connected app.
* `jwt_token` - (Optional) JSON web token (JWT) that authorizes access to Salesforce records.
* `oauth2_grant_type` - (Optional) OAuth 2.0 grant type used when requesting an access token. Valid values are `CLIENT_CREDENTIALS`, `AUTHORIZATION_CODE`, and `JWT_BEARER`.
* `oauth_request` - (Optional) See [`oauth_request`](#oauth_request-block) below.
* `refresh_token` - (Optional) Ucredentials used to acquire new access tokens.

### `connector_profile_credentials` `sapo_data` Block

* `basic_auth_credentials` - (Optional) SAPOData basic authentication credentials. See [`basic_auth_credentials`](#basic_auth_credentials-block) below.
* `oauth_credentials` - (Optional) SAPOData OAuth type authentication credentials. See [`oauth_credentials`](#oauth_credentials-block) below.

### `basic_auth_credentials` Block

* `password` - (Required) Upassword to use to connect to a resource.
* `username` - (Required) Uusername to use to connect to a resource.

### `oauth_credentials` Block

* `access_token` - (Optional) Uaccess token used to access protected SAPOData resources.
* `client_id` - (Required) Uidentifier for the desired client.
* `client_secret` - (Required) Uclient secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) See [`oauth_request`](#oauth_request-block) below.
* `refresh_token` - (Optional) Urefresh token used to refresh expired access token.

### `connector_profile_credentials` `service_now` Block

* `password` - (Required) Password that corresponds to the user name.
* `username` - (Required) Name of the user.

### `singular` Block

* `api_key` - (Required) Unique alphanumeric identifier used to authenticate a user, developer, or calling program to your API.

### `connector_profile_credentials` `slack` Block

* `access_token` - (Optional) Ucredentials used to access protected Slack resources.
* `client_id` - (Required) Uidentifier for the client.
* `client_secret` - (Required) Uclient secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) See [`oauth_request`](#oauth_request-block) below.

### `connector_profile_credentials` `snowflake` Block

* `password` - (Required) Password that corresponds to the user name.
* `username` - (Required) Name of the user.

### `trendmicro` Block

* `api_secret_key` - (Required) Secret Access Key portion of the credentials.

### `veeva` Block

* `password` - (Required) Password that corresponds to the user name.
* `username` - (Required) Name of the user.

### `connector_profile_credentials` `zendesk` Block

* `access_token` - (Optional) Ucredentials used to access protected Zendesk resources.
* `client_id` - (Required) Uidentifier for the desired client.
* `client_secret` - (Required) Uclient secret used by the OAuth client to authenticate to the authorization server.
* `oauth_request` - (Optional) See [`oauth_request`](#oauth_request-block) below.

### `oauth_request` Block

* `auth_code` - (Optional) Ucode provided by the connector when it has been authenticated via the connected app.
* `redirect_uri` - (Optional) URL to which the authentication server redirects the browser after authorization has been granted.
### `connector_profile_properties` Block

* `custom_connector` - (Optional) See [`custom_connector`](#connector_profile_properties-custom_connector-block) below.
* `datadog` - (Optional) See [`datadog`](#connector_profile_properties-datadog-block) below.
* `dynatrace` - (Optional) See [`dynatrace`](#connector_profile_properties-dynatrace-block) below.
* `infor_nexus` - (Optional) See [`infor_nexus`](#connector_profile_properties-infor_nexus-block) below.
* `marketo` - (Optional) See [`marketo`](#connector_profile_properties-marketo-block) below.
* `redshift` - (Optional) See [`redshift`](#connector_profile_properties-redshift-block) below.
* `salesforce` - (Optional) See [`salesforce`](#connector_profile_properties-salesforce-block) below.
* `sapo_data` - (Optional) See [`sapo_data`](#connector_profile_properties-sapo_data-block) below.
* `service_now` - (Optional) See [`service_now`](#connector_profile_properties-service_now-block) below.
* `slack` - (Optional) See [`slack`](#connector_profile_properties-slack-block) below.
* `snowflake` - (Optional) See [`snowflake`](#connector_profile_properties-snowflake-block) below.
* `veeva` - (Optional) See [`veeva`](#connector_profile_properties-veeva-block) below.
* `zendesk` - (Optional) See [`zendesk`](#connector_profile_properties-zendesk-block) below.

### `connector_profile_properties` `custom_connector` Block

* `oauth2_properties` - (Optional) OAuth 2.0 properties required for OAuth 2.0 authentication. See [`oauth2_properties`](#oauth2_properties-block) below.
* `profile_properties` - (Optional) Map of properties that are required to create a profile for the custom connector.

### `oauth2_properties` Block

* `oauth2_grant_type` - (Required) OAuth 2.0 grant type used by connector for OAuth 2.0 authentication. One of: `AUTHORIZATION_CODE`, `CLIENT_CREDENTIALS`.
* `token_url` - (Required) Utoken URL required for OAuth 2.0 authentication.
* `token_url_custom_properties` - (Optional) Associates your token URL with a map of properties that you define.

### `connector_profile_properties` `datadog` Block

* `instance_url` - (Required) Location of the Datadog resource.

### `connector_profile_properties` `dynatrace` Block

* `instance_url` - (Required) Location of the Dynatrace resource.

### `connector_profile_properties` `infor_nexus` Block

* `instance_url` - (Required) Location of the Infor Nexus resource.

### `connector_profile_properties` `marketo` Block

* `instance_url` - (Required) Location of the Marketo resource.

### `connector_profile_properties` `redshift` Block

* `bucket_name` - (Required) Name for the associated Amazon S3 bucket.
* `bucket_prefix` - (Optional) Uobject key for the destination bucket in which Amazon AppFlow places the files.
* `cluster_identifier` - (Optional) Uunique ID that's assigned to an Amazon Redshift cluster.
* `data_api_role_arn` - (Optional) ARN of the IAM role that permits AppFlow to access the database through Data API.
* `database_name` - (Optional) Uname of an Amazon Redshift database.
* `database_url` - (Optional) JDBC URL of the Amazon Redshift cluster.
* `role_arn` - (Required) ARN of the IAM role.

### `connector_profile_properties` `salesforce` Block

* `instance_url` - (Optional) Ulocation of the Salesforce resource.
* `is_sandbox_environment` - (Optional) Whether the connector profile applies to a sandbox or production environment.
* `use_privatelink_for_metadata_and_authorization` - (Optional) Whether Amazon AppFlow uses the private network to send metadata and authorization calls to Salesforce.

### `connector_profile_properties` `sapo_data` Block

* `application_host_url` - (Required) Ulocation of the SAPOData resource.
* `application_service_path` - (Required) Uapplication path to catalog service.
* `client_number` - (Required) Uclient number for the client creating the connection.
* `logon_language` - (Optional) Ulogon language of SAPOData instance.
* `oauth_properties` - (Optional) SAPOData OAuth properties required for OAuth type authentication. See [`oauth_properties`](#oauth_properties-block) below.
* `port_number` - (Required) Uport number of the SAPOData instance.
* `private_link_service_name` - (Optional) SAPOData Private Link service name to be used for private data transfers.

### `oauth_properties` Block

* `auth_code_url` - (Required) Uauthorization code url required to redirect to SAP Login Page to fetch authorization code for OAuth type authentication.
* `oauth_scopes` - (Required) OAuth scopes required for OAuth type authentication.
* `token_url` - (Required) Utoken url required to fetch access/refresh tokens using authorization code and also to refresh expired access token using refresh token.

### `connector_profile_properties` `service_now` Block

* `instance_url` - (Required) Location of the ServiceNow resource.

### `connector_profile_properties` `slack` Block

* `instance_url` - (Required) Location of the Slack resource.

### `connector_profile_properties` `snowflake` Block

* `account_name` - (Optional) Uname of the account.
* `bucket_name` - (Required) Uname of the Amazon S3 bucket associated with Snowflake.
* `bucket_prefix` - (Optional) Ubucket path that refers to the Amazon S3 bucket associated with Snowflake.
* `private_link_service_name` - (Optional) Snowflake Private Link service name to be used for private data transfers.
* `region` - (Optional) AWS Region of the Snowflake account.
* `stage` - (Required) Name of the Amazon S3 stage that was created while setting up an Amazon S3 stage in the Snowflake account. This is written in the following format: `<Database>.<Schema>.<Stage Name>`.
* `warehouse` - (Required) Uname of the Snowflake warehouse.

### `connector_profile_properties` `veeva` Block

* `instance_url` - (Required) Location of the Veeva resource.

### `connector_profile_properties` `zendesk` Block

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

[1]: https://docs.aws.amazon.com/appflow/1.0/APIReference/Welcome.html
[2]: https://docs.aws.amazon.com/appflow/1.0/APIReference/API_CreateConnectorProfile.html
