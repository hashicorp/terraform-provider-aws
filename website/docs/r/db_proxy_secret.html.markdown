---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_proxy_secret"
description: |-
  Manages an RDS DB proxy secret association.
---

# Resource: aws_db_proxy_secret

Manages an association between an RDS DB proxy and a Secrets Manager secret. This allows you to add additional secrets to an existing proxy's authentication configuration independently from the `aws_db_proxy` resource.

## Example Usage

```terraform
resource "aws_db_proxy" "example" {
  name                   = "example"
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.example.arn
  vpc_subnet_ids         = aws_subnet.example[*].id
  require_tls            = true

  auth {
    auth_scheme = "SECRETS"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.initial.arn
  }
}

resource "aws_secretsmanager_secret" "additional" {
  name = "example-additional"
}

resource "aws_secretsmanager_secret_version" "additional" {
  secret_id     = aws_secretsmanager_secret.additional.id
  secret_string = jsonencode({
    username = "additional_user"
    password = "example_password"
  })
}

resource "aws_db_proxy_secret" "example" {
  db_proxy_name = aws_db_proxy.example.name
  secret_arn    = aws_secretsmanager_secret.additional.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `db_proxy_name` - (Required, Forces new resource) The name of the DB proxy to associate the secret with.
* `secret_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the Secrets Manager secret that the proxy uses to authenticate to the database.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The `db_proxy_name` and `secret_arn` separated by a forward slash (`/`).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RDS DB Proxy Secrets using the `db_proxy_name` and `secret_arn` separated by a forward slash (`/`). For example:

```terraform
import {
  to = aws_db_proxy_secret.example
  id = "example-proxy/arn:aws:secretsmanager:us-east-1:123456789012:secret:example"
}
```

**Using `terraform import` to import** RDS DB Proxy Secrets using the `db_proxy_name` and `secret_arn` separated by a forward slash (`/`). For example:

```console
% terraform import aws_db_proxy_secret.example example-proxy/arn:aws:secretsmanager:us-east-1:123456789012:secret:example
```
