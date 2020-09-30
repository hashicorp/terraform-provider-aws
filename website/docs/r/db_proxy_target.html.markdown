---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_db_proxy_target"
description: |-
  Provides an RDS DB proxy target resource.
---

# Resource: aws_db_proxy_target

Provides an RDS DB proxy target resource.

## Example Usage

```hcl
resource "aws_db_proxy" "example" {
  name                   = "example"
  debug_logging          = false
  engine_family          = "MYSQL"
  idle_client_timeout    = 1800
  require_tls            = true
  role_arn               = aws_iam_role.example.arn
  vpc_security_group_ids = [aws_security_group.example.id]
  vpc_subnet_ids         = [aws_subnet.example.id]

  auth {
    auth_scheme = "SECRETS"
    description = "example"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.example.arn
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

resource "aws_db_proxy_target" "example" {
  db_proxy_name          = aws_db_proxy.example.db_proxy_name
  target_group_name      = aws_db_proxy_default_target_group.example.name
  db_instance_identifier = aws_db_instance.example.id
  # db_cluster_identifier  = ""
}
```

## Argument Reference

The following arguments are supported:

* `db_proxy_name` - (Required, Forces new resource) The name of the DB proxy.
* `target_group_name` - (Required, Forces new resource) The name of the target group.
* `db_instance_identifier` - (Optional, Forces new resource) DB instance identifier.
* `db_cluster_identifier` - (Optional, Forces new resource) DB cluster identifier.

**NOTE:** Either `db_instance_identifier` or `db_cluster_identifier` should be specified and both should not be specified together

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) for the proxy.
* `endpoint` - The writer endpoint for the RDS DB instance or Aurora DB cluster.
* `port` - The port that the RDS Proxy uses to connect to the target RDS DB instance or Aurora DB cluster.
* `rds_resource_id` - The identifier representing the target. It can be the instance identifier for an RDS DB instance, or the cluster identifier for an Aurora DB cluster.
* `target_arn` - The Amazon Resource Name (ARN) for the RDS DB instance or Aurora DB cluster.
* `tracked_cluster_id` - The DB cluster identifier when the target represents an Aurora DB cluster. This field is blank when the target represents an RDS DB instance.
* `type` - Specifies the kind of database, such as an RDS DB instance or an Aurora DB cluster, that the target represents.

### Timeouts

`aws_db_proxy_target` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `30 minutes`) Used for registering DB proxy targets.
- `delete` - (Default `30 minutes`) Used for deregistering DB proxy targets.

## Import

DB proxy targets can be imported using the `db_proxy_name`/`target_group_name`/`rds_resource_id`, e.g.

```
$ terraform import aws_db_proxy_target.example example/default/example-db-identifier
```
