---
subcategory: "Lambda Web"
layout: "aws"
page_title: "AWS: aws_lambdaweb_function"
description: |-
  Manages an AWS Lambda Web function.
---

# Resource: aws_lambdaweb_function

Manages an AWS Lambda Web function, including an optional initial revision (code + runtime + execution configuration) and an optional endpoint.

~> **Note:** Lambda Web is available in select regions. As of August 2026 the API is active in 17 commercial regions, including `us-east-1` and `eu-west-1`. In regions where the service is not yet deployed, API calls fail with `AccessDeniedException`.

## Example Usage

### Basic Usage

The function's deployment package must live in an S3 bucket that meets two hard service requirements: the bucket **must have versioning enabled**, and its bucket policy **must grant `s3:GetObject` and `s3:GetObjectVersion` to the `lambda.amazonaws.com` service principal**. Function creation fails with a `ValidationException` if either is missing.

```terraform
resource "aws_s3_bucket" "code" {
  bucket = "example-lambdaweb-code"
}

# REQUIRED: the service only accepts versioned buckets (REFERENCE storage mode).
resource "aws_s3_bucket_versioning" "code" {
  bucket = aws_s3_bucket.code.id
  versioning_configuration {
    status = "Enabled"
  }
}

# REQUIRED: the Lambda service principal must be able to read the package.
resource "aws_s3_bucket_policy" "code" {
  bucket = aws_s3_bucket.code.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = ["s3:GetObject", "s3:GetObjectVersion"]
      Resource  = "${aws_s3_bucket.code.arn}/*"
    }]
  })
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket_versioning.code.bucket
  key    = "function.zip"
  source = "function.zip"
}

# REQUIRED: the execution role must be assumable by lambda.amazonaws.com.
resource "aws_iam_role" "example" {
  name = "example-lambdaweb"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambdaweb_function" "example" {
  depends_on = [aws_s3_bucket_policy.code, aws_iam_role_policy_attachment.example]

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

### Updating code

Revisions are immutable: any change to `revision_config` (new package, runtime, environment variables, timeouts, ...) rolls a new revision and waits until it is `Active`. `latest_revision_id` always tracks the newest revision.

~> **Note:** Traffic only follows new revisions automatically on endpoints with `auto_deployment_mode = "LatestRevision"`. On an endpoint with `auto_deployment_mode = "Disabled"` (which `MultiRegion` and `PerRegion` require), publishing a revision does **not** shift traffic, and the `endpoint_config` block cannot shift it either: traffic weights are only settable through the `UpdateWebFunctionEndpoint` API, which has no equivalent in `CreateWebFunction`'s `endpointConfig`. A function whose inline endpoint is `MultiRegion` or `PerRegion` therefore keeps serving the revision the endpoint was created with, even though the apply succeeds and the following plan is empty. To deploy new revisions to such an endpoint, keep the inline `endpoint_config` on `LatestRevision` and manage the traffic-shifted endpoint as a separate [`aws_lambdaweb_endpoint`](lambdaweb_endpoint.html.markdown) resource, whose `revision_weights` can reference this function's `latest_revision_id`. The provider emits a plan warning whenever an apply would publish a revision that the inline endpoint will not serve.

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name of the function, up to 64 characters. Changing this forces a new resource to be created.

The following arguments are optional:

* `endpoint_config` - (Optional) Configuration block for the function's endpoint. [See below](#endpoint_config-block).
* `region` - (Optional) Region where this resource will be managed. Defaults to the Region set in the provider configuration.
* `revision_config` - (Optional) Configuration block for the function's initial revision. [See below](#revision_config-block).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `revision_config` Block

* `build_config` - (Required) Code source and runtime. [See below](#build_config-block).
* `description` - (Optional) Description of the revision.
* `kms_key_arn` - (Optional) ARN of the customer managed KMS key used to encrypt the function's code and environment variables.

~> **Note:** A revision encrypted with a single-Region customer managed key cannot be deployed to the other Regions of a `MultiRegion` or `PerRegion` endpoint: those Regions report `Failed to deploy all the specified revision(s)` while the home Region goes `Active`, so the endpoint ends up `Failed` even though the function and the revision report `Active`. Either keep `kms_key_arn` with a `HomeRegion` endpoint, or use an [AWS KMS multi-Region key](https://docs.aws.amazon.com/kms/latest/developerguide/multi-region-keys-overview.html) replicated into every Region of the endpoint, which is verified to work.
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

* `environment_variables` - (Optional) Map of environment variables, up to 32 KB across all names and values.
* `execution_role_arn` - (Required) ARN of the IAM execution role. The trust principal must be `lambda.amazonaws.com`.
* `max_concurrency_per_environment` - (Optional) Maximum concurrent requests per execution environment (1-128, default 64).
* `telemetry_config` - (Optional) Telemetry configuration. The service assigns default telemetry (a log group and `INFO` log levels) to every revision; Terraform tracks this block as written, so leaving it out does not produce drift against those defaults. [See below](#telemetry_config-block).
* `timeout_seconds` - (Optional) Request timeout in seconds (3-900, default 30).

~> **Note:** The runtime starts `index.js` and expects it to listen on `0.0.0.0:3000`. To serve an entry point with another name, set `AWS_LAMBDA_ENTRYPOINT` in `environment_variables` to that file. Without it a package whose server lives elsewhere starts nothing, and the endpoint answers an opaque `HTTP 500` with no log line explaining it.

##### `telemetry_config` Block

* `logging_config` - (Optional) Logging configuration. [See below](#logging_config-block).

###### `logging_config` Block

* `application_log_level` - (Optional) Application log level. Valid values: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`.
* `log_group` - (Optional) CloudWatch Logs log group for the function's logs.
* `system_log_level` - (Optional) System log level. Valid values: `DEBUG`, `INFO`, `WARN`.

### `endpoint_config` Block

* `auth_type` - (Required) Authentication type. Valid values: `ApplicationManaged`, `IamAuth`. The authentication type is chosen at creation and cannot be edited afterward, so changing it forces a new resource to be created.
* `auto_deployment_mode` - (Optional) Automatic deployment mode. Valid values: `LatestRevision`, `Disabled`. Defaults to `LatestRevision` for `HomeRegion` endpoints: the provider sends it explicitly, because the API would otherwise create the endpoint `Disabled` and pin it to the initial revision. `MultiRegion` and `PerRegion` endpoints require `Disabled`, which must be set explicitly.
* `description` - (Optional) Description of the endpoint.
* `endpoint_name` - (Required) Name of the endpoint (typically `default`), up to 64 characters. Changing this forces a new resource to be created.
* `endpoint_type` - (Required) Endpoint type. Valid values: `HomeRegion`, `MultiRegion`, `PerRegion`. Changing this forces a new resource to be created.
* `regions` - (Optional) List of Regions for the endpoint (maximum 5). `MultiRegion` and `PerRegion` endpoints require at least 2 distinct regions, or none at all: the home region is added automatically.
* `scaling_config` - (Optional) Scaling limits for the endpoint, an object (assigned with `=`, not a block) with a single `max_environments` attribute: the maximum number of concurrent execution environments, minimum 2. When unset, the service applies account-level defaults and reports no value.
* `throttle_config` - (Optional) Request throttling for the endpoint, an object (assigned with `=`, not a block) with a single `rate_limit` attribute: the maximum request rate in requests per second, quantized (`0`, `100`-`1000` in steps of 100, `2000`-`10000` in steps of 1000). When unset, the service applies account-level defaults and reports no value.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the function in the format `arn:aws:lambda:{region}:{account}:web-function/{name}`.
* `domain_name` - Domain name of the endpoint, when an `endpoint_config` is configured.
* `latest_revision_id` - ID of the most recently published revision. Reference this from `aws_lambdaweb_endpoint` `revision_weights` to route traffic.
* `regional_domain_names` - Map of Region to domain name. Populated for `PerRegion` endpoints, which serve an independent domain per Region; `MultiRegion` endpoints route through the single global `domain_name`.
* `state` - Current state of the function.
* `state_reason` - Reason for the current state, useful when the function is `Pending` or `Failed`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

Import fully populates the resource state from the API, including `revision_config` and `endpoint_config`, with one exception: `telemetry_config`, which the service fills with defaults on every revision, is left unset. The first plan after import reports no changes when the configuration matches the deployed function and omits `telemetry_config` (add the block after importing if the revision was created with explicit telemetry settings).

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_lambdaweb_function.example
  identity = {
    function_name = "example"
  }
}

resource "aws_lambdaweb_function" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `function_name` (String) Name of the function.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Lambda Web Function using the `function_name`. For example:

```terraform
import {
  to = aws_lambdaweb_function.example
  id = "example"
}
```

Using `terraform import`, import a Lambda Web Function using the `function_name`. For example:

```console
% terraform import aws_lambdaweb_function.example example
```
