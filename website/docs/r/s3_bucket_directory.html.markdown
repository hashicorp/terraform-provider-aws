---
layout: "aws"
page_title: "AWS: aws_s3_bucket_directory"
sidebar_current: "docs-aws-resource-s3-bucket-directory"
description: |-
  Provides an S3 bucket directory resource.
---

# aws_s3_bucket_directory

Provides an S3 bucket directory resource.

## Example Usage

### Uploading a directory to a bucket

```hcl
resource "aws_s3_bucket_directory" "directory" {
  bucket = "your_bucket_name"
  source = "path/to/dir"
  target = "new_directory_name"
}
```

###  Static website

```hcl
resource "aws_s3_bucket" "example_bucket" {
  bucket = "examplebuckettftest"
  acl    = "private"
}

resource "aws_cloudfront_distribution" "example_cdn_" {
  bucket = "examplebuckettftest"
  acl    = "private"
}

resource "aws_s3_bucket_directory" "examplebucket_directory" {
  key        = "somedirectory"
  bucket     = "${aws_s3_bucket.examplebucket.id}"
  source     = "index.html"
  kms_key_id = "${aws_kms_key.examplekms.arn}"
}

resource "null_resource" "invalidate_cloudfront" {
  triggers {
    hash = "${aws_s3_bucket_directory.monkeys.hash}"
  }

  provisioner "local-exec" {
    command = <<EOF
    aws cloudfront create-invalidation --profile "profile" --distribution-id ${var.cloudfront_id} --paths "/*"
    EOF
  }
}
```

## Argument Reference

-> **Note:** Currently `source` is required, any configuration on `files` will therefor be overwritten. These two arguments are mutually-exclusive.

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to put the file in.
* `source` - (Required) The path to a directory, which content will be read and uploaded as raw bytes for the directory content.
* `target` - (Required) The name of the directory once it is in the bucket.
* `exclude` - (Optional) A list of files to exclude from uploading.
* `files` - (Optional) A list of all included files. See details below.

**files** requires:

* `source` - (Required) The path to the local file.
* `target` - (Required) The key of the file on S3.
* `etag` - (Required) MD5 sum of the file.
* `content_type` - (Required) A standard MIME type describing the format of the directory data, e.g. application/octet-stream. All Valid MIME Types are valid for this input.


## Attributes Reference

The following attributes are exported

* `id` - the `target` of the resource supplied above
* `etag` - the ETag generated for the directory (an MD5 sum of the ETag's generated for the content of the directory).
