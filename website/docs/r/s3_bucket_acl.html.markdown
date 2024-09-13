---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_acl"
description: |-
  Provides an S3 bucket ACL resource.
---

# Resource: aws_s3_bucket_acl

Provides an S3 bucket ACL resource.

~> **Note:** `terraform destroy` does not delete the S3 Bucket ACL but does remove the resource from Terraform state.

-> This resource cannot be used with S3 directory buckets.

## Example Usage

### With `private` ACL

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "my-tf-example-bucket"
}

resource "aws_s3_bucket_ownership_controls" "example" {
  bucket = aws_s3_bucket.example.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "example" {
  depends_on = [aws_s3_bucket_ownership_controls.example]

  bucket = aws_s3_bucket.example.id
  acl    = "private"
}
```

### With `public-read` ACL

-> This example explicitly disables the default S3 bucket security settings. This
should be done with caution, as all bucket objects become publicly exposed.

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "my-tf-example-bucket"
}

resource "aws_s3_bucket_ownership_controls" "example" {
  bucket = aws_s3_bucket.example.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_public_access_block" "example" {
  bucket = aws_s3_bucket.example.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_acl" "example" {
  depends_on = [
    aws_s3_bucket_ownership_controls.example,
    aws_s3_bucket_public_access_block.example,
  ]

  bucket = aws_s3_bucket.example.id
  acl    = "public-read"
}
```

### With Grants

```terraform
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket" "example" {
  bucket = "my-tf-example-bucket"
}

resource "aws_s3_bucket_ownership_controls" "example" {
  bucket = aws_s3_bucket.example.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "example" {
  depends_on = [aws_s3_bucket_ownership_controls.example]

  bucket = aws_s3_bucket.example.id
  access_control_policy {
    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "READ"
    }

    grant {
      grantee {
        type = "Group"
        uri  = "http://acs.amazonaws.com/groups/s3/LogDelivery"
      }
      permission = "READ_ACP"
    }

    owner {
      id = data.aws_canonical_user_id.current.id
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `acl` - (Optional, One of `acl` or `access_control_policy` is required) Canned ACL to apply to the bucket.
* `access_control_policy` - (Optional, One of `access_control_policy` or `acl` is required) Configuration block that sets the ACL permissions for an object per grantee. [See below](#access_control_policy).
* `bucket` - (Required, Forces new resource) Bucket to which to apply the ACL.
* `expected_bucket_owner` - (Optional, Forces new resource) Account ID of the expected bucket owner.

### access_control_policy

The `access_control_policy` configuration block supports the following arguments:

* `grant` - (Required) Set of `grant` configuration blocks. [See below](#grant).
* `owner` - (Required) Configuration block for the bucket owner's display name and ID. [See below](#owner).

### grant

The `grant` configuration block supports the following arguments:

* `grantee` - (Required) Configuration block for the person being granted permissions. [See below](#grantee).
* `permission` - (Required) Logging permissions assigned to the grantee for the bucket. Valid values: `FULL_CONTROL`, `WRITE`, `WRITE_ACP`, `READ`, `READ_ACP`. See [What permissions can I grant?](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#permissions) for more details about what each permission means in the context of buckets.

### owner

The `owner` configuration block supports the following arguments:

* `id` - (Required) ID of the owner.
* `display_name` - (Optional) Display name of the owner.

### grantee

The `grantee` configuration block supports the following arguments:

* `email_address` - (Optional) Email address of the grantee. See [Regions and Endpoints](https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region) for supported AWS regions where this argument can be specified.
* `id` - (Optional) Canonical user ID of the grantee.
* `type` - (Required) Type of grantee. Valid values: `CanonicalUser`, `AmazonCustomerByEmail`, `Group`.
* `uri` - (Optional) URI of the grantee group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The `bucket`, `expected_bucket_owner` (if configured), and `acl` (if configured) separated by commas (`,`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 bucket ACL using `bucket`, `expected_bucket_owner`, and/or `acl`, depending on your situation. For example:

If the owner (account ID) of the source bucket is the _same_ account used to configure the Terraform AWS Provider, and the source bucket is **not configured** with a
[canned ACL][1] (i.e. predefined grant), import using the `bucket`:

```terraform
import {
  to = aws_s3_bucket_acl.example
  id = "bucket-name"
}
```

If the owner (account ID) of the source bucket is the _same_ account used to configure the Terraform AWS Provider, and the source bucket is **configured** with a
[canned ACL][1] (i.e. predefined grant), import using the `bucket` and `acl` separated by a comma (`,`):

```terraform
import {
  to = aws_s3_bucket_acl.example
  id = "bucket-name,private"
}
```

If the owner (account ID) of the source bucket _differs_ from the account used to configure the Terraform AWS Provider, and the source bucket is **not configured** with a [canned ACL][1] (i.e. predefined grant), imported using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```terraform
import {
  to = aws_s3_bucket_acl.example
  id = "bucket-name,123456789012"
}
```

If the owner (account ID) of the source bucket _differs_ from the account used to configure the Terraform AWS Provider, and the source bucket is **configured** with a
[canned ACL][1] (i.e. predefined grant), imported using the `bucket`, `expected_bucket_owner`, and `acl` separated by commas (`,`):

```terraform
import {
  to = aws_s3_bucket_acl.example
  id = "bucket-name,123456789012,private"
}
```

**Using `terraform import` to import** using `bucket`, `expected_bucket_owner`, and/or `acl`, depending on your situation. For example:

If the owner (account ID) of the source bucket is the _same_ account used to configure the Terraform AWS Provider, and the source bucket is **not configured** with a
[canned ACL][1] (i.e. predefined grant), import using the `bucket`:

```console
% terraform import aws_s3_bucket_acl.example bucket-name
```

If the owner (account ID) of the source bucket is the _same_ account used to configure the Terraform AWS Provider, and the source bucket is **configured** with a [canned ACL][1] (i.e. predefined grant), import using the `bucket` and `acl` separated by a comma (`,`):

```console
% terraform import aws_s3_bucket_acl.example bucket-name,private
```

If the owner (account ID) of the source bucket _differs_ from the account used to configure the Terraform AWS Provider, and the source bucket is **not configured** with a [canned ACL][1] (i.e. predefined grant), imported using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```console
% terraform import aws_s3_bucket_acl.example bucket-name,123456789012
```

If the owner (account ID) of the source bucket _differs_ from the account used to configure the Terraform AWS Provider, and the source bucket is **configured** with a [canned ACL][1] (i.e. predefined grant), imported using the `bucket`, `expected_bucket_owner`, and `acl` separated by commas (`,`):

```console
% terraform import aws_s3_bucket_acl.example bucket-name,123456789012,private
```

[1]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl
