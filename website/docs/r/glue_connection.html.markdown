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
  connection_properties = {
    JDBC_CONNECTION_URL = "jdbc:mysql://example.com/exampledatabase"
    PASSWORD            = "examplepassword"
    USERNAME            = "exampleusername"
  }

  name = "example"
}
```

### Non-VPC Connection with secret manager reference

```terraform

data "aws_secretmanager_secret" "example" {
  name = "example-secret"
}

resource "aws_glue_connection" "example" {
  connection_properties = {
    JDBC_CONNECTION_URL = "jdbc:mysql://example.com/exampledatabase"
    SECRET_ID           = data.aws_secretmanager_secret.example.name
  }

  name = "example"
}
```

### VPC Connection

For more information, see the [AWS Documentation](https://docs.aws.amazon.com/glue/latest/dg/populate-add-connection.html#connection-JDBC-VPC).

```terraform
resource "aws_glue_connection" "example" {
  connection_properties = {
    JDBC_CONNECTION_URL = "jdbc:mysql://${aws_rds_cluster.example.endpoint}/exampledatabase"
    PASSWORD            = "examplepassword"
    USERNAME            = "exampleusername"
  }

  name = "example"

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

data "aws_secretmanager_secret" "example" {
  name = "example-secret"
}

resource "aws_glue_connection" "example_connector" {
  connection_type = "CUSTOM"

  connection_properties = {
    CONNECTOR_CLASS_NAME = "net.snowflake.client.jdbc.SnowflakeDriver"
    CONNECTION_TYPE      = "Jdbc"
    CONNECTOR_URL        = "s3://example/snowflake-jdbc.jar" # S3 path to the snowflake jdbc jar
    JDBC_CONNECTION_URL  = "[[\"default=jdbc:snowflake://example.com/?user=$${user}&password=$${password}\"],\",\"]"
  }

  name = "example_connector"

  match_criteria = ["template-connection"]
}

# Reference the connector using match_criteria with the connector created above.

resource "aws_glue_connection" "example_connection" {
  connection_type = "CUSTOM"

  connection_properties = {
    CONNECTOR_CLASS_NAME = "net.snowflake.client.jdbc.SnowflakeDriver"
    CONNECTION_TYPE      = "Jdbc"
    CONNECTOR_URL        = "s3://example/snowflake-jdbc.jar"
    JDBC_CONNECTION_URL  = "jdbc:snowflake://example.com/?user=$${user}&password=$${password}"
    SECRET_ID            = data.aws_secretmanager_secret.example.name
  }
  name           = "example"
  match_criteria = ["Connection", aws_glue_connection.example_connector.name]
}

```

## Argument Reference

This resource supports the following arguments:

* `catalog_id` – (Optional) The ID of the Data Catalog in which to create the connection. If none is supplied, the AWS account ID is used by default.
* `connection_properties` – (Optional) A map of key-value pairs used as parameters for this connection.
* `connection_type` – (Optional) The type of the connection. Supported are: `CUSTOM`, `JDBC`, `KAFKA`, `MARKETPLACE`, `MONGODB`, and `NETWORK`. Defaults to `JDBC`.
* `description` – (Optional) Description of the connection.
* `match_criteria` – (Optional) A list of criteria that can be used in selecting this connection.
* `name` – (Required) The name of the connection.
* `physical_connection_requirements` - (Optional) A map of physical connection requirements, such as VPC and SecurityGroup. Defined below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### physical_connection_requirements

* `availability_zone` - (Optional) The availability zone of the connection. This field is redundant and implied by `subnet_id`, but is currently an api requirement.
* `security_group_id_list` - (Optional) The security group ID list used by the connection.
* `subnet_id` - (Optional) The subnet ID used by the connection.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Catalog ID and name of the connection
* `arn` - The ARN of the Glue Connection.
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
