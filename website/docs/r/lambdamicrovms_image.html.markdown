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

## Argument Reference

The following arguments are required:

* `base_image_arn` - (Required) ARN of the base MicroVM image. AWS-managed base images use ARNs of the form `arn:aws:lambda:<region>:aws:microvm-image:al2023-1`.
* `build_role_arn` - (Required) ARN of the IAM role used to build the image. The role must be assumable by `lambda.amazonaws.com` and have access to the code artifact.
* `code_artifact` - (Required) Code artifact containing the application code and metadata for the image. [See below](#code_artifact-block).
* `name` - (Required) Name of the MicroVM image. Changing this value creates a new resource.

The following arguments are optional:

* `additional_os_capabilities` - (Optional) List of additional OS capabilities granted to the MicroVM runtime environment. Valid values: `ALL`.
* `base_image_version` - (Optional) Major version number of the base MicroVM image to use (e.g., `1`). If omitted, the service selects a version.
* `description` - (Optional) Description of the MicroVM image.
* `egress_network_connectors` - (Optional) List of egress network connectors available to the MicroVM at runtime. Defaults to `["INTERNET_EGRESS"]`.
* `environment_variables` - (Optional) Map of environment variables set in the MicroVM runtime environment.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `code_artifact` Block

The `code_artifact` block supports the following:

* `uri` - (Required) S3 URI of the zip archive containing the application code and Dockerfile (e.g., `s3://bucket/code.zip`).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Image.
* `latest_active_image_version` - Latest active version of the image.
* `latest_failed_image_version` - Latest failed version of the image, if any.
* `state` - Current state of the image (e.g., `CREATED`).

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
