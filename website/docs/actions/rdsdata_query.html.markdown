---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rdsdata_query"
description: |-
  Executes SQL queries against Aurora Serverless databases using the RDS Data API.
---

# Action: aws_rdsdata_query

~> **Note:** `aws_rdsdata_query` is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Executes SQL queries against Aurora Serverless databases using the RDS Data API. This action allows for imperative SQL execution with support for parameterized queries and transaction management.

For information about the RDS Data API, see the [RDS Data API Developer Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/data-api.html). For specific information about executing statements, see the [ExecuteStatement](https://docs.aws.amazon.com/rdsdataservice/latest/APIReference/API_ExecuteStatement.html) page in the RDS Data Service API Reference.

## Example Usage

### Basic Query

```terraform
resource "aws_rds_cluster" "example" {
  cluster_identifier   = "example-cluster"
  engine               = "aurora-mysql"
  database_name        = "example"
  master_username      = "admin"
  master_password      = "password123"
  enable_http_endpoint = true
  skip_final_snapshot  = true

  serverlessv2_scaling_configuration {
    max_capacity = 1
    min_capacity = 0.5
  }
}

resource "aws_secretsmanager_secret" "example" {
  name = "example-db-credentials"
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({
    username = aws_rds_cluster.example.master_username
    password = aws_rds_cluster.example.master_password
  })
}

action "aws_rdsdata_query" "example" {
  config {
    resource_arn = aws_rds_cluster.example.arn
    secret_arn   = aws_secretsmanager_secret_version.example.arn
    sql          = "SELECT COUNT(*) FROM users"
  }
}

resource "terraform_data" "example" {
  input = "trigger-query"

  lifecycle {
    action_trigger {
      events  = [before_create]
      actions = [action.aws_rdsdata_query.example]
    }
  }
}
```

### Parameterized Query

```terraform
action "aws_rdsdata_query" "user_lookup" {
  config {
    resource_arn = aws_rds_cluster.example.arn
    secret_arn   = aws_secretsmanager_secret_version.example.arn
    database     = "example"
    sql          = "SELECT * FROM users WHERE status = :status AND created_date > :date"

    parameters {
      name  = "status"
      value = "active"
    }

    parameters {
      name  = "date"
      value = "2024-01-01"
    }
  }
}
```

### Data Modification Query

```terraform
action "aws_rdsdata_query" "update_status" {
  config {
    resource_arn = aws_rds_cluster.example.arn
    secret_arn   = aws_secretsmanager_secret_version.example.arn
    database     = "example"
    sql          = "UPDATE users SET last_login = NOW() WHERE user_id = :user_id"

    parameters {
      name  = "user_id"
      value = "12345"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` - (Required) The Amazon Resource Name (ARN) of the Aurora Serverless DB cluster.
* `secret_arn` - (Required) The ARN of the secret that enables access to the DB cluster. The secret must contain the database credentials.
* `sql` - (Required) The SQL statement to execute.

The following arguments are optional:

* `database` - (Optional) The name of the database to execute the SQL statement against.
* `parameters` - (Optional) Parameters for the SQL statement. See [Parameters](#parameters) below.

### Parameters

The `parameters` block supports the following:

* `name` - (Required) The name of the parameter.
* `value` - (Required) The value of the parameter.
