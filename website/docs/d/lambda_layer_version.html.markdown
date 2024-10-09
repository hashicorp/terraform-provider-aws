---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_layer_version"
description: |-
  Provides a Lambda Layer Version data source.
---

# aws_lambda_layer_version

Provides information about a Lambda Layer Version.

## Example Usage

```terraform
variable "layer_name" {
  type = string
}

data "aws_lambda_layer_version" "existing" {
  layer_name = var.layer_name
}
```

## Argument Reference

This data source supports the following arguments:

* `layer_name` - (Required) Name of the lambda layer.
* `version` - (Optional) Specific layer version. Conflicts with `compatible_runtime` and `compatible_architecture`. If omitted, the latest available layer version will be used.
* `compatible_runtime` (Optional) Specific runtime the layer version must support. Conflicts with `version`. If specified, the latest available layer version supporting the provided runtime will be used.
* `compatible_architecture` (Optional) Specific architecture the layer version could support. Conflicts with `version`. If specified, the latest available layer version supporting the provided architecture will be used.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `code_sha256` - Base64-encoded representation of raw SHA-256 sum of the zip file.
* `description` - Description of the specific Lambda Layer version.
* `license_info` - License info associated with the specific Lambda Layer version.
* `compatible_runtimes` - List of [Runtimes][1] the specific Lambda Layer version is compatible with.
* `compatible_architectures` - A list of [Architectures][2] the specific Lambda Layer version is compatible with.
* `arn` - ARN of the Lambda Layer with version.
* `layer_arn` - ARN of the Lambda Layer without version.
* `created_date` - Date this resource was created.
* `signing_job_arn` - ARN of a signing job.
* `signing_profile_version_arn` - The ARN for a signing profile version.
* `source_code_hash` - (**Deprecated** use `code_sha256` instead) Base64-encoded representation of raw SHA-256 sum of the zip file.
* `source_code_size` - Size in bytes of the function .zip file.
* `version` - This Lambda Layer version.

[1]: https://docs.aws.amazon.com/lambda/latest/dg/API_GetLayerVersion.html#SSS-GetLayerVersion-response-CompatibleRuntimes
[2]: https://docs.aws.amazon.com/lambda/latest/dg/API_GetLayerVersion.html#SSS-GetLayerVersion-response-CompatibleArchitectures
