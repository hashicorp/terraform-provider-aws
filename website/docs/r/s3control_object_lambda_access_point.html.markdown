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

* `account_id` - (Optional) AWS account ID for the owner of the bucket for which you want to create an Object Lambda Access Point. Defaults to automatically determined account ID of the Terraform AWS provider.
* `configuration` - (Required) Configuration block containing details about the Object Lambda Access Point. See [`configuration` Block](#configuration-block) below for more details.
* `name` - (Required) Name for this Object Lambda Access Point.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `configuration` Block

The `configuration` block supports the following:

* `allowed_features` - (Optional) Allowed features. Valid values: `GetObject-Range`, `GetObject-PartNumber`.
* `cloud_watch_metrics_enabled` - (Optional) Whether or not the CloudWatch metrics configuration is enabled.
* `supporting_access_point` - (Required) Standard access point associated with the Object Lambda Access Point.
* `transformation_configuration` - (Required) List of transformation configurations for the Object Lambda Access Point. See [`transformation_configuration` Block](#transformation_configuration-block) below for more details.

### `transformation_configuration` Block

The `transformation_configuration` block supports the following:

* `actions` - (Required) Actions of an Object Lambda Access Point configuration. Valid values: `GetObject`.
* `content_transformation` - (Required) Content transformation of an Object Lambda Access Point configuration. See [`content_transformation` Block](#content_transformation-block) below for more details.

### `content_transformation` Block

The `content_transformation` block supports the following:

* `aws_lambda` - (Required) Configuration for an AWS Lambda function. See [`aws_lambda` Block](#aws_lambda-block) below for more details.

### `aws_lambda` Block

The `aws_lambda` block supports the following:

* `function_arn` - (Required) Amazon Resource Name (ARN) of the AWS Lambda function.
* `function_payload` - (Optional) Additional JSON that provides supplemental data to the Lambda function used to transform objects.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `alias` - Alias for the S3 Object Lambda Access Point.
* `arn` - Amazon Resource Name (ARN) of the Object Lambda Access Point.
* `id` - AWS account ID and access point name separated by a colon (`:`).

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
