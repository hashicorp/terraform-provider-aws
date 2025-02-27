---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_proxy"
description: |-
  Provides an RDS DB proxy resource.
---

# Resource: aws_db_proxy

Provides an RDS DB proxy resource. For additional information, see the [RDS User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-proxy.html).

~> **Note:** Not all Availability Zones (AZs) support DB proxies. Specifying `vpc_subnet_ids` for AZs that do not support proxies will not trigger an error as long as at least one `vpc_subnet_id` is valid. However, this will cause Terraform to continuously detect differences between the configuration and the actual infrastructure. Refer to the [Unsupported Availability Zones](#unsupported-availability-zones) section below for potential workarounds.

## Example Usage

### Basic Usage

```terraform
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
```

### Unsupported Availability Zones

Terraform may report constant differences if you use `vpc_subnet_ids` that correspond to Availability Zones (AZs) that do not support a DB proxy. While this typically does not result in an error, AWS only returns `vpc_subnet_ids` for AZs that support DB proxies. As a result, Terraform detects a mismatch between your configuration and the actual infrastructure, leading it to report that changes are required. Below are some ways to avoid this issue.

One solution is to exclude AZs that do not support DB proxies by using the [`aws_availability_zones` data source](/docs/providers/aws/d/availability_zones.html). The example below demonstrates how to configure this for the `us-east-1` region, excluding the `use1-az3` AZ. (Keep in mind that AZ names can vary between accounts, while AZ IDs remain consistent.) If the `us-east-1` region has six AZs in total and you aim to configure the maximum number of subnets, you would exclude one AZ and configure five subnets:

```terraform
data "aws_availability_zones" "available" {
  exclude_zone_ids = ["use1-az3"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example" {
  count             = 5
  cidr_block        = cidrsubnet(aws_vpc.example.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.example.id
}

resource "aws_db_proxy" "example" {
  name           = "example"
  vpc_subnet_ids = [aws_subnet.example.id]

  # additional configuration...
}
```

Another approach is to use the [`lifecycle` `ignore_changes`](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes) meta-argument. With this method, Terraform will stop detecting differences for the `vpc_subnet_ids` argument. However, note that this approach disables Terraform's ability to track and manage updates to `vpc_subnet_ids`, so use it carefully to avoid unintended drift between your configuration and the actual infrastructure.

```terraform
resource "aws_db_proxy" "example" {
  name           = "example"
  vpc_subnet_ids = [aws_subnet.example.id]

  lifecycle {
    ignore_changes = [vpc_subnet_ids]
  }

  # additional configuration...
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The identifier for the proxy. This name must be unique for all proxies owned by your AWS account in the specified AWS Region. An identifier must begin with a letter and must contain only ASCII letters, digits, and hyphens; it can't end with a hyphen or contain two consecutive hyphens.
* `auth` - (Required) Configuration block(s) with authorization mechanisms to connect to the associated instances or clusters. Described below.
* `debug_logging` - (Optional) Whether the proxy includes detailed information about SQL statements in its logs. This information helps you to debug issues involving SQL behavior or the performance and scalability of the proxy connections. The debug information includes the text of SQL statements that you submit through the proxy. Thus, only enable this setting when needed for debugging, and only when you have security measures in place to safeguard any sensitive information that appears in the logs.
* `engine_family` - (Required, Forces new resource) The kinds of databases that the proxy can connect to. This value determines which database network protocol the proxy recognizes when it interprets network traffic to and from the database. For Aurora MySQL, RDS for MariaDB, and RDS for MySQL databases, specify `MYSQL`. For Aurora PostgreSQL and RDS for PostgreSQL databases, specify `POSTGRESQL`. For RDS for Microsoft SQL Server, specify `SQLSERVER`. Valid values are `MYSQL`, `POSTGRESQL`, and `SQLSERVER`.
* `idle_client_timeout` - (Optional) The number of seconds that a connection to the proxy can be inactive before the proxy disconnects it. You can set this value higher or lower than the connection timeout limit for the associated database.
* `require_tls` - (Optional) A Boolean parameter that specifies whether Transport Layer Security (TLS) encryption is required for connections to the proxy. By enabling this setting, you can enforce encrypted TLS connections to the proxy.
* `role_arn` - (Required) The Amazon Resource Name (ARN) of the IAM role that the proxy uses to access secrets in AWS Secrets Manager.
* `vpc_security_group_ids` - (Optional) One or more VPC security group IDs to associate with the new proxy.
* `vpc_subnet_ids` - (Required) One or more VPC subnet IDs to associate with the new proxy.
* `tags` - (Optional) A mapping of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

`auth` blocks support the following:

* `auth_scheme` - (Optional) The type of authentication that the proxy uses for connections from the proxy to the underlying database. One of `SECRETS`.
* `client_password_auth_type` - (Optional) The type of authentication the proxy uses for connections from clients. Valid values are `MYSQL_CACHING_SHA2_PASSWORD`, `MYSQL_NATIVE_PASSWORD`, `POSTGRES_SCRAM_SHA_256`, `POSTGRES_MD5`, and `SQL_SERVER_AUTHENTICATION`.
* `description` - (Optional) A user-specified description about the authentication used by a proxy to log in as a specific database user.
* `iam_auth` - (Optional) Whether to require or disallow AWS Identity and Access Management (IAM) authentication for connections to the proxy. One of `DISABLED`, `REQUIRED`.
* `secret_arn` - (Optional) The Amazon Resource Name (ARN) representing the secret that the proxy uses to authenticate to the RDS DB instance or Aurora DB cluster. These secrets are stored within Amazon Secrets Manager.
* `username` - (Optional) The name of the database user to which the proxy connects.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) for the proxy.
* `arn` - The Amazon Resource Name (ARN) for the proxy.
* `endpoint` - The endpoint that you can use to connect to the proxy. You include the endpoint value in the connection string for a database client application.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `update` - (Default `30m`)
- `delete` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DB proxies using the `name`. For example:

```terraform
import {
  to = aws_db_proxy.example
  id = "example"
}
```

Using `terraform import`, import DB proxies using the `name`. For example:

```console
% terraform import aws_db_proxy.example example
```
