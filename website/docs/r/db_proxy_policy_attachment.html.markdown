---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_proxy_policy_attachment"
description: |-
  Manages an RDS DB proxy IAM policy attachment.
---

# Resource: aws_db_proxy_policy_attachment

Manages an inline IAM policy attachment on the execution role of an RDS DB Proxy. This allows you to grant the proxy's IAM role additional permissions (for example, access to Secrets Manager secrets) independently from the `aws_db_proxy` resource.

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
    secret_arn  = aws_secretsmanager_secret.example.arn
  }
}

resource "aws_db_proxy_policy_attachment" "example" {
  db_proxy_name = aws_db_proxy.example.name
  policy_name   = "rds-proxy-access"

  policy_document = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
        ]
        Resource = [
          aws_secretsmanager_secret.example.arn,
        ]
      },
    ]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `db_proxy_name` - (Required, Forces new resource) The name of the DB proxy whose execution role will receive the policy.
* `policy_name` - (Required, Forces new resource) The name of the inline IAM policy to attach to the proxy's execution role.
* `policy_document` - (Required) The JSON IAM policy document to attach.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The `db_proxy_name` and `policy_name` separated by a forward slash (`/`).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `update` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RDS DB Proxy Policy Attachments using the `db_proxy_name` and `policy_name` separated by a forward slash (`/`). For example:

```terraform
import {
  to = aws_db_proxy_policy_attachment.example
  id = "example-proxy/rds-proxy-access"
}
```

**Using `terraform import` to import** RDS DB Proxy Policy Attachments using the `db_proxy_name` and `policy_name` separated by a forward slash (`/`). For example:

```console
% terraform import aws_db_proxy_policy_attachment.example example-proxy/rds-proxy-access
```
