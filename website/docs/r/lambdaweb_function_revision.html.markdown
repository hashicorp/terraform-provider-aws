---
subcategory: "Lambda Web"
layout: "aws"
page_title: "AWS: aws_lambdaweb_function_revision"
description: |-
  Manages an AWS Lambda Web function revision.
---

# Resource: aws_lambdaweb_function_revision

Manages an AWS Lambda Web function revision. A revision represents a specific version of the function code and configuration and is immutable (updating any attribute recreates the resource).

## Example Usage

### Basic Usage

```terraform
resource "aws_lambdaweb_function_revision" "example" {
  function_name = aws_lambdaweb_function.example.function_name

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

resource "aws_lambdaweb_function" "example" {
  function_name = "example-function"
}
```

## Argument Reference

The following arguments are required:

* `build_config` - (Required) Build configuration for the revision. See [`build_config` Block](#service_config-block) below.
* `function_name` - (Required) Web function name.
* `service_config` - (Required) Service configuration for the revision. See [`service_config` Block](#service_config-block) below.

The following arguments are optional:

* `description` - (Optional) Description of the revision.
* `kms_key_arn` - (Optional) ARN of the KMS key used to encrypt the revision's code and environment variables.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

#### `build_config` Block

* `code_config` - (Required) Code configuration. See [`code_config` Block](#service_config-block) below.
* `runtime_config` - (Required) Runtime configuration for the revision. See [`runtime_config` Block](#service_config-block) below.

##### `code_config` Block

* `s3_object` - (Required) Amazon S3 location of the deployment artifact. See [`s3_object` Block](#service_config-block) below.

###### `s3_object` Block

* `bucket` - (Required) Name of the Amazon S3 bucket.
* `key` - (Required) Amazon S3 object key.
* `version_id` - (Optional) Version ID of the Amazon S3 object.

##### `runtime_config` Block

* `runtime` - (Required) Runtime identifier for the web function, e.g. `nodejs24.x`.

#### `service_config` Block

* `egress_network_connector_arn` - (Optional) ARN of the Lambda network connector that gives the web function access to resources in a VPC. Specify the fully qualified ARN, including the connector version, which pins the revision to that version.
* `environment_variables` - (Optional) Map of environment variable key-value pairs available to the web function at runtime.
* `execution_role_arn` - (Required) ARN of the IAM role that the web function assumes when it runs. The trust principal must be `lambda.amazonaws.com`.
* `max_concurrency_per_environment` - (Optional) Maximum number of concurrent requests handled per execution environment. The default is `64`.
* `telemetry_config` - (Optional) Telemetry configuration for the web function. See [`telemetry_config` Block](#service_config-block) below.
* `timeout_seconds` - (Optional) Amount of time (in seconds) that Lambda allows the web function to run before stopping it. The default is `30`.

##### `telemetry_config` Block

* `logging_config` - (Optional) Logging configuration for the web function. See [`logging_config` Block](#service_config-block) below.

###### `logging_config` Block

* `application_log_level` - (Optional) Log level for application logs emitted by the web function. Valid values: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`. The default is `INFO`.
* `log_group` - (Optional) Name of the Amazon CloudWatch Logs log group the web function sends logs to. The default is `/aws/lambda/web/{function_name}`.
* `system_log_level` - (Optional) Log level for system logs emitted by the Lambda runtime. Valid values: `DEBUG`, `INFO`, `WARN`. The default is INFO.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the revision.
* `revision_id` - ID of the revision.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_lambdaweb_function_revision.example

  identity = {
    function_name = "example-function"
    revision_id   = "rev-sglsh"
  }
}

resource "aws_lambdaweb_function_revision" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `function_name` (String) Web function name.
* `revision_id` (String) ID of the revision.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a revision using `function_name` and `revision_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_lambdaweb_function_revision.example
  id = "example-function,rev-sglsh"
}
```

Using `terraform import`, import a revision using `function_name` and `revision_id` separated by a comma (`,`). For example:

```console
% terraform import aws_lambdaweb_function_revision.example example-function,rev-sglsh
```
