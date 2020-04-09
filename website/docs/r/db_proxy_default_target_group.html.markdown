---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_db_proxy_default_target_group"
description: |-
  Manage an RDS DB proxy default target group resource.
---

# Resource: aws_db_proxy_default_target_group

Provides a resource to manage an RDS DB proxy default target group resource.

The `aws_db_proxy_default_target_group` behaves differently from normal resources, in that Terraform does not _create_ this resource, but instead "adopts" it into management.

## Example Usage

```hcl
resource "aws_db_proxy" "example" {
  name                   = "example"
  debug_logging          = false
  engine_family          = "MYSQL"
  idle_client_timeout    = 1800
  require_tls            = true
  role_arn               = "arn:aws:iam:us-east-1:123456789012:role/example"
  vpc_security_group_ids = ["sg-12345678901234567"]
  vpc_subnet_ids         = ["subnet-12345678901234567"]

  auth {
    auth_scheme = "SECRETS"
    description = "example"
    iam_auth    = "DISABLED"
    secret_arn  = "arn:aws:secretsmanager:us-east-1:123456789012:secret:example"
  }

  tags = {
    Name = "example"
    Key  = "value"
  }
}

resource "aws_db_proxy_default_target_group" "example" {
  db_proxy_name = aws_db_proxy.example.name

  connection_pool_config {
    connection_borrow_timeout    = 120
    init_query                   = "SET x=1, y=2"
    max_connections_percent      = 100
    max_idle_connections_percent = 50
    session_pinning_filters      = ["EXCLUDE_VARIABLE_SETS"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `db_proxy_name` - (Required) The name of the new proxy to which to assign the target group.
* `name` - (Optional) The name of the new target group to assign to the proxy.
* `connection_pool_config` - (Optional) The settings that determine the size and behavior of the connection pool for the target group.

`connection_pool_config` blocks support the following:

* `connection_borrow_timeout` - (Optional) The number of seconds for a proxy to wait for a connection to become available in the connection pool. Only applies when the proxy has opened its maximum number of connections and all connections are busy with client sessions.
* `init_query` - (Optional) One or more SQL statements for the proxy to run when opening each new database connection. Typically used with `SET` statements to make sure that each connection has identical settings such as time zone and character set. This setting is empty by default. For multiple statements, use semicolons as the separator. You can also include multiple variables in a single `SET` statement, such as `SET x=1, y=2`.
* `max_connections_percent` - (Optional) The maximum size of the connection pool for each target in a target group. For Aurora MySQL, it is expressed as a percentage of the max_connections setting for the RDS DB instance or Aurora DB cluster used by the target group.
* `max_idle_connections_percent` - (Optional) Controls how actively the proxy closes idle database connections in the connection pool. A high value enables the proxy to leave a high percentage of idle connections open. A low value causes the proxy to close idle client connections and return the underlying database connections to the connection pool. For Aurora MySQL, it is expressed as a percentage of the max_connections setting for the RDS DB instance or Aurora DB cluster used by the target group.
* `session_pinning_filters` - (Optional) Each item in the list represents a class of SQL operations that normally cause all later statements in a session using a proxy to be pinned to the same underlying database connection. Including an item in the list exempts that class of SQL operations from the pinning behavior. Currently, the only allowed value is `EXCLUDE_VARIABLE_SETS`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) for the proxy.
* `arn` - The Amazon Resource Name (ARN) representing the target group.

### Timeouts

`aws_db_proxy_default_target_group` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `update` - (Default `30 minutes`) Used for modifying DB proxy target groups.

## Import

DB proxy default target groups can be imported using the `db_proxy_name`, e.g.

```
$ terraform import aws_db_proxy_default_target_group.example example
```
