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
  bucket  = "your_bucket_name"
  source  = "path/to/dir"
  target  = "new_directory_name"
  exclude = ["excluded_file"]
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to put the directory in.
* `source` - (Required) The path to a directory, which content will be read and uploaded as raw bytes for the directory content.
* `target` - (Required) The name of the directory once it is in the bucket.
* `exclude` - (Optional) A list of files to exclude from uploading.
* `files` - (Computed) A list of all included files, this is computed based on `source`. See details below.

**files** will present the following attributes for each file:

* `source` - (Computed) The path to the local file.
* `target` - (Computed) The key of the file on S3.
* `etag` - (Computed) MD5 sum of the file.
* `content_type` - (Computed) A standard MIME type describing the format of the directory data, e.g. application/octet-stream.

## Attributes Reference

The following attributes are exported:

* `id` - the `target` of the resource supplied above
* `etag` - the ETag generated for the directory (an MD5 sum of the ETag's generated for the content of the directory).
