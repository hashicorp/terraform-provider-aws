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

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the DB Proxy.
* `auth` - Configuration(s) with authorization mechanisms to connect to the associated instance or cluster.
* `debug_logging` - Whether the proxy includes detailed information about SQL statements in its logs.
* `endpoint` - Endpoint that you can use to connect to the DB proxy.
* `engine_family` - Kinds of databases that the proxy can connect to.
* `idle_client_timeout` - Number of seconds a connection to the proxy can have no activity before the proxy drops the client connection.
* `require_tls` - Whether Transport Layer Security (TLS) encryption is required for connections to the proxy.
* `role_arn` - ARN for the IAM role that the proxy uses to access Amazon Secrets Manager.
* `vpc_id` - Provides the VPC ID of the DB proxy.
* `vpc_security_group_ids` - Provides a list of VPC security groups that the proxy belongs to.
* `vpc_subnet_ids` - EC2 subnet IDs for the proxy.
