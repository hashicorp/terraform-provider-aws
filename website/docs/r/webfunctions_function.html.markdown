---
subcategory: "Lambda Web Functions"
layout: "aws"
page_title: "AWS: aws_webfunctions_function"
description: |-
  Manages an AWS Lambda Web Functions function.
---

# Resource: aws_webfunctions_function

Manages an AWS Lambda Web Functions function, including an optional initial revision (code + runtime + execution configuration) and an optional endpoint.

~> **Note:** Lambda Web Functions is available in select regions. Regions active as of July 2026: `us-east-1`, `eu-west-1`. The `us-west-2` rollout is pending.

~> **Note:** Tags are not currently supported: the pre-GA Lambda Web Functions API only accepts tags at creation time and provides no APIs to read or update them, so Terraform cannot manage them without permanent drift. Tag support will be added when the GA API ships tag CRUD operations.

## Example Usage

### Basic Usage

```terraform
resource "aws_webfunctions_function" "example" {
  function_name = "example"

  revision_config {
    build_config {
      runtime_config {
        runtime = "nodejs24.x"
      }

      code_config {
        s3_object {
          bucket = aws_s3_object.example.bucket
          key    = aws_s3_object.example.key
        }
      }
    }

    service_config {
      execution_role_arn = aws_iam_role.example.arn
    }
  }

  endpoint_config {
    endpoint_name = "default"
    endpoint_type = "HomeRegion"
    auth_type     = "ApplicationManaged"
    regions       = ["us-east-1"]
  }
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name of the function. Changing this forces a new resource to be created.

The following arguments are optional:

* `endpoint_config` - (Optional) Configuration block for the function's endpoint. [See below](#endpoint_config-block).
* `region` - (Optional) Region where this resource will be managed. Defaults to the Region set in the provider configuration.
* `revision_config` - (Optional) Configuration block for the function's initial revision. [See below](#revision_config-block).

### `revision_config` Block

* `build_config` - (Required) Code source and runtime. [See below](#build_config-block).
* `description` - (Optional) Description of the revision.
* `kms_key_arn` - (Optional) ARN of the customer managed KMS key used to encrypt the function's code and environment variables.
* `service_config` - (Required) Execution environment configuration. [See below](#service_config-block).

#### `build_config` Block

* `code_config` - (Required) Function code source. [See below](#code_config-block).
* `runtime_config` - (Required) Runtime configuration. [See below](#runtime_config-block).

##### `code_config` Block

* `s3_object` - (Required) S3 location of the function's deployment package. [See below](#s3_object-block).

###### `s3_object` Block

* `bucket` - (Required) S3 bucket of the deployment package.
* `key` - (Required) S3 key of the deployment package.
* `version_id` - (Optional) Object version of the deployment package.

##### `runtime_config` Block

* `runtime` - (Required) Identifier of the function's runtime, e.g. `nodejs24.x`.

#### `service_config` Block

* `environment_variables` - (Optional) Map of environment variables.
* `execution_role_arn` - (Required) ARN of the IAM execution role. The trust principal must be `lambda.amazonaws.com`.
* `max_concurrency_per_environment` - (Optional) Maximum concurrent requests per execution environment (1-128, default 64).
* `telemetry_config` - (Optional) Telemetry configuration. [See below](#telemetry_config-block).
* `timeout_seconds` - (Optional) Request timeout in seconds (3-900, default 30).

##### `telemetry_config` Block

* `logging_config` - (Optional) Logging configuration. [See below](#logging_config-block).

###### `logging_config` Block

* `application_log_level` - (Optional) Application log level. Valid values: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`.
* `log_group` - (Optional) CloudWatch Logs log group for the function's logs.
* `system_log_level` - (Optional) System log level. Valid values: `DEBUG`, `INFO`, `WARN`.

### `endpoint_config` Block

* `auth_type` - (Required) Authentication type. Valid values: `ApplicationManaged`, `IamAuth`. (`AWS_SERVICE_AUTH` was removed in V2.)
* `auto_deployment_mode` - (Optional) Automatic deployment mode. Valid values: `LatestRevision`, `Disabled`.
* `description` - (Optional) Description of the endpoint.
* `endpoint_name` - (Required) Name of the endpoint (typically `default`). Changing this forces a new resource to be created.
* `endpoint_type` - (Required) Endpoint type. Valid values: `HomeRegion`, `MultiRegion`, `PerRegion`. Changing this forces a new resource to be created.
* `regions` - (Optional) List of Regions for the endpoint (maximum 5).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the function in the format `arn:aws:lambda:{region}:{account}:web-function/{name}`.
* `domain_name` - Domain name of the endpoint, when an `endpoint_config` is configured.
* `latest_revision_id` - ID of the most recently published revision. Reference this from `aws_webfunctions_endpoint` `revision_weights` to route traffic.
* `state` - Current state of the function.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_webfunctions_function.example
  identity = {
    function_name = "example"
  }
}

resource "aws_webfunctions_function" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `function_name` (String) Name of the function.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Web Functions Function using the `function_name`. For example:

```terraform
import {
  to = aws_webfunctions_function.example
  id = "example"
}
```

Using `terraform import`, import a Web Functions Function using the `function_name`. For example:

```console
% terraform import aws_webfunctions_function.example example
```
