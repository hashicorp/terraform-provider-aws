---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_resource"
description: |-
  Registers a Lake Formation resource as managed by the Data Catalog.
---

# Resource: aws_lakeformation_resource

Registers a Lake Formation resource (e.g., S3 bucket) as managed by the Data Catalog. In other words, the S3 path is added to the data lake.

Choose a role that has read/write access to the chosen Amazon S3 path or use the service-linked role.
When you register the S3 path, the service-linked role and a new inline policy are created on your behalf.
Lake Formation adds the first path to the inline policy and attaches it to the service-linked role.
When you register subsequent paths, Lake Formation adds the path to the existing policy.

## Example Usage

```terraform
data "aws_s3_bucket" "example" {
  bucket = "an-example-bucket"
}

resource "aws_lakeformation_resource" "example" {
  arn = data.aws_s3_bucket.example.arn
}
```

## Argument Reference

The following arguments are required:

* `arn` – (Required) Amazon Resource Name (ARN) of the resource.

The following arguments are optional:

* `role_arn` – (Optional) Role that has read/write access to the resource.
* `use_service_linked_role` - (Optional) Designates an AWS Identity and Access Management (IAM) service-linked role by registering this role with the Data Catalog.
* `hybrid_access_enabled` - (Optional) Flag to enable AWS LakeFormation hybrid access permission mode.

~> **NOTE:** AWS does not support registering an S3 location with an IAM role and subsequently updating the S3 location registration to a service-linked role.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `last_modified` - Date and time the resource was last modified in [RFC 3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
