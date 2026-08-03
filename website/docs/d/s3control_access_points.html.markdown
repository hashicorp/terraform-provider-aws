---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_access_points"
description: |-
  Provides details about an AWS S3 Control Access Points.
---

# Data Source: aws_s3control_access_points

Provides details about an AWS S3 Control Access Points.

## Example Usage

### Basic Usage

```terraform
data "aws_s3control_access_points" "example" {}
```

### Filter by Bucket Name

```terraform
data "aws_s3control_access_points" "example" {
  bucket = aws_s3_bucket.example.bucket
}
```

## Argument Reference

The following arguments are optional:

* `account_id` - (Optional) AWS account ID for the account that owns the specified access points. If omitted, defaults to the caller's account ID.
* `bucket` - (Optional) Name of the bucket associated with the access points.
* `data_source_id` - Unique identifier for the access points data source.
* `data_source_type` - Type of the data source that the access points are attached to. To return all access points set this argument to `ALL`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `access_points` - A list of access points matching the search criteria. See [`access_points`](#access_points-attribute-reference) below.

### `access_points` Attribute Reference

* `access_point_arn` - Access point ARN.
* `alias` - Access point alias.
* `bucket` - Name of the bucket associated with the access point.
* `bucket_account_id` - AWS account ID associated with the S3 bucket associated with the access point.
* `data_source_id` - Unique identifier for the data source of the access point.
* `data_source_type` - Type of the data source that the access point is attached to.
* `name` - Name of the access point.
* `network_origin` - Indicates whether the access point allows access from the public Internet.
* `vpc_configuration` - VPC configuration for the access point. See [`vpc_configuration`](#vpc_configuration-attribute-reference) below.

#### `vpc_configuration` Attribute Reference

* `vpc_id` - Access point will only allow connections from this VPC.
