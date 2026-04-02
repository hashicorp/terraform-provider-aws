---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_proxy_auth_item"
description: |-
  Terraform resource for managing an AWS RDS (Relational Database) Proxy Auth Association.
---

# Resource: aws_db_proxy_auth_item

Terraform resource for managing an AWS RDS (Relational Database) Proxy Auth Association. For additional information, see the [RDS Proxy Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-proxy.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_db_proxy_auth_item" "example" {
  db_proxy_name = "main-proxy"
  
  secret_arn    = "arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456"
}
```

## Argument Reference

The following arguments are required:

* `db_proxy_name` - (Required) Name of the RDS DB Proxy.
* `secret_arn` - (Required) The Amazon Resource Name (ARN) representing the secret that the proxy uses to authenticate to the RDS DB instance or Aurora DB cluster. These secrets are stored within Amazon Secrets Manager.

The following arguments are optional:

* `auth_scheme` - (Optional) The type of authentication that the proxy uses for connections from the proxy to the underlying database. One of `SECRETS`.
* `client_password_auth_type` - (Optional) The type of authentication the proxy uses for connections from clients. Valid values are `MYSQL_NATIVE_PASSWORD`, `POSTGRES_SCRAM_SHA_256`, `POSTGRES_MD5`, and `SQL_SERVER_AUTHENTICATION`.
* `description` - (Optional) A user-specified description about the authentication used by a proxy to log in as a specific database user.
* `iam_auth` - (Optional) Whether to require or disallow AWS Identity and Access Management (IAM) authentication for connections to the proxy. One of `DISABLED`, `REQUIRED`.
* `username` - (Optional) The name of the database user to which the proxy connects.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the RDS DB Proxy and Secret ARN separated by `/`, `DB-PROXY-NAME/SECRET-ARN`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RDS (Relational Database) Proxy Auth Item using the RDS Proxy name and Secret ARN. For example:

```terraform
import {
  to = aws_db_proxy_auth_item.example
  id = "proxy-name/arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456"
}
```

Using `terraform import`, import RDS (Relational Database) Proxy Auth Association using the RDS Proxy name and Secret ARN. For example:

```console
% terraform import aws_db_proxy_auth_item.example proxy-name/arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456
```
