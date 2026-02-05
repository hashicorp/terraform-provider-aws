---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_connection"
description: |-
  Provides an Glue Connection resource.
---

# Resource: aws_glue_connection

Provides a Glue Connection resource.

## Example Usage

### Non-VPC Connection

```terraform
resource "aws_glue_connection" "example" {
  name = "example"
  connection_properties = {
    JDBC_CONNECTION_URL = "jdbc:mysql://example.com/exampledatabase"
    PASSWORD            = "examplepassword"
    USERNAME            = "exampleusername"
  }
}
```

### Non-VPC Connection with secret manager reference

```terraform
data "aws_secretsmanager_secret" "example" {
  name = "example-secret"
}

resource "aws_glue_connection" "example" {
  name = "example"
  connection_properties = {
    JDBC_CONNECTION_URL = "jdbc:mysql://example.com/exampledatabase"
    SECRET_ID           = data.aws_secretsmanager_secret.example.name
  }
}
```

### VPC Connection

For more information, see the [AWS Documentation](https://docs.aws.amazon.com/glue/latest/dg/populate-add-connection.html#connection-JDBC-VPC).

```terraform
resource "aws_glue_connection" "example" {
  name = "example"
  connection_properties = {
    JDBC_CONNECTION_URL = "jdbc:mysql://${aws_rds_cluster.example.endpoint}/exampledatabase"
    PASSWORD            = "examplepassword"
    USERNAME            = "exampleusername"
  }
  physical_connection_requirements {
    availability_zone      = aws_subnet.example.availability_zone
    security_group_id_list = [aws_security_group.example.id]
    subnet_id              = aws_subnet.example.id
  }
}
```

### Connection using a custom connector

```terraform
# Define the custom connector using the connection_type of `CUSTOM` with the match_criteria of `template_connection`
# Example here being a snowflake jdbc connector with a secret having user and password as keys

data "aws_secretsmanager_secret" "example" {
  name = "example-secret"
}

resource "aws_glue_connection" "example1" {
  name            = "example1"
  connection_type = "CUSTOM"
  connection_properties = {
    CONNECTOR_CLASS_NAME = "net.snowflake.client.jdbc.SnowflakeDriver"
    CONNECTION_TYPE      = "Jdbc"
    CONNECTOR_URL        = "s3://example/snowflake-jdbc.jar" # S3 path to the snowflake jdbc jar
    JDBC_CONNECTION_URL  = "[[\"default=jdbc:snowflake://example.com/?user=$${user}&password=$${password}\"],\",\"]"
  }
  match_criteria = ["template-connection"]
}

# Reference the connector using match_criteria with the connector created above.

resource "aws_glue_connection" "example2" {
  name            = "example2"
  connection_type = "CUSTOM"
  connection_properties = {
    CONNECTOR_CLASS_NAME = "net.snowflake.client.jdbc.SnowflakeDriver"
    CONNECTION_TYPE      = "Jdbc"
    CONNECTOR_URL        = "s3://example/snowflake-jdbc.jar"
    JDBC_CONNECTION_URL  = "jdbc:snowflake://example.com/?user=$${user}&password=$${password}"
    SECRET_ID            = data.aws_secretsmanager_secret.example.name
  }
  match_criteria = ["Connection", aws_glue_connection.example1.name]
}
```

### Azure Cosmos Connection

For more information, see the [AWS Documentation](https://docs.aws.amazon.com/glue/latest/dg/connection-properties.html#connection-properties-azurecosmos).

```terraform
resource "aws_secretsmanager_secret" "example" {
  name = "example-secret"
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({
    username = "exampleusername"
    password = "examplepassword"
  })
}

resource "aws_glue_connection" "example" {
  name            = "example"
  connection_type = "AZURECOSMOS"
  connection_properties = {
    SparkProperties = jsonencode({
      secretId                       = aws_secretsmanager_secret.example.name
      "spark.cosmos.accountEndpoint" = "https://exampledbaccount.documents.azure.com:443/"
    })
  }
}
```

### Azure SQL Connection

For more information, see the [AWS Documentation](https://docs.aws.amazon.com/glue/latest/dg/connection-properties.html#connection-properties-azuresql).

```terraform
resource "aws_secretsmanager_secret" "example" {
  name = "example-secret"
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({
    username = "exampleusername"
    password = "examplepassword"
  })
}

resource "aws_glue_connection" "example" {
  name            = "example"
  connection_type = "AZURECOSMOS"
  connection_properties = {
    SparkProperties = jsonencode({
      secretId = aws_secretsmanager_secret.example.name
      url      = "jdbc:sqlserver:exampledbserver.database.windows.net:1433;database=exampledatabase"
    })
  }
}
```

### Google BigQuery Connection

For more information, see the [AWS Documentation](https://docs.aws.amazon.com/glue/latest/dg/connection-properties.html#connection-properties-bigquery).

```terraform
resource "aws_secretsmanager_secret" "example" {
  name = "example-secret"
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({
    credentials = base64encode(<<-EOT
      {
        "type": "service_account",
        "project_id": "example-project",
        "private_key_id": "example-key",
        "private_key": "-----BEGIN RSA PRIVATE KEY-----\nREDACTED\n-----END RSA PRIVATE KEY-----",
        "client_email": "example-project@appspot.gserviceaccount.com",
        "client_id": example-client",
        "auth_uri": "https://accounts.google.com/o/oauth2/auth",
        "token_uri": "https://oauth2.googleapis.com/token",
        "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
        "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/example-project%%40appspot.gserviceaccount.com",
        "universe_domain": "googleapis.com"
      }
      EOT
    )
  })
}

resource "aws_glue_connection" "example" {
  name            = "example"
  connection_type = "BIGQUERY"
  connection_properties = {
    SparkProperties = jsonencode({
      secretId = aws_secretsmanager_secret.example.name
    })
  }
}
```

### OpenSearch Service Connection

For more information, see the [AWS Documentation](https://docs.aws.amazon.com/glue/latest/dg/connection-properties.html#connection-properties-opensearch).

```terraform
resource "aws_secretsmanager_secret" "example" {
  name = "example-secret"
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({
    "opensearch.net.http.auth.user" = "exampleusername"
    "opensearch.net.http.auth.pass" = "examplepassword"
  })
}

resource "aws_glue_connection" "example" {
  name            = "example"
  connection_type = "OPENSEARCH"
  connection_properties = {
    SparkProperties = jsonencode({
      secretId                       = aws_secretsmanager_secret.example.name
      "opensearch.nodes"             = "https://search-exampledomain-ixlmh4jieahrau3bfebcgp8cnm.us-east-1.es.amazonaws.com"
      "opensearch.port"              = "443"
      "opensearch.aws.sigv4.region"  = "us-east-1"
      "opensearch.nodes.wan.only"    = "true"
      "opensearch.aws.sigv4.enabled" = "true"
    })
  }
}
```

### Snowflake Connection

For more information, see the [AWS Documentation](https://docs.aws.amazon.com/glue/latest/dg/connection-properties.html#connection-properties-snowflake).

```terraform
resource "aws_secretsmanager_secret" "example" {
  name = "example-secret"
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({
    sfUser     = "exampleusername"
    sfPassword = "examplepassword"
  })
}

resource "aws_glue_connection" "example" {
  name            = "example"
  connection_type = "SNOWFLAKE"
  connection_properties = {
    SparkProperties = jsonencode({
      secretId = aws_secretsmanager_secret.example.name
      sfRole   = "EXAMPLEETLROLE"
      sfUrl    = "exampleorg-exampleconnection.snowflakecomputing.com"
    })
  }
}
```

### DynamoDB Connection

```terraform
resource "aws_glue_connection" "test" {
  name = "example"

  connection_type = "DYNAMODB"
  athena_properties = {
    lambda_function_arn      = "arn:aws:lambda:us-east-1:123456789012:function:athenafederatedcatalog_athena_abcdefgh"
    disable_spill_encryption = "false"
    spill_bucket             = "example-bucket"
  }
}
```

### MySQL Federated Connection

```terraform
resource "aws_glue_connection" "example" {
  name            = "athenafederatedcatalog_mysql"
  connection_type = "MYSQL"

  athena_properties = {
    lambda_function_arn = "arn:aws:lambda:us-east-1:123456789012:function:athenafederatedcatalog_mysql"
    spill_bucket        = aws_s3_bucket.example.bucket
  }

  connection_properties = {
    HOST     = aws_rds_cluster.example.endpoint
    PORT     = aws_rds_cluster.example.port
    DATABASE = aws_rds_cluster.example.database_name
  }

  authentication_configuration {
    authentication_type = "BASIC"
    secret_arn          = aws_secretsmanager_secret.example.arn
  }

  physical_connection_requirements {
    availability_zone      = aws_subnet.example.availability_zone
    security_group_id_list = [aws_security_group.example.id]
    subnet_id              = aws_subnet.example.id
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the connection.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `catalog_id` - (Optional) ID of the Data Catalog in which to create the connection. If none is supplied, the AWS account ID is used by default.
* `athena_properties` - (Optional) Map of key-value pairs used as connection properties specific to the Athena compute environment.
* `authentication_configuration` - (Optional) Configuration block for authentication options. See [`authentication_configuration` Block](#authentication_configuration-block) for details.
* `connection_properties` - (Optional) Map of key-value pairs used as parameters for this connection. For more information, see the [AWS Documentation](https://docs.aws.amazon.com/glue/latest/dg/connection-properties.html).

  **Note:** Some connection types require the `SparkProperties` property with a JSON document that contains the actual connection properties. For specific examples, refer to [Example Usage](#example-usage).
* `connection_type` - (Optional) Type of the connection. Valid values: `AZURECOSMOS`, `AZURESQL`, `BIGQUERY`, `CUSTOM`, `DYNAMODB`, `JDBC`, `KAFKA`, `MARKETPLACE`, `MONGODB`, `NETWORK`, `OPENSEARCH`, `SNOWFLAKE`. Defaults to `JDBC`.
* `description` - (Optional) Description of the connection.
* `match_criteria` - (Optional) List of criteria that can be used in selecting this connection.
* `physical_connection_requirements` - (Optional) Map of physical connection requirements, such as VPC and SecurityGroup. See [`physical_connection_requirements` Block](#physical_connection_requirements-block) for details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `authentication_configuration` Block

The `authentication_configuration` configuration block supports the following arguments:

* `authentication_type` - (Required) Type of authentication. Valid values: `BASIC`, `CUSTOM`, `IAM`, `OAUTH2`.
* `basic_authentication_credentials` - (Optional) Basic authentication credentials. See [`basic_authentication_credentials` Block](#basic_authentication_credentials-block) for details.
* `custom_authentication_credentials` - (Optional) Map of custom authentication credentials.
* `kms_key_arn` - (Optional) ARN of the KMS key used for encryption.
* `oauth2_properties` - (Optional) OAuth2 properties. See [`oauth2_properties` Block](#oauth2_properties-block) for details.
* `secret_arn` - (Optional) ARN of the Secrets Manager secret containing credentials.

#### `basic_authentication_credentials` Block

The `basic_authentication_credentials` configuration block supports the following arguments:

* `password` - (Required) Password for authentication.
* `username` - (Required) Username for authentication.

#### `oauth2_properties` Block

The `oauth2_properties` configuration block supports the following arguments:

* `authorization_code_properties` - (Optional) Authorization code properties. See [`authorization_code_properties` Block](#authorization_code_properties-block) for details.
* `oauth2_client_application` - (Optional) OAuth2 client application details. See [`oauth2_client_application` Block](#oauth2_client_application-block) for details.
* `oauth2_credentials` - (Optional) OAuth2 credentials. See [`oauth2_credentials` Block](#oauth2_credentials-block) for details.
* `oauth2_grant_type` - (Optional) OAuth2 grant type. Valid values: `AUTHORIZATION_CODE`, `CLIENT_CREDENTIALS`, `JWT_BEARER`.
* `token_url` - (Optional) Token URL for OAuth2 authentication.
* `token_url_parameters_map` - (Optional) Map of additional parameters for the token URL.

##### `authorization_code_properties` Block

The `authorization_code_properties` configuration block supports the following arguments:

* `authorization_code` - (Required) Authorization code.
* `redirect_uri` - (Required) Redirect URI for OAuth2 flow.

##### `oauth2_client_application` Block

The `oauth2_client_application` configuration block supports the following arguments:

* `aws_managed_client_application_reference` - (Optional) Reference to an AWS-managed client application.
* `user_managed_client_application_client_id` - (Optional) Client ID for a user-managed client application.

##### `oauth2_credentials` Block

The `oauth2_credentials` configuration block supports the following arguments:

* `access_token` - (Optional) OAuth2 access token.
* `jwt_token` - (Optional) JWT token.
* `refresh_token` - (Optional) OAuth2 refresh token.
* `user_managed_client_application_client_secret` - (Optional) Client secret for user-managed client application.

### `physical_connection_requirements` Block

The `physical_connection_requirements` configuration block supports the following arguments:

* `availability_zone` - (Optional) The availability zone of the connection. This field is redundant and implied by `subnet_id`, but is currently an api requirement.
* `security_group_id_list` - (Optional) The security group ID list used by the connection.
* `subnet_id` - (Optional) The subnet ID used by the connection.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Glue Connection.
* `id` - Catalog ID and name of the connection.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Connections using the `CATALOG-ID` (AWS account ID if not custom) and `NAME`. For example:

```terraform
import {
  to = aws_glue_connection.MyConnection
  id = "123456789012:MyConnection"
}
```

Using `terraform import`, import Glue Connections using the `CATALOG-ID` (AWS account ID if not custom) and `NAME`. For example:

```console
% terraform import aws_glue_connection.MyConnection 123456789012:MyConnection
```
