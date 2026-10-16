---
subcategory: "Lambda MicroVMs"
layout: "aws"
page_title: "AWS: aws_lambdamicrovms_image"
description: |-
  Manages an AWS Lambda MicroVMs Image.
---

# Resource: aws_lambdamicrovms_image

Manages an AWS Lambda MicroVMs Image. Use this resource to define the base image, application code, and runtime configuration from which MicroVMs are launched.

## Example Usage

### Basic Usage

```terraform
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_iam_role" "example" {
  name = "example"

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

resource "aws_iam_role_policy" "example" {
  name = "example"
  role = aws_iam_role.example.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = ["s3:GetObject"]
      Effect   = "Allow"
      Resource = "${aws_s3_bucket.example.arn}/*"
    }]
  })
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket.example.bucket
  key    = "code.zip"
  source = "code.zip"
}

resource "aws_lambdamicrovms_image" "example" {
  name           = "example"
  base_image_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.region}:aws:microvm-image:al2023-1"
  build_role_arn = aws_iam_role.example.arn

  code_artifact {
    uri = "s3://${aws_s3_bucket.example.bucket}/${aws_s3_object.example.key}"
  }
}
```

### Lifecycle Hooks and Logging

```terraform
resource "aws_lambdamicrovms_image" "example" {
  # ... other configuration ...

  hooks {
    port = 9000

    microvm_hooks {
      run                          = "ENABLED"
      run_timeout_in_seconds       = 60
      terminate                    = "ENABLED"
      terminate_timeout_in_seconds = 60
    }

    microvm_image_hooks {
      ready = "ENABLED"
    }
  }

  logging {
    cloudwatch {
      log_group = "/aws/lambda/microvms/example"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `base_image_arn` - (Required) ARN of the base MicroVM image. AWS-managed base images use ARNs of the form `arn:aws:lambda:<region>:aws:microvm-image:al2023-1`.
* `build_role_arn` - (Required) ARN of the IAM role used to build the image. The role must be assumable by `lambda.amazonaws.com` and have access to the code artifact.
* `code_artifact` - (Required) Code artifact containing the application code and metadata for the image. [See below](#code_artifact-block).
* `name` - (Required) Name of the MicroVM image. Changing this value creates a new resource.

The following arguments are optional:

* `additional_os_capabilities` - (Optional) List of additional OS capabilities granted to the MicroVM runtime environment. Valid values: `ALL`.
* `base_image_version` - (Optional) Major version number of the base MicroVM image to use (e.g., `1`). If omitted, the service selects a version.
* `cpu_configuration` - (Optional) CPU configuration for the MicroVM. See [`cpu_configuration` Block](#cpu_configuration-block) below.
* `description` - (Optional) Description of the MicroVM image.
* `egress_network_connectors` - (Optional) List of egress network connectors available to the MicroVM at runtime. Defaults to `["INTERNET_EGRESS"]`.
* `environment_variables` - (Optional) Map of environment variables set in the MicroVM runtime environment.
* `hooks` - (Optional) Lifecycle hook configuration for MicroVMs and MicroVM image builds. See [`hooks` Block](#hooks-block) below.
* `logging` - (Optional) Logging configuration for the image's build-time logs and the runtime logs of MicroVMs launched from it. See [`logging` Block](#logging-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resources` - (Optional) Resource requirements for MicroVMs launched from this image. If omitted, the service default is used. See [`resources` Block](#resources-block) below.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `code_artifact` Block

The `code_artifact` block supports the following:

* `uri` - (Required) S3 URI of the zip archive containing the application code and Dockerfile (e.g., `s3://bucket/code.zip`).

### `cpu_configuration` Block

The `cpu_configuration` block supports the following:

* `architecture` - (Required) CPU architecture for the MicroVM. Valid values are `x86_64` and `arm64`.

### `hooks` Block

The `hooks` block supports the following:

* `microvm_hooks` - (Optional) Lifecycle hooks invoked during MicroVM events. [See below](#microvm_hooks-block).
* `microvm_image_hooks` - (Optional) Hooks invoked during MicroVM image build events. [See below](#microvm_image_hooks-block).
* `port` - (Required) Port number on which the hooks listener runs in the MicroVM. Valid values: `1`-`65535`. `port` must be set whenever the `hooks` block is configured, which is stricter than the API, which only requires a port when a hook is enabled.

### `microvm_hooks` Block

The `microvm_hooks` block supports the following:

* `resume` - (Optional) Whether the hook invoked when the MicroVM resumes from a suspended state is enabled. Valid values: `ENABLED`, `DISABLED`.
* `resume_timeout_in_seconds` - (Optional) Maximum time in seconds for the resume hook to complete. Valid values: `1`-`60`.
* `run` - (Optional) Whether the hook invoked when the MicroVM starts running is enabled. Valid values: `ENABLED`, `DISABLED`. Enabling any MicroVM hook requires the `ready` MicroVM image hook to be enabled.
* `run_timeout_in_seconds` - (Optional) Maximum time in seconds for the run hook to complete. Valid values: `1`-`60`.
* `suspend` - (Optional) Whether the hook invoked when the MicroVM is suspended is enabled. Valid values: `ENABLED`, `DISABLED`.
* `suspend_timeout_in_seconds` - (Optional) Maximum time in seconds for the suspend hook to complete. Valid values: `1`-`60`.
* `terminate` - (Optional) Whether the hook invoked when the MicroVM is terminated is enabled. Valid values: `ENABLED`, `DISABLED`.
* `terminate_timeout_in_seconds` - (Optional) Maximum time in seconds for the terminate hook to complete. Valid values: `1`-`60`.

### `microvm_image_hooks` Block

The `microvm_image_hooks` block supports the following:

* `ready` - (Optional) Whether the hook invoked when the MicroVM image build is ready is enabled. Valid values: `ENABLED`, `DISABLED`.
* `ready_timeout_in_seconds` - (Optional) Maximum time in seconds for the ready hook to complete. Valid values: `1`-`3600`.
* `validate` - (Optional) Whether the hook invoked to validate the MicroVM image build is enabled. Valid values: `ENABLED`, `DISABLED`.
* `validate_timeout_in_seconds` - (Optional) Maximum time in seconds for the validate hook to complete. Valid values: `1`-`3600`.

### `logging` Block

The `logging` block supports exactly one of the following:

* `cloudwatch` - (Optional) Send MicroVM runtime logs to Amazon CloudWatch Logs. [See below](#cloudwatch-block).
* `disabled` - (Optional) Disable build-time and runtime logging. Specify an empty block: `disabled {}`.

### `cloudwatch` Block

The `cloudwatch` block supports the following:

* `log_group` - (Optional) Name of the CloudWatch Logs log group to send logs to.
* `log_stream` - (Optional) Name of the CloudWatch Logs log stream within the log group.

### `resources` Block

The `resources` block supports the following:

* `minimum_memory_in_mib` - (Required) Minimum amount of memory in MiB to allocate to the MicroVM.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Image.
* `created_at` - RFC3339 timestamp when the image was created.
* `image_version` - Current version of the image.
* `latest_active_image_version` - Latest active version of the image.
* `latest_failed_image_version` - Latest failed version of the image, if any.
* `state` - Current state of the image (e.g., `CREATED`).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - RFC3339 timestamp when the image was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_lambdamicrovms_image.example
  identity = {
    "arn" = "arn:aws:lambda:us-east-1:123456789012:microvm-image:example"
  }
}

resource "aws_lambdamicrovms_image" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the Lambda MicroVMs Image.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda MicroVMs Image using the `arn`. For example:

```terraform
import {
  to = aws_lambdamicrovms_image.example
  id = "arn:aws:lambda:us-east-1:123456789012:microvm-image:example"
}
```

Using `terraform import`, import Lambda MicroVMs Image using the `arn`. For example:

```console
% terraform import aws_lambdamicrovms_image.example arn:aws:lambda:us-east-1:123456789012:microvm-image:example
```
