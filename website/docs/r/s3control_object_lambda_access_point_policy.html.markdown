---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_object_lambda_access_point_policy"
description: |-
  Provides a resource to manage an S3 Object Lambda Access Point resource policy.
---

# Resource: aws_s3control_object_lambda_access_point_policy

Provides a resource to manage an S3 Object Lambda Access Point resource policy.

## Example Usage

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_access_point" "example" {
  bucket = aws_s3_bucket.example.id
  name   = "example"
}

resource "aws_s3control_object_lambda_access_point" "example" {
  name = "example"

  configuration {
    supporting_access_point = aws_s3_access_point.example.arn

    transformation_configuration {
      actions = ["GetObject"]

      content_transformation {
        aws_lambda {
          function_arn = aws_lambda_function.example.arn
        }
      }
    }
  }
}

resource "aws_s3control_object_lambda_access_point_policy" "example" {
  name = aws_s3control_object_lambda_access_point.example.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "s3-object-lambda:GetObject"
      Principal = {
        AWS = data.aws_caller_identity.current.account_id
      }
      Resource = aws_s3control_object_lambda_access_point.example.arn
    }]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The AWS account ID for the account that owns the Object Lambda Access Point. Defaults to automatically determined account ID of the Terraform AWS provider.
* `name` - (Required) The name of the Object Lambda Access Point.
* `policy` - (Required) The Object Lambda Access Point resource policy document.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `has_public_access_policy` - Indicates whether this access point currently has a policy that allows public access.
* `id` - The AWS account ID and access point name separated by a colon (`:`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Object Lambda Access Point policies using the `account_id` and `name`, separated by a colon (`:`). For example:

```terraform
import {
  to = aws_s3control_object_lambda_access_point_policy.example
  id = "123456789012:example"
}
```

Using `terraform import`, import Object Lambda Access Point policies using the `account_id` and `name`, separated by a colon (`:`). For example:

```console
% terraform import aws_s3control_object_lambda_access_point_policy.example 123456789012:example
```
