---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_resource"
description: |-
  Manages the data (Amazon S3 buckets and folders) that is being registered with AWS Lake Formation
---

# Resource: aws_lakeformation_resource

Manages the data (Amazon S3 buckets and folders) that is being registered with AWS Lake Formation.

## Example Usage

```hcl
data "aws_s3_bucket" "example" {
  bucket = "an-example-bucket"
}

resource "aws_lakeformation_resource" "example" {
  resource_arn            = "${data.aws_s3_bucket.example.arn}"
  use_service_linked_role = true
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` – (Required) The Amazon Resource Name (ARN) of the resource.

* `use_service_linked_role` – (Required) Designates a trusted caller, an IAM principal, by registering this caller with the Data Catalog.

The following arguments are optional:

* `role_arn` – (Optional) The IAM role that registered a resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `last_modified` - (Optional) The date and time the resource was last modified.
