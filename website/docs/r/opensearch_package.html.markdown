---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_package"
description: |-
  Terraform resource for managing an AWS OpenSearch package.
---

# Resource: aws_opensearch_package

Manages an AWS Opensearch Package.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3_bucket" "my_opensearch_packages" {
  bucket = "my-opensearch-packages"
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket.my_opensearch_packages.bucket
  key    = "example.txt"
  source = "./example.txt"
  etag   = filemd5("./example.txt")
}

resource "aws_opensearch_package" "example" {
  package_name = "example-txt"
  package_source {
    s3_bucket_name = aws_s3_bucket.my_opensearch_packages.bucket
    s3_key         = aws_s3_object.example.key
  }
  package_type = "TXT-DICTIONARY"
}
```

## Argument Reference

This resource supports the following arguments:

* `package_name` - (Required, Forces new resource) Unique name for the package.
* `package_type` - (Required, Forces new resource) The type of package.
* `package_source` - (Required, Forces new resource) Configuration block for the package source options.
* `package_description` - (Optional, Forces new resource) Description of the package.

### package_source

* `s3_bucket_name` - (Required, Forces new resource) The name of the Amazon S3 bucket containing the package.
* `s3_key` - (Required, Forces new resource) Key (file name) of the package.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Id of the package.
* `available_package_version` - The current version of the package.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS Opensearch Packages using the Package ID. For example:

```terraform
import {
  to = aws_opensearch_package.example
  id = "package-id"
}
```

Using `terraform import`, import AWS Opensearch Packages using the Package ID. For example:

```console
% terraform import aws_opensearch_package.example package-id
```
