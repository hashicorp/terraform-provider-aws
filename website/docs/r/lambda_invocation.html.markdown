---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_invocation"
description: |-
  Manages an AWS Lambda Function invocation.
---

# Resource: aws_lambda_invocation

Manages an AWS Lambda Function invocation. Use this resource to invoke a Lambda function with the [RequestResponse](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_RequestSyntax) invocation type.

~> **Note:** By default this resource _only_ invokes the function when the arguments call for a create or replace. After an initial invocation on _apply_, if the arguments do not change, a subsequent _apply_ does not invoke the function again. To dynamically invoke the function, see the `triggers` example below. To always invoke a function on each _apply_, see the [`aws_lambda_invocation` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/lambda_invocation). To invoke the Lambda function when the Terraform resource is updated and deleted, see the [CRUD Lifecycle Management](#crud-lifecycle-management) example below.

~> **Note:** If you get a `KMSAccessDeniedException: Lambda was unable to decrypt the environment variables because KMS access was denied` error when invoking a Lambda function with environment variables, the IAM role associated with the function may have been deleted and recreated after the function was created. You can fix the problem two ways: 1) updating the function's role to another role and then updating it back again to the recreated role, or 2) by using Terraform to `taint` the function and `apply` your configuration again to recreate the function. (When you create a function, Lambda grants permissions on the KMS key to the function's IAM role. If the IAM role is recreated, the grant is no longer valid. Changing the function's role or recreating the function causes Lambda to update the grant.)

## Example Usage

### Basic Invocation

```terraform
# Lambda function to invoke
resource "aws_lambda_function" "example" {
  filename      = "function.zip"
  function_name = "data_processor"
  role          = aws_iam_role.lambda_role.arn
  handler       = "index.handler"
  runtime       = "python3.12"
}

# Invoke the function once during resource creation
resource "aws_lambda_invocation" "example" {
  function_name = aws_lambda_function.example.function_name

  input = jsonencode({
    operation = "initialize"
    config = {
      environment = "production"
      debug       = false
    }
  })
}

# Use the result in other resources
output "initialization_result" {
  value = jsondecode(aws_lambda_invocation.example.result)["status"]
}
```

### Dynamic Invocation with Triggers

```terraform
resource "aws_lambda_invocation" "example" {
  function_name = aws_lambda_function.example.function_name

  # Re-invoke when function environment changes
  triggers = {
    function_version = aws_lambda_function.example.version
    config_hash = sha256(jsonencode({
      environment = var.environment
      timestamp   = timestamp()
    }))
  }

  input = jsonencode({
    operation   = "process_data"
    environment = var.environment
    batch_id    = random_uuid.batch_id.result
  })
}
```

### CRUD Lifecycle Management

```terraform
resource "aws_lambda_invocation" "example" {
  function_name = aws_lambda_function.example.function_name

  input = jsonencode({
    resource_name = "database_setup"
    database_url  = aws_db_instance.example.endpoint
    credentials = {
      username = var.db_username
      password = var.db_password
    }
  })

  lifecycle_scope = "CRUD"
}
```

~> **Note:** `lifecycle_scope = "CRUD"` will inject a key `tf` in the input event to pass lifecycle information! This allows the Lambda function to handle different lifecycle transitions uniquely. If you need to use a key `tf` in your own input JSON, the default key name can be overridden with the `terraform_key` argument.

The lifecycle key gets added with subkeys:

* `action` - Action Terraform performs on the resource. Values are `create`, `update`, or `delete`.
* `prev_input` - Input JSON payload from the previous invocation. This can be used to handle update and delete events.

When the resource from the CRUD example above is created, the Lambda will receive the following JSON payload:

```json
{
  "resource_name": "database_setup",
  "database_url": "mydb.cluster-xyz.us-west-2.rds.amazonaws.com:5432",
  "credentials": {
    "username": "admin",
    "password": "secret123"
  },
  "tf": {
    "action": "create",
    "prev_input": null
  }
}
```

If the `database_url` changes, the Lambda will be invoked again with:

```json
{
  "resource_name": "database_setup",
  "database_url": "mydb-new.cluster-abc.us-west-2.rds.amazonaws.com:5432",
  "credentials": {
    "username": "admin",
    "password": "secret123"
  },
  "tf": {
    "action": "update",
    "prev_input": {
      "resource_name": "database_setup",
      "database_url": "mydb.cluster-xyz.us-west-2.rds.amazonaws.com:5432",
      "credentials": {
        "username": "admin",
        "password": "secret123"
      }
    }
  }
}
```

When the invocation resource is removed, the final invocation will have:

```json
{
  "resource_name": "database_setup",
  "database_url": "mydb-new.cluster-abc.us-west-2.rds.amazonaws.com:5432",
  "credentials": {
    "username": "admin",
    "password": "secret123"
  },
  "tf": {
    "action": "delete",
    "prev_input": {
      "resource_name": "database_setup",
      "database_url": "mydb-new.cluster-abc.us-west-2.rds.amazonaws.com:5432",
      "credentials": {
        "username": "admin",
        "password": "secret123"
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name of the Lambda function.
* `input` - (Required) JSON payload to the Lambda function.

The following arguments are optional:

* `lifecycle_scope` - (Optional) Lifecycle scope of the resource to manage. Valid values are `CREATE_ONLY` and `CRUD`. Defaults to `CREATE_ONLY`. `CREATE_ONLY` will invoke the function only on creation or replacement. `CRUD` will invoke the function on each lifecycle event, and augment the input JSON payload with additional lifecycle information.
* `qualifier` - (Optional) Qualifier (i.e., version) of the Lambda function. Defaults to `$LATEST`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `terraform_key` - (Optional) JSON key used to store lifecycle information in the input JSON payload. Defaults to `tf`. This additional key is only included when `lifecycle_scope` is set to `CRUD`.
* `triggers` - (Optional) Map of arbitrary keys and values that, when changed, will trigger a re-invocation. To force a re-invocation without changing these keys/values, use the [`terraform taint` command](https://developer.hashicorp.com/terraform/cli/commands/taint).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `result` - String result of the Lambda function invocation.
