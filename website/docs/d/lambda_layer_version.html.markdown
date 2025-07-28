---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_layer_version"
description: |-
  Provides details about an AWS Lambda Layer Version.
---

# Data Source: aws_lambda_layer_version

Provides details about an AWS Lambda Layer Version. Use this data source to retrieve information about a specific layer version or find the latest version compatible with your runtime and architecture requirements.

## Example Usage

### Get Latest Layer Version

```terraform
data "aws_lambda_layer_version" "example" {
  layer_name = "my-shared-utilities"
}

# Use the layer in a Lambda function
resource "aws_lambda_function" "example" {
  filename      = "function.zip"
  function_name = "example_function"
  role          = aws_iam_role.lambda_role.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"

  layers = [data.aws_lambda_layer_version.example.arn]
}
```

### Get Specific Layer Version

```terraform
data "aws_lambda_layer_version" "example" {
  layer_name = "production-utilities"
  version    = 5
}

output "layer_info" {
  value = {
    arn         = data.aws_lambda_layer_version.example.arn
    version     = data.aws_lambda_layer_version.example.version
    description = data.aws_lambda_layer_version.example.description
  }
}
```

### Get Latest Compatible Layer Version

```terraform
# Find latest layer version compatible with Python 3.12
data "aws_lambda_layer_version" "python_layer" {
  layer_name         = "python-dependencies"
  compatible_runtime = "python3.12"
}

# Find latest layer version compatible with ARM64 architecture
data "aws_lambda_layer_version" "arm_layer" {
  layer_name              = "optimized-libraries"
  compatible_architecture = "arm64"
}

# Use both layers in a function
resource "aws_lambda_function" "example" {
  filename      = "function.zip"
  function_name = "multi_layer_function"
  role          = aws_iam_role.lambda_role.arn
  handler       = "app.handler"
  runtime       = "python3.12"
  architectures = ["arm64"]

  layers = [
    data.aws_lambda_layer_version.python_layer.arn,
    data.aws_lambda_layer_version.arm_layer.arn,
  ]
}
```

### Compare Layer Versions

```terraform
# Get latest version
data "aws_lambda_layer_version" "latest" {
  layer_name = "shared-layer"
}

# Get specific version for comparison
data "aws_lambda_layer_version" "stable" {
  layer_name = "shared-layer"
  version    = 3
}

locals {
  use_latest_layer = data.aws_lambda_layer_version.latest.version > 5
  selected_layer   = local.use_latest_layer ? data.aws_lambda_layer_version.latest.arn : data.aws_lambda_layer_version.stable.arn
}

output "selected_layer_version" {
  value = local.use_latest_layer ? data.aws_lambda_layer_version.latest.version : data.aws_lambda_layer_version.stable.version
}
```

## Argument Reference

The following arguments are required:

* `layer_name` - (Required) Name of the Lambda layer.

The following arguments are optional:

* `compatible_architecture` - (Optional) Specific architecture the layer version must support. Conflicts with `version`. If specified, the latest available layer version supporting the provided architecture will be used.
* `compatible_runtime` - (Optional) Specific runtime the layer version must support. Conflicts with `version`. If specified, the latest available layer version supporting the provided runtime will be used.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `version` - (Optional) Specific layer version. Conflicts with `compatible_runtime` and `compatible_architecture`. If omitted, the latest available layer version will be used.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Lambda Layer with version.
* `code_sha256` - Base64-encoded representation of raw SHA-256 sum of the zip file.
* `compatible_architectures` - List of [Architectures](https://docs.aws.amazon.com/lambda/latest/dg/API_GetLayerVersion.html#SSS-GetLayerVersion-response-CompatibleArchitectures) the specific Lambda Layer version is compatible with.
* `compatible_runtimes` - List of [Runtimes](https://docs.aws.amazon.com/lambda/latest/dg/API_GetLayerVersion.html#SSS-GetLayerVersion-response-CompatibleRuntimes) the specific Lambda Layer version is compatible with.
* `created_date` - Date this resource was created.
* `description` - Description of the specific Lambda Layer version.
* `layer_arn` - ARN of the Lambda Layer without version.
* `license_info` - License info associated with the specific Lambda Layer version.
* `signing_job_arn` - ARN of a signing job.
* `signing_profile_version_arn` - ARN for a signing profile version.
* `source_code_hash` - (**Deprecated** use `code_sha256` instead) Base64-encoded representation of raw SHA-256 sum of the zip file.
* `source_code_size` - Size in bytes of the function .zip file.
* `version` - Lambda Layer version.
