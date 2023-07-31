---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_object_lambda_access_point"
description: |-
  Provides a resource to manage an S3 Object Lambda Access Point.
---

# Resource: aws_s3control_object_lambda_access_point

Provides a resource to manage an S3 Object Lambda Access Point.
An Object Lambda access point is associated with exactly one [standard access point](s3_access_point.html) and thus one Amazon S3 bucket.

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
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The AWS account ID for the owner of the bucket for which you want to create an Object Lambda Access Point. Defaults to automatically determined account ID of the Terraform AWS provider.
* `configuration` - (Required) A configuration block containing details about the Object Lambda Access Point. See [Configuration](#configuration) below for more details.
* `name` - (Required) The name for this Object Lambda Access Point.

### Configuration

The `configuration` block supports the following:

* `allowed_features` - (Optional) Allowed features. Valid values: `GetObject-Range`, `GetObject-PartNumber`.
* `cloud_watch_metrics_enabled` - (Optional) Whether or not the CloudWatch metrics configuration is enabled.
* `supporting_access_point` - (Required) Standard access point associated with the Object Lambda Access Point.
* `transformation_configuration` - (Required) List of transformation configurations for the Object Lambda Access Point. See [Transformation Configuration](#transformation-configuration) below for more details.

### Transformation Configuration

The `transformation_configuration` block supports the following:

* `actions` - (Required) The actions of an Object Lambda Access Point configuration. Valid values: `GetObject`.
* `content_transformation` - (Required) The content transformation of an Object Lambda Access Point configuration. See [Content Transformation](#content-transformation) below for more details.

### Content Transformation

The `content_transformation` block supports the following:

* `aws_lambda` - (Required) Configuration for an AWS Lambda function. See [AWS Lambda](#aws-lambda) below for more details.

### AWS Lambda

The `aws_lambda` block supports the following:

* `function_arn` - (Required) The Amazon Resource Name (ARN) of the AWS Lambda function.
* `function_payload` - (Optional) Additional JSON that provides supplemental data to the Lambda function used to transform objects.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Object Lambda Access Point.
* `id` - The AWS account ID and access point name separated by a colon (`:`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Object Lambda Access Points using the `account_id` and `name`, separated by a colon (`:`). For example:

```terraform
import {
  to = aws_s3control_object_lambda_access_point.example
  id = "123456789012:example"
}
```

Using `terraform import`, import Object Lambda Access Points using the `account_id` and `name`, separated by a colon (`:`). For example:

```console
% terraform import aws_s3control_object_lambda_access_point.example 123456789012:example
```
