---
subcategory: 'ElasticSearch'
layout: 'aws'
page_title: 'AWS: aws_elasticsearch_package'
description: |-
  Terraform resource for managing an AWS Elasticsearch Package.
---

# Resource: aws_elasticsearch_package

Manages an AWS Elasticsearch Package.

## Example Usage

### Basic Usage

```hcl
resource "aws_s3_bucket" "example" {
  bucket = "foo-bucket"
}

resource "aws_s3_bucket_object" "example" {
  bucket  = "${aws_s3_bucket.example.bucket}"
  key     = "synonyms.txt"
  content = "foo, bar"
}

resource "aws_elasticsearch_package" "example" {
  name         = "synonyms"
  type         = "TXT-DICTIONARY"

  source  {
    s3_bucket_name = "${aws_s3_bucket.example.bucket}"
    s3_key         = "${aws_s3_bucket_object.example.key}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Unique name for the package.
* `type` - (Required) Type of the package. Currently supports only `TXT-DICTIONARY`.
* `description` - (Optional) Description of the package.
* `source` - (Required) S3 bucket and key for the package. See below.

**source** supports the following attributes:

* `s3_bucket_name` - (Required) Name of the bucket containing the package.
* `s3_key` - (Required) Key (file name) of the package.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the package.
