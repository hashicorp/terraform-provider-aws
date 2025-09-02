---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_layer_version"
description: |-
  Manages an AWS Lambda Layer Version.
---

# Resource: aws_lambda_layer_version

Manages an AWS Lambda Layer Version. Use this resource to share code and dependencies across multiple Lambda functions.

For information about Lambda Layers and how to use them, see [AWS Lambda Layers](https://docs.aws.amazon.com/lambda/latest/dg/configuration-layers.html).

~> **Note:** Setting `skip_destroy` to `true` means that the AWS Provider will not destroy any layer version, even when running `terraform destroy`. Layer versions are thus intentional dangling resources that are not managed by Terraform and may incur extra expense in your AWS account.

## Example Usage

### Basic Layer

```terraform
resource "aws_lambda_layer_version" "example" {
  filename   = "lambda_layer_payload.zip"
  layer_name = "lambda_layer_name"

  compatible_runtimes = ["nodejs20.x"]
}
```

### Layer with S3 Source

```terraform
resource "aws_lambda_layer_version" "example" {
  s3_bucket = aws_s3_object.lambda_layer_zip.bucket
  s3_key    = aws_s3_object.lambda_layer_zip.key

  layer_name = "lambda_layer_name"

  compatible_runtimes      = ["nodejs20.x", "python3.12"]
  compatible_architectures = ["x86_64", "arm64"]
}
```

### Layer with Multiple Runtimes and Architectures

```terraform
resource "aws_lambda_layer_version" "example" {
  filename         = "lambda_layer_payload.zip"
  layer_name       = "multi_runtime_layer"
  description      = "Shared utilities for Lambda functions"
  license_info     = "MIT"
  source_code_hash = filebase64sha256("lambda_layer_payload.zip")

  compatible_runtimes = [
    "nodejs18.x",
    "nodejs20.x",
    "python3.11",
    "python3.12"
  ]

  compatible_architectures = ["x86_64", "arm64"]
}
```

## Specifying the Deployment Package

AWS Lambda Layers expect source code to be provided as a deployment package whose structure varies depending on which `compatible_runtimes` this layer specifies. See [Runtimes](https://docs.aws.amazon.com/lambda/latest/dg/API_PublishLayerVersion.html#SSS-PublishLayerVersion-request-CompatibleRuntimes) for the valid values of `compatible_runtimes`.

Once you have created your deployment package you can specify it either directly as a local file (using the `filename` argument) or indirectly via Amazon S3 (using the `s3_bucket`, `s3_key` and `s3_object_version` arguments). When providing the deployment package via S3 it may be useful to use [the `aws_s3_object` resource](s3_object.html) to upload it.

For larger deployment packages it is recommended by Amazon to upload via S3, since the S3 API has better support for uploading large files efficiently.

## Argument Reference

The following arguments are required:

* `layer_name` - (Required) Unique name for your Lambda Layer.

The following arguments are optional:

* `compatible_architectures` - (Optional) List of [Architectures](https://docs.aws.amazon.com/lambda/latest/dg/API_PublishLayerVersion.html#SSS-PublishLayerVersion-request-CompatibleArchitectures) this layer is compatible with. Currently `x86_64` and `arm64` can be specified.
* `compatible_runtimes` - (Optional) List of [Runtimes](https://docs.aws.amazon.com/lambda/latest/dg/API_PublishLayerVersion.html#SSS-PublishLayerVersion-request-CompatibleRuntimes) this layer is compatible with. Up to 15 runtimes can be specified.
* `description` - (Optional) Description of what your Lambda Layer does.
* `filename` - (Optional) Path to the function's deployment package within the local filesystem. If defined, The `s3_`-prefixed options cannot be used.
* `license_info` - (Optional) License info for your Lambda Layer. See [License Info](https://docs.aws.amazon.com/lambda/latest/dg/API_PublishLayerVersion.html#SSS-PublishLayerVersion-request-LicenseInfo).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `s3_bucket` - (Optional) S3 bucket location containing the function's deployment package. Conflicts with `filename`. This bucket must reside in the same AWS region where you are creating the Lambda function.
* `s3_key` - (Optional) S3 key of an object containing the function's deployment package. Conflicts with `filename`.
* `s3_object_version` - (Optional) Object version containing the function's deployment package. Conflicts with `filename`.
* `skip_destroy` - (Optional) Whether to retain the old version of a previously deployed Lambda Layer. Default is `false`. When this is not set to `true`, changing any of `compatible_architectures`, `compatible_runtimes`, `description`, `filename`, `layer_name`, `license_info`, `s3_bucket`, `s3_key`, `s3_object_version`, or `source_code_hash` forces deletion of the existing layer version and creation of a new layer version.
* `source_code_hash` - (Optional) Virtual attribute used to trigger replacement when source code changes. Must be set to a base64-encoded SHA256 hash of the package file specified with either `filename` or `s3_key`. The usual way to set this is `filebase64sha256("file.zip")` (Terraform 0.11.12 or later) or `base64sha256(file("file.zip"))` (Terraform 0.11.11 and earlier), where "file.zip" is the local filename of the lambda layer source archive.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Lambda Layer with version.
* `code_sha256` - Base64-encoded representation of raw SHA-256 sum of the zip file.
* `created_date` - Date this resource was created.
* `layer_arn` - ARN of the Lambda Layer without version.
* `signing_job_arn` - ARN of a signing job.
* `signing_profile_version_arn` - ARN for a signing profile version.
* `source_code_size` - Size in bytes of the function .zip file.
* `version` - Lambda Layer version.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Layers using `arn`. For example:

```terraform
import {
  to = aws_lambda_layer_version.example
  id = "arn:aws:lambda:us-west-2:123456789012:layer:example:1"
}
```

Using `terraform import`, import Lambda Layers using `arn`. For example:

```console
% terraform import aws_lambda_layer_version.example arn:aws:lambda:us-west-2:123456789012:layer:example:1
```
