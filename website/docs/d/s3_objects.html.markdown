---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_objects"
description: |-
    Returns keys and metadata of S3 objects
---

# Data Source: aws_s3_objects

~> **NOTE on `max_keys`:** Retrieving very large numbers of keys can adversely affect Terraform's performance.

The objects data source returns keys (i.e., file names) and other metadata about objects in an S3 bucket.

## Example Usage

The following example retrieves a list of all object keys in an S3 bucket and creates corresponding Terraform object data sources:

```terraform
data "aws_s3_objects" "my_objects" {
  bucket = "ourcorp"
}

data "aws_s3_object" "object_info" {
  count  = length(data.aws_s3_objects.my_objects.keys)
  key    = element(data.aws_s3_objects.my_objects.keys, count.index)
  bucket = data.aws_s3_objects.my_objects.id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `bucket` - (Required) Lists object keys in this S3 bucket. Alternatively, an [S3 access point](https://docs.aws.amazon.com/AmazonS3/latest/dev/using-access-points.html) ARN can be specified
* `prefix` - (Optional) Limits results to object keys with this prefix (Default: none)
* `delimiter` - (Optional) Character used to group keys (Default: none)
* `encoding_type` - (Optional) Encodes keys using this method (Default: none; besides none, only "url" can be used)
* `max_keys` - (Optional) Maximum object keys to return (Default: 1000)
* `start_after` - (Optional) Returns key names lexicographically after a specific object key in your bucket (Default: none; S3 lists object keys in UTF-8 character encoding in lexicographical order)
* `fetch_owner` - (Optional) Boolean specifying whether to populate the owner list (Default: false)
* `request_payer` - (Optional) Confirms that the requester knows that they will be charged for the request. Bucket owners need not specify this parameter in their requests. If included, the only valid value is `requester`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `keys` - List of strings representing object keys
* `common_prefixes` - List of any keys between `prefix` and the next occurrence of `delimiter` (i.e., similar to subdirectories of the `prefix` "directory"); the list is only returned when you specify `delimiter`
* `id` - S3 Bucket.
* `owners` - List of strings representing object owner IDs (see `fetch_owner` above)
* `request_charged` - If present, indicates that the requester was successfully charged for the request.
