---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_trust_store"
description: |-
  Manages an AWS CloudFront Trust Store.
---

# Resource: aws_cloudfront_trust_store

Manages an AWS CloudFront Trust Store.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudfront_trust_store" "example" {
  name = "example-trust-store"

  ca_certificates_bundle_source {
    ca_certificates_bundle_s3_location {
      bucket = "example-bucket"
      key    = "ca-certificates.pem"
      region = "us-east-1"
    }
  }
}
```

### With S3 Object Version

```terraform
resource "aws_cloudfront_trust_store" "example" {
  name = "example-trust-store"

  ca_certificates_bundle_source {
    ca_certificates_bundle_s3_location {
      bucket  = "example-bucket"
      key     = "ca-certificates.pem"
      region  = "us-east-1"
      version = "abc123"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the trust store. Changing this forces a new resource to be created.
* `ca_certificates_bundle_source` - (Required) Configuration block for the CA certificates bundle source. See [`ca_certificates_bundle_source`](#ca_certificates_bundle_source) below.

### ca_certificates_bundle_source

* `ca_certificates_bundle_s3_location` - (Required) Configuration block for the S3 location of the CA certificates bundle. See [`ca_certificates_bundle_s3_location`](#ca_certificates_bundle_s3_location) below.

### ca_certificates_bundle_s3_location

* `bucket` - (Required) S3 bucket name containing the CA certificates bundle.
* `key` - (Required) S3 object key for the CA certificates bundle.
* `region` - (Required) AWS region of the S3 bucket.
* `version` - (Optional) S3 object version ID for the CA certificates bundle.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the trust store.
* `etag` - ETag of the trust store.
* `id` - ID of the trust store.
* `last_modified_time` - Date and time when the trust store was last modified.
* `number_of_ca_certificates` - Number of CA certificates in the trust store.
* `status` - Status of the trust store.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront Trust Store using the trust store ID. For example:

```terraform
import {
  to = aws_cloudfront_trust_store.example
  id = "ts_12abcXYZhA4Q6RS6tuvW5Xy0ZZZ"
}
```

Using `terraform import`, import CloudFront Trust Store using the trust store ID. For example:

```console
% terraform import aws_cloudfront_trust_store.example ts_12abcXYZhA4Q6RS6tuvW5Xy0ZZZ
```
