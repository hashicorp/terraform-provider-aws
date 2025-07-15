---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3_access_point"
description: |-
  Provides details about a specific S3 access point
---

# Data Source: aws_s3_access_point

Provides details about a specific S3 access point.

## Example Usage

```terraform
data "aws_s3_access_point" "example" {
  name = "example-access-point"
}
```

## Argument Reference

This data source supports the following arguments:

* `account_id` - (Optional) AWS account ID for the account that owns the specified access point.
* `name` - (Required) Name of the access point.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `alias` - Access point alias.
* `bucket_account_id` - AWS account ID associated with the S3 bucket associated with this access point.
* `data_source_id` - Unique identifier for the data source of the access point.
* `data_source_type` - Type of the data source that the access point is attached to.
* `endpoints` - VPC endpoint for the access point.
