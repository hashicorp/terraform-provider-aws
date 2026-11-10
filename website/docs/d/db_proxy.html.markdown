---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_proxy"
description: |-
  Get information on a DB Proxy.
---

# Data Source: aws_db_proxy

Use this data source to get information about a DB Proxy.

## Example Usage

```terraform
data "aws_db_proxy" "proxy" {
  name = "my-test-db-proxy"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the DB proxy.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the DB Proxy.
* `auth` - Configuration(s) with authorization mechanisms to connect to the associated instance or cluster. See the [`auth`](#auth-block) block below.
* `debug_logging` - Whether the proxy includes detailed information about SQL statements in its logs.
* `default_auth_scheme` - Default authentication scheme that the proxy uses for client connections to the proxy and connections from the proxy to the underlying database.
* `endpoint` - Endpoint that you can use to connect to the DB proxy.
* `endpoint_network_type` - Network type of the DB proxy endpoint.
* `engine_family` - Kinds of databases that the proxy can connect to.
* `idle_client_timeout` - Number of seconds a connection to the proxy can have no activity before the proxy drops the client connection.
* `require_tls` - Whether TLS encryption is required for connections to the proxy.
* `role_arn` - ARN for the IAM role that the proxy uses to access Amazon Secrets Manager.
* `target_connection_network_type` - Network type that the proxy uses to connect to the target database.
* `vpc_id` - Provides the VPC ID of the DB proxy.
* `vpc_security_group_ids` - Provides a list of VPC security groups that the proxy belongs to.
* `vpc_subnet_ids` - EC2 subnet IDs for the proxy.

### `auth` Block

The `auth` block exports the following attributes:

* `auth_scheme` - Type of authentication that the proxy uses for connections from the proxy to the underlying database.
* `client_password_auth_type` - Type of authentication the proxy uses for connections from clients.
* `description` - User-specified description about the authentication used by a proxy to log in as a specific database user.
* `iam_auth` - Whether to require or disallow AWS Identity and Access Management (IAM) authentication for connections to the proxy.
* `secret_arn` - ARN representing the secret that the proxy uses to authenticate to the RDS DB instance or Aurora DB cluster.
* `username` - Name of the database user to which the proxy connects.
