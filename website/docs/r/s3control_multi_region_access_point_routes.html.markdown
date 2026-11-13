---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_multi_region_access_point_routes"
description: |-
  Provides a resource to manage the routing configuration for an S3 Multi-Region Access Point.
---

# Resource: aws_s3control_multi_region_access_point_routes

Provides a resource to manage the routing configuration for an S3 Multi-Region Access Point.

~> Note: Destruction of this resource only removes it from state. It __does not__ alter the configured traffic routing percentages.

## Example Usage

### Active-Active Configuration

```terraform
resource "aws_s3_bucket" "primary" {
  bucket = "example-bucket-primary"
}

resource "aws_s3_bucket" "secondary" {
  provider = awsalternate

  bucket = "example-bucket-secondary"
}

resource "aws_s3control_multi_region_access_point" "example" {
  details {
    name = "example"

    region {
      bucket = aws_s3_bucket.primary.bucket
    }

    region {
      bucket = aws_s3_bucket.secondary.bucket
    }
  }
}

resource "aws_s3control_multi_region_access_point_routes" "example" {
  mrap = aws_s3control_multi_region_access_point.example.arn

  route {
    bucket                  = aws_s3_bucket.primary.bucket
    region                  = aws_s3_bucket.primary.bucket_region
    traffic_dial_percentage = 100
  }

  route {
    bucket                  = aws_s3_bucket.secondary.bucket
    region                  = aws_s3_bucket.secondary.bucket_region
    traffic_dial_percentage = 100
  }
}
```

### Failover Configuration

```terraform
resource "aws_s3control_multi_region_access_point_routes" "example" {
  mrap = aws_s3control_multi_region_access_point.example.arn

  route {
    bucket                  = aws_s3_bucket.primary.bucket
    region                  = aws_s3_bucket.primary.bucket_region
    traffic_dial_percentage = 0
  }

  route {
    bucket                  = aws_s3_bucket.secondary.bucket
    region                  = aws_s3_bucket.secondary.bucket_region
    traffic_dial_percentage = 100
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) AWS account ID for the owner of the Multi-Region Access Point. Defaults to automatically determined account ID of the Terraform AWS provider.
* `mrap` - (Required) ARN of the Multi-Region Access Point.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `route` - (Required) Route configurations. At least one route must have a `traffic_dial_percentage` of `100`. See [`route`](#route-block) below.

### `route` Block

* `bucket` - (Required) Name of the Amazon S3 bucket.
* `region` - (Required) AWS Region where the bucket is located.
* `traffic_dial_percentage` - (Required) Traffic routing configuration. A value of `0` indicates a passive status (traffic will not be routed to the Region), and a value of `100` indicates an active status (traffic will be routed to the Region).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_s3control_multi_region_access_point_routes.example
  identity = {
    mrap = "arn:aws:s3::0123456789012:accesspoint/example"
  }
}

resource "aws_s3control_multi_region_access_point_routes" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `mrap` (String) ARN of the Multi-Region Access Point.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Multi-Region Access Point Routes using the `mrap` argument. For example:

```terraform
import {
  to = aws_s3control_multi_region_access_point_routes.example
  id = "arn:aws:s3::0123456789012:accesspoint/example"
}
```

Using `terraform import`, import Multi-Region Access Point Routes using the `mrap` argument. For example:

```console
% terraform import aws_s3control_multi_region_access_point_routes.example arn:aws:s3::0123456789012:accesspoint/example
```
