---
subcategory: "Web Functions"
layout: "aws"
page_title: "AWS: aws_webfunctions_function"
description: |-
  Manages an AWS Lambda Web Functions function.
---

# Resource: aws_webfunctions_function

Manages an AWS Lambda Web Functions function, including an optional initial revision (code + runtime + execution configuration) and an optional endpoint.

~> **Note:** Lambda Web Functions (V2, `aws lambda-web` CLI) reached general availability in select regions. The V1 API (`aws lite` / `aws lambda-lite`, `LiteFunction` operations) was revoked in July 2026. This resource targets the V2 contract. Regions with V2 active as of July 2026: `us-east-1`, `eu-west-1`. The `us-west-2` rollout is pending.

## Example Usage

### Basic Usage

```terraform
resource "aws_webfunctions_function" "example" {
  function_name = "example"

  revision_config {
    build_config {
      runtime = "nodejs24.x"

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

* `revision_config` - (Optional) Configuration block for the function's initial revision. [See below](#revision_config).
* `endpoint_config` - (Optional) Configuration block for the function's endpoint. [See below](#endpoint_config).
* `region` - (Optional) Region where this resource will be managed. Defaults to the Region set in the provider configuration.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

### revision_config

* `build_config` - (Required) Code source and runtime. [See below](#build_config).
* `service_config` - (Required) Execution environment configuration. [See below](#service_config).
* `description` - (Optional) Description of the revision.
* `kms_key_arn` - (Optional) ARN of the customer managed KMS key used to encrypt the function's code and environment variables.

#### build_config

* `runtime` - (Required) Identifier of the function's runtime, e.g. `nodejs24.x`.
* `code_config` - (Required) Function code source. Exactly one of `s3_object` or `zip_file`. [See below](#code_config).

##### code_config

* `s3_object` - (Optional) S3 location of the function's deployment package. [See below](#s3_object).
* `zip_file` - (Optional) Base64-encoded inline deployment package.

###### s3_object

* `bucket` - (Required) S3 bucket of the deployment package.
* `key` - (Required) S3 key of the deployment package.
* `version_id` - (Optional) Object version of the deployment package.

#### service_config

* `execution_role_arn` - (Required) ARN of the IAM execution role. The trust principal must be `lambda.amazonaws.com`.
* `timeout_seconds` - (Optional) Request timeout in seconds (3-900, default 30).
* `max_concurrency_per_environment` - (Optional) Maximum concurrent requests per execution environment (1-128, default 64).
* `environment_variables` - (Optional) Map of environment variables.

### endpoint_config

* `endpoint_name` - (Required) Name of the endpoint (typically `default`). Changing this forces a new resource to be created.
* `endpoint_type` - (Required) Endpoint type. Valid values: `HomeRegion`, `MultiRegion`, `PerRegion`. Changing this forces a new resource to be created.
* `auth_type` - (Required) Authentication type. Valid values: `ApplicationManaged`, `IamAuth`. (`AWS_SERVICE_AUTH` was removed in V2.)
* `auto_deployment_mode` - (Optional) Automatic deployment mode. Valid values: `LatestRevision`, `Disabled`.
* `description` - (Optional) Description of the endpoint.
* `regions` - (Optional) List of Regions for the endpoint (maximum 5).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the function in the format `arn:aws:lambda:{region}:{account}:web-function/{name}`.
* `domain_name` - Domain name of the endpoint, when an `endpoint_config` is configured.
* `latest_revision_id` - ID of the most recently published revision. Reference this from `aws_webfunctions_endpoint` `revision_weights` to route traffic.
* `state` - Current state of the function.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider `default_tags` configuration block.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

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
