---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_lifecycle_configuration"
description: |-
  Provides a S3 bucket lifecycle configuration resource.
---

# Resource: aws_s3_bucket_lifecycle_configuration

Provides an independent configuration resource for S3 bucket [lifecycle configuration](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html).

An S3 Lifecycle configuration consists of one or more Lifecycle rules. Each rule consists of the following:

* Rule metadata (`id` and `status`)
* [Filter](#filter) identifying objects to which the rule applies
* One or more transition or expiration actions

For more information see the Amazon S3 User Guide on [`Lifecycle Configuration Elements`](https://docs.aws.amazon.com/AmazonS3/latest/userguide/intro-lifecycle-rules.html).

~> S3 Buckets only support a single lifecycle configuration. Declaring multiple `aws_s3_bucket_lifecycle_configuration` resources to the same S3 Bucket will cause a perpetual difference in configuration.

~> Lifecycle configurations may take some time to fully propagate to all AWS S3 systems.
Running Terraform operations shortly after creating a lifecycle configuration may result in changes that affect configuration idempotence.
See the Amazon S3 User Guide on [setting lifecycle configuration on a bucket](https://docs.aws.amazon.com/AmazonS3/latest/userguide/how-to-set-lifecycle-configuration-intro.html).

## Example Usage

### With neither a filter nor prefix specified

The Lifecycle rule applies to a subset of objects based on the key name prefix (`""`).

This configuration is intended to replicate the default behavior of the `lifecycle_rule`
parameter in the Terraform AWS Provider `aws_s3_bucket` resource prior to `v4.0`.

```terraform
resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id = "rule-1"

    # ... other transition/expiration actions ...

    status = "Enabled"
  }
}
```

### Specifying an empty filter

The Lifecycle rule applies to all objects in the bucket.

```terraform
resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id = "rule-1"

    filter {}

    # ... other transition/expiration actions ...

    status = "Enabled"
  }
}
```

### Specifying a filter using key prefixes

The Lifecycle rule applies to a subset of objects based on the key name prefix (`logs/`).

```terraform
resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id = "rule-1"

    filter {
      prefix = "logs/"
    }

    # ... other transition/expiration actions ...

    status = "Enabled"
  }
}
```

If you want to apply a Lifecycle action to a subset of objects based on different key name prefixes, specify separate rules.

```terraform
resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id = "rule-1"

    filter {
      prefix = "logs/"
    }

    # ... other transition/expiration actions ...

    status = "Enabled"
  }

  rule {
    id = "rule-2"

    filter {
      prefix = "tmp/"
    }

    # ... other transition/expiration actions ...

    status = "Enabled"
  }
}
```

### Specifying a filter based on an object tag

The Lifecycle rule specifies a filter based on a tag key and value. The rule then applies only to a subset of objects with the specific tag.

```terraform
resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id = "rule-1"

    filter {
      tag {
        key   = "Name"
        value = "Staging"
      }
    }

    # ... other transition/expiration actions ...

    status = "Enabled"
  }
}
```

### Specifying a filter based on multiple tags

The Lifecycle rule directs Amazon S3 to perform lifecycle actions on objects with two tags (with the specific tag keys and values). Notice `tags` is wrapped in the `and` configuration block.

```terraform
resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id = "rule-1"

    filter {
      and {
        tags = {
          Key1 = "Value1"
          Key2 = "Value2"
        }
      }
    }

    # ... other transition/expiration actions ...

    status = "Enabled"
  }
}
```

### Specifying a filter based on both prefix and one or more tags

The Lifecycle rule directs Amazon S3 to perform lifecycle actions on objects with the specified prefix and two tags (with the specific tag keys and values). Notice both `prefix` and `tags` are wrapped in the `and` configuration block.

```terraform
resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id = "rule-1"

    filter {
      and {
        prefix = "logs/"
        tags = {
          Key1 = "Value1"
          Key2 = "Value2"
        }
      }
    }

    # ... other transition/expiration actions ...

    status = "Enabled"
  }
}
```

### Specifying a filter based on object size

Object size values are in bytes. Maximum filter size is 5TB. Amazon S3 applies a default behavior to your Lifecycle configuration that prevents objects smaller than 128 KB from being transitioned to any storage class. You can allow smaller objects to transition by adding a minimum size (`object_size_greater_than`) or a maximum size (`object_size_less_than`) filter that specifies a smaller size to the configuration. This example allows any object smaller than 128 KB to transition to the S3 Glacier Instant Retrieval storage class:

```terraform
resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id = "Allow small object transitions"

    filter {
      object_size_greater_than = 1
    }

    status = "Enabled"

    transition {
      days          = 365
      storage_class = "GLACIER_IR"
    }
  }
}
```

### Specifying a filter based on object size range and prefix

The `object_size_greater_than` must be less than the `object_size_less_than`. Notice both the object size range and prefix are wrapped in the `and` configuration block.

```terraform
resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id = "rule-1"

    filter {
      and {
        prefix                   = "logs/"
        object_size_greater_than = 500
        object_size_less_than    = 64000
      }
    }

    # ... other transition/expiration actions ...

    status = "Enabled"
  }
}
```

### Creating a Lifecycle Configuration for a bucket with versioning

```terraform
resource "aws_s3_bucket" "bucket" {
  bucket = "my-bucket"
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "bucket-config" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    id = "log"

    expiration {
      days = 90
    }

    filter {
      and {
        prefix = "log/"

        tags = {
          rule      = "log"
          autoclean = "true"
        }
      }
    }

    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "GLACIER"
    }
  }

  rule {
    id = "tmp"

    filter {
      prefix = "tmp/"
    }

    expiration {
      date = "2023-01-13T00:00:00Z"
    }

    status = "Enabled"
  }
}

resource "aws_s3_bucket" "versioning_bucket" {
  bucket = "my-versioning-bucket"
}

resource "aws_s3_bucket_acl" "versioning_bucket_acl" {
  bucket = aws_s3_bucket.versioning_bucket.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "versioning" {
  bucket = aws_s3_bucket.versioning_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "versioning-bucket-config" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.versioning]

  bucket = aws_s3_bucket.versioning_bucket.id

  rule {
    id = "config"

    filter {
      prefix = "config/"
    }

    noncurrent_version_expiration {
      noncurrent_days = 90
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }

    noncurrent_version_transition {
      noncurrent_days = 60
      storage_class   = "GLACIER"
    }

    status = "Enabled"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket` - (Required) Name of the source S3 bucket you want Amazon S3 to monitor.
* `expected_bucket_owner` - (Optional) Account ID of the expected bucket owner. If the bucket is owned by a different account, the request will fail with an HTTP 403 (Access Denied) error.
* `rule` - (Required) List of configuration blocks describing the rules managing the replication. [See below](#rule).
* `transition_default_minimum_object_size` - (Optional) The default minimum object size behavior applied to the lifecycle configuration. Valid values: `all_storage_classes_128K` (default), `varies_by_storage_class`. To customize the minimum object size for any transition you can add a `filter` that specifies a custom `object_size_greater_than` or `object_size_less_than` value. Custom filters always take precedence over the default transition behavior.

### rule

~> The `filter` argument, while Optional, is required if the `rule` configuration block does not contain a `prefix` **and** you intend to override the default behavior of setting the rule to filter objects with the empty string prefix (`""`).
Since `prefix` is deprecated by Amazon S3 and will be removed in the next major version of the Terraform AWS Provider, we recommend users either specify `filter` or leave both `filter` and `prefix` unspecified.

~> A rule cannot be updated from having a filter (via either the `rule.filter` parameter or when neither `rule.filter` and `rule.prefix` are specified) to only having a prefix via the `rule.prefix` parameter.

~> Terraform cannot distinguish a difference between configurations that use `rule.filter {}` and configurations that neither use `rule.filter` nor `rule.prefix`, so a rule cannot be updated from applying to all objects in the bucket via `rule.filter {}` to applying to a subset of objects based on the key prefix `""` and vice versa.

The `rule` configuration block supports the following arguments:

* `abort_incomplete_multipart_upload` - (Optional) Configuration block that specifies the days since the initiation of an incomplete multipart upload that Amazon S3 will wait before permanently removing all parts of the upload. [See below](#abort_incomplete_multipart_upload).
* `expiration` - (Optional) Configuration block that specifies the expiration for the lifecycle of the object in the form of date, days and, whether the object has a delete marker. [See below](#expiration).
* `filter` - (Optional) Configuration block used to identify objects that a Lifecycle Rule applies to. [See below](#filter). If not specified, the `rule` will default to using `prefix`.
* `id` - (Required) Unique identifier for the rule. The value cannot be longer than 255 characters.
* `noncurrent_version_expiration` - (Optional) Configuration block that specifies when noncurrent object versions expire. [See below](#noncurrent_version_expiration).
* `noncurrent_version_transition` - (Optional) Set of configuration blocks that specify the transition rule for the lifecycle rule that describes when noncurrent objects transition to a specific storage class. [See below](#noncurrent_version_transition).
* `prefix` - (Optional) **DEPRECATED** Use `filter` instead. This has been deprecated by Amazon S3. Prefix identifying one or more objects to which the rule applies. Defaults to an empty string (`""`) if `filter` is not specified.
* `status` - (Required) Whether the rule is currently being applied. Valid values: `Enabled` or `Disabled`.
* `transition` - (Optional) Set of configuration blocks that specify when an Amazon S3 object transitions to a specified storage class. [See below](#transition).

### abort_incomplete_multipart_upload

The `abort_incomplete_multipart_upload` configuration block supports the following arguments:

* `days_after_initiation` - Number of days after which Amazon S3 aborts an incomplete multipart upload.

### expiration

The `expiration` configuration block supports the following arguments:

* `date` - (Optional) Date the object is to be moved or deleted. The date value must be in [RFC3339 full-date format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.6) e.g. `2023-08-22`.
* `days` - (Optional) Lifetime, in days, of the objects that are subject to the rule. The value must be a non-zero positive integer.
* `expired_object_delete_marker` - (Optional, Conflicts with `date` and `days`) Indicates whether Amazon S3 will remove a delete marker with no noncurrent versions. If set to `true`, the delete marker will be expired; if set to `false` the policy takes no action.

### filter

~> The `filter` configuration block must either be specified as the empty configuration block (`filter {}`) or with exactly one of `prefix`, `tag`, `and`, `object_size_greater_than` or `object_size_less_than` specified.

The `filter` configuration block supports the following arguments:

* `and`- (Optional) Configuration block used to apply a logical `AND` to two or more predicates. [See below](#and). The Lifecycle Rule will apply to any object matching all the predicates configured inside the `and` block.
* `object_size_greater_than` - (Optional) Minimum object size (in bytes) to which the rule applies.
* `object_size_less_than` - (Optional) Maximum object size (in bytes) to which the rule applies.
* `prefix` - (Optional) Prefix identifying one or more objects to which the rule applies. Defaults to an empty string (`""`) if not specified.
* `tag` - (Optional) Configuration block for specifying a tag key and value. [See below](#tag).

### noncurrent_version_expiration

The `noncurrent_version_expiration` configuration block supports the following arguments:

* `newer_noncurrent_versions` - (Optional) Number of noncurrent versions Amazon S3 will retain. Must be a non-zero positive integer.
* `noncurrent_days` - (Optional) Number of days an object is noncurrent before Amazon S3 can perform the associated action. Must be a positive integer.

### noncurrent_version_transition

The `noncurrent_version_transition` configuration block supports the following arguments:

* `newer_noncurrent_versions` - (Optional) Number of noncurrent versions Amazon S3 will retain. Must be a non-zero positive integer.
* `noncurrent_days` - (Optional) Number of days an object is noncurrent before Amazon S3 can perform the associated action.
* `storage_class` - (Required) Class of storage used to store the object. Valid Values: `GLACIER`, `STANDARD_IA`, `ONEZONE_IA`, `INTELLIGENT_TIERING`, `DEEP_ARCHIVE`, `GLACIER_IR`.

### transition

The `transition` configuration block supports the following arguments:

~> Only one of `date` or `days` should be specified. If neither are specified, the `transition` will default to 0 `days`.

* `date` - (Optional, Conflicts with `days`) Date objects are transitioned to the specified storage class. The date value must be in [RFC3339 full-date format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.6) e.g. `2023-08-22`.
* `days` - (Optional, Conflicts with `date`) Number of days after creation when objects are transitioned to the specified storage class. The value must be a positive integer. If both `days` and `date` are not specified, defaults to `0`. Valid values depend on `storage_class`, see [Transition objects using Amazon S3 Lifecycle](https://docs.aws.amazon.com/AmazonS3/latest/userguide/lifecycle-transition-general-considerations.html) for more details.
* `storage_class` - Class of storage used to store the object. Valid Values: `GLACIER`, `STANDARD_IA`, `ONEZONE_IA`, `INTELLIGENT_TIERING`, `DEEP_ARCHIVE`, `GLACIER_IR`.

### and

The `and` configuration block supports the following arguments:

* `object_size_greater_than` - (Optional) Minimum object size to which the rule applies. Value must be at least `0` if specified. Defaults to 128000 (128 KB) for all `storage_class` values unless `transition_default_minimum_object_size` specifies otherwise.
* `object_size_less_than` - (Optional) Maximum object size to which the rule applies. Value must be at least `1` if specified.
* `prefix` - (Optional) Prefix identifying one or more objects to which the rule applies.
* `tags` - (Optional) Key-value map of resource tags. All of these tags must exist in the object's tag set in order for the rule to apply.

### tag

The `tag` configuration block supports the following arguments:

* `key` - (Required) Name of the object key.
* `value` - (Required) Value of the tag.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The `bucket` or `bucket` and `expected_bucket_owner` separated by a comma (`,`) if the latter is provided.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an S3 bucket lifecycle configuration using the `bucket` or the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```terraform
import {
  to = aws_s3_bucket_lifecycle_configuration.example
  id = "bucket-name"
}
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```terraform
import {
  to = aws_s3_bucket_lifecycle_configuration.example
  id = "bucket-name,123456789012"
}
```

Using `terraform import`, import an S3 bucket lifecycle configuration using the `bucket` or the `bucket` and `expected_bucket_owner` separated by a comma (`,`). For example:

If the owner (account ID) of the source bucket is the same account used to configure the Terraform AWS Provider, import using the `bucket`:

```console
% terraform import aws_s3_bucket_lifecycle_configuration.example bucket-name
```

If the owner (account ID) of the source bucket differs from the account used to configure the Terraform AWS Provider, import using the `bucket` and `expected_bucket_owner` separated by a comma (`,`):

```console
% terraform import aws_s3_bucket_lifecycle_configuration.example bucket-name,123456789012
```
