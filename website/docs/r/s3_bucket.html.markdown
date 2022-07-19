---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket"
description: |-
  Provides a S3 bucket resource.
---

# Resource: aws_s3_bucket

Provides a S3 bucket resource.

-> This functionality is for managing S3 in an AWS Partition. To manage [S3 on Outposts](https://docs.aws.amazon.com/AmazonS3/latest/dev/S3onOutposts.html), see the [`aws_s3control_bucket`](/docs/providers/aws/r/s3control_bucket.html) resource.

~> **NOTE on S3 Bucket Accelerate Configuration:** S3 Bucket Accelerate can be configured in either the standalone resource [`aws_s3_bucket_accelerate_configuration`](s3_bucket_accelerate_configuration.html)
or with the deprecated parameter `acceleration_status` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket canned ACL Configuration:** S3 Bucket canned ACL can be configured in either the standalone resource [`aws_s3_bucket_acl`](s3_bucket_acl.html.markdown)
or with the deprecated parameter `acl` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket ACL Grants Configuration:** S3 Bucket grants can be configured in either the standalone resource [`aws_s3_bucket_acl`](s3_bucket_acl.html.markdown)
or with the deprecated parameter `grant` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket CORS Configuration:** S3 Bucket CORS can be configured in either the standalone resource [`aws_s3_bucket_cors_configuration`](s3_bucket_cors_configuration.html.markdown)
or with the deprecated parameter `cors_rule` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket Lifecycle Configuration:** S3 Bucket Lifecycle can be configured in either the standalone resource [`aws_s3_bucket_lifecycle_configuration`](s3_bucket_lifecycle_configuration.html)
or with the deprecated parameter `lifecycle_rule` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket Logging Configuration:** S3 Bucket logging can be configured in either the standalone resource [`aws_s3_bucket_logging`](s3_bucket_logging.html.markdown)
or with the deprecated parameter `logging` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket Object Lock Configuration:** S3 Bucket Object Lock can be configured in either the standalone resource [`aws_s3_bucket_object_lock_configuration`](s3_bucket_object_lock_configuration.html)
or with the deprecated parameter `object_lock_configuration` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket Policy Configuration:** S3 Bucket Policy can be configured in either the standalone resource [`aws_s3_bucket_policy`](s3_bucket_policy.html)
or with the deprecated parameter `policy` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket Replication Configuration:** S3 Bucket Replication can be configured in either the standalone resource [`aws_s3_bucket_replication_configuration`](s3_bucket_replication_configuration.html)
or with the deprecated parameter `replication_configuration` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket Request Payment Configuration:** S3 Bucket Request Payment can be configured in either the standalone resource [`aws_s3_bucket_request_payment_configuration`](s3_bucket_request_payment_configuration.html)
or with the deprecated parameter `request_payer` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket Server Side Encryption Configuration:** S3 Bucket Server Side Encryption can be configured in either the standalone resource [`aws_s3_bucket_server_side_encryption_configuration`](s3_bucket_server_side_encryption_configuration.html)
or with the deprecated parameter `server_side_encryption_configuration` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket Versioning Configuration:** S3 Bucket versioning can be configured in either the standalone resource [`aws_s3_bucket_versioning`](s3_bucket_versioning.html.markdown)
or with the deprecated parameter `versioning` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

~> **NOTE on S3 Bucket Website Configuration:** S3 Bucket Website can be configured in either the standalone resource [`aws_s3_bucket_website_configuration`](s3_bucket_website_configuration.html.markdown)
or with the deprecated parameter `website` in the resource `aws_s3_bucket`.
Configuring with both will cause inconsistencies and may overwrite configuration.

## Example Usage

### Private Bucket w/ Tags

```terraform
resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"

  tags = {
    Name        = "My bucket"
    Environment = "Dev"
  }
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.b.id
  acl    = "private"
}
```

### Static Website Hosting

-> **NOTE:** The parameter `website` is deprecated.
Use the resource [`aws_s3_bucket_website_configuration`](s3_bucket_website_configuration.html.markdown) instead.

```terraform
resource "aws_s3_bucket" "b" {
  bucket = "s3-website-test.hashicorp.com"
  acl    = "public-read"
  policy = file("policy.json")

  website {
    index_document = "index.html"
    error_document = "error.html"

    routing_rules = <<EOF
[{
    "Condition": {
        "KeyPrefixEquals": "docs/"
    },
    "Redirect": {
        "ReplaceKeyPrefixWith": "documents/"
    }
}]
EOF
  }
}
```

### Using CORS

-> **NOTE:** The parameter `cors_rule` is deprecated.
Use the resource [`aws_s3_bucket_cors_configuration`](s3_bucket_cors_configuration.html.markdown) instead.

```terraform
resource "aws_s3_bucket" "b" {
  bucket = "s3-website-test.hashicorp.com"
  acl    = "public-read"

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://s3-website-test.hashicorp.com"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}
```

### Using versioning

-> **NOTE:** The parameter `versioning` is deprecated.
Use the resource [`aws_s3_bucket_versioning`](s3_bucket_versioning.html.markdown) instead.

```terraform
resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  versioning {
    enabled = true
  }
}
```

### Enable Logging

-> **NOTE:** The parameter `logging` is deprecated.
Use the resource [`aws_s3_bucket_logging`](s3_bucket_logging.html.markdown) instead.

```terraform
resource "aws_s3_bucket" "log_bucket" {
  bucket = "my-tf-log-bucket"
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  logging {
    target_bucket = aws_s3_bucket.log_bucket.id
    target_prefix = "log/"
  }
}
```

### Using object lifecycle

-> **NOTE:** The parameter `lifecycle_rule` is deprecated.
Use the resource [`aws_s3_bucket_lifecycle_configuration`](s3_bucket_lifecycle_configuration.html) instead.

```terraform
resource "aws_s3_bucket" "bucket" {
  bucket = "my-bucket"
  acl    = "private"

  lifecycle_rule {
    id      = "log"
    enabled = true

    prefix = "log/"

    tags = {
      rule      = "log"
      autoclean = "true"
    }

    transition {
      days          = 30
      storage_class = "STANDARD_IA" # or "ONEZONE_IA"
    }

    transition {
      days          = 60
      storage_class = "GLACIER"
    }

    expiration {
      days = 90
    }
  }

  lifecycle_rule {
    id      = "tmp"
    prefix  = "tmp/"
    enabled = true

    expiration {
      date = "2016-01-12"
    }
  }
}

resource "aws_s3_bucket" "versioning_bucket" {
  bucket = "my-versioning-bucket"
  acl    = "private"

  versioning {
    enabled = true
  }

  lifecycle_rule {
    prefix  = "config/"
    enabled = true

    noncurrent_version_transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    noncurrent_version_transition {
      days          = 60
      storage_class = "GLACIER"
    }

    noncurrent_version_expiration {
      days = 90
    }
  }
}
```

### Using object lock configuration

-> **NOTE:** The parameter `object_lock_configuration` is deprecated.
To **enable** Object Lock on a **new** bucket, use the `object_lock_enabled` argument in **this** resource.
To configure the default retention rule of the Object Lock configuration use the resource [`aws_s3_bucket_object_lock_configuration` resource](s3_bucket_object_lock_configuration.html.markdown) instead.
To **enable** Object Lock on an **existing** bucket, please contact AWS Support and refer to the [Object lock configuration for an existing bucket](s3_bucket_object_lock_configuration.html.markdown#object-lock-configuration-for-an-existing-bucket) example for more details.

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "my-tf-example-bucket"

  object_lock_configuration {
    object_lock_enabled = "Enabled"

    rule {
      default_retention {
        mode = "COMPLIANCE"
        days = 5
      }
    }
  }
}
```

### Using replication configuration

-> **NOTE:** The parameter `replication_configuration` is deprecated.
Use the resource [`aws_s3_bucket_replication_configuration`](s3_bucket_replication_configuration.html) instead.

```terraform
provider "aws" {
  region = "eu-west-1"
}

provider "aws" {
  alias  = "central"
  region = "eu-central-1"
}

resource "aws_iam_role" "replication" {
  name = "tf-iam-role-replication-12345"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_policy" "replication" {
  name = "tf-iam-role-policy-replication-12345"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:GetReplicationConfiguration",
        "s3:ListBucket"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.source.arn}"
      ]
    },
    {
      "Action": [
        "s3:GetObjectVersionForReplication",
        "s3:GetObjectVersionAcl",
         "s3:GetObjectVersionTagging"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.source.arn}/*"
      ]
    },
    {
      "Action": [
        "s3:ReplicateObject",
        "s3:ReplicateDelete",
        "s3:ReplicateTags"
      ],
      "Effect": "Allow",
      "Resource": "${aws_s3_bucket.destination.arn}/*"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "replication" {
  role       = aws_iam_role.replication.name
  policy_arn = aws_iam_policy.replication.arn
}

resource "aws_s3_bucket" "destination" {
  bucket = "tf-test-bucket-destination-12345"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "source" {
  provider = aws.central
  bucket   = "tf-test-bucket-source-12345"
  acl      = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.replication.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      filter {
        tags = {}
      }
      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"

        replication_time {
          status  = "Enabled"
          minutes = 15
        }

        metrics {
          status  = "Enabled"
          minutes = 15
        }
      }
    }
  }
}
```

### Enable Default Server Side Encryption

-> **NOTE:** The parameter `server_side_encryption_configuration` is deprecated.
Use the resource [`aws_s3_bucket_server_side_encryption_configuration`](s3_bucket_server_side_encryption_configuration.html) instead.

```terraform
resource "aws_kms_key" "mykey" {
  description             = "This key is used to encrypt bucket objects"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "mybucket" {
  bucket = "mybucket"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.mykey.arn
        sse_algorithm     = "aws:kms"
      }
    }
  }
}
```

### Using ACL policy grants

-> **NOTE:** The parameters `acl` and `grant` are deprecated.
Use the resource [`aws_s3_bucket_acl`](s3_bucket_acl.html.markdown) instead.

```terraform
data "aws_canonical_user_id" "current_user" {}

resource "aws_s3_bucket" "bucket" {
  bucket = "mybucket"

  grant {
    id          = data.aws_canonical_user_id.current_user.id
    type        = "CanonicalUser"
    permissions = ["FULL_CONTROL"]
  }

  grant {
    type        = "Group"
    permissions = ["READ_ACP", "WRITE"]
    uri         = "http://acs.amazonaws.com/groups/s3/LogDelivery"
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Optional, Forces new resource) The name of the bucket. If omitted, Terraform will assign a random, unique name. Must be lowercase and less than or equal to 63 characters in length. A full list of bucket naming rules [may be found here](https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html).
* `bucket_prefix` - (Optional, Forces new resource) Creates a unique bucket name beginning with the specified prefix. Conflicts with `bucket`. Must be lowercase and less than or equal to 37 characters in length. A full list of bucket naming rules [may be found here](https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html).
* `acceleration_status` - (Optional, **Deprecated**) Sets the accelerate configuration of an existing bucket. Can be `Enabled` or `Suspended`. Cannot be used in `cn-north-1` or `us-gov-west-1`. Terraform will only perform drift detection if a configuration value is provided.
  Use the resource [`aws_s3_bucket_accelerate_configuration`](s3_bucket_accelerate_configuration.html) instead.
* `acl` - (Optional, **Deprecated**) The [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl) to apply. Valid values are `private`, `public-read`, `public-read-write`, `aws-exec-read`, `authenticated-read`, and `log-delivery-write`. Defaults to `private`.  Conflicts with `grant`. Terraform will only perform drift detection if a configuration value is provided. Use the resource [`aws_s3_bucket_acl`](s3_bucket_acl.html.markdown) instead.
* `grant` - (Optional, **Deprecated**) An [ACL policy grant](https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#sample-acl). See [Grant](#grant) below for details. Conflicts with `acl`. Terraform will only perform drift detection if a configuration value is provided. Use the resource [`aws_s3_bucket_acl`](s3_bucket_acl.html.markdown) instead.
* `cors_rule` - (Optional, **Deprecated**) A rule of [Cross-Origin Resource Sharing](https://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html). See [CORS rule](#cors-rule) below for details. Terraform will only perform drift detection if a configuration value is provided. Use the resource [`aws_s3_bucket_cors_configuration`](s3_bucket_cors_configuration.html.markdown) instead.
* `force_destroy` - (Optional, Default:`false`) A boolean that indicates all objects (including any [locked objects](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock-overview.html)) should be deleted from the bucket so that the bucket can be destroyed without error. These objects are *not* recoverable.
* `lifecycle_rule` - (Optional, **Deprecated**) A configuration of [object lifecycle management](http://docs.aws.amazon.com/AmazonS3/latest/dev/object-lifecycle-mgmt.html). See [Lifecycle Rule](#lifecycle-rule) below for details. Terraform will only perform drift detection if a configuration value is provided.
  Use the resource [`aws_s3_bucket_lifecycle_configuration`](s3_bucket_lifecycle_configuration.html) instead.
* `logging` - (Optional, **Deprecated**) A configuration of [S3 bucket logging](https://docs.aws.amazon.com/AmazonS3/latest/UG/ManagingBucketLogging.html) parameters. See [Logging](#logging) below for details. Terraform will only perform drift detection if a configuration value is provided.
  Use the resource [`aws_s3_bucket_logging`](s3_bucket_logging.html.markdown) instead.
* `object_lock_enabled` - (Optional, Default:`false`, Forces new resource) Indicates whether this bucket has an Object Lock configuration enabled. Valid values are `true` or `false`.
* `object_lock_configuration` - (Optional, **Deprecated**) A configuration of [S3 object locking](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock.html). See [Object Lock Configuration](#object-lock-configuration) below for details.
  Terraform wil only perform drift detection if a configuration value is provided.
  Use the `object_lock_enabled` parameter and the resource [`aws_s3_bucket_object_lock_configuration`](s3_bucket_object_lock_configuration.html.markdown) instead.
* `policy` - (Optional, **Deprecated**) A valid [bucket policy](https://docs.aws.amazon.com/AmazonS3/latest/dev/example-bucket-policies.html) JSON document. Note that if the policy document is not specific enough (but still valid), Terraform may view the policy as constantly changing in a `terraform plan`. In this case, please make sure you use the verbose/specific version of the policy. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).
  Terraform will only perform drift detection if a configuration value is provided.
  Use the resource [`aws_s3_bucket_policy`](s3_bucket_policy.html) instead.
* `replication_configuration` - (Optional, **Deprecated**) A configuration of [replication configuration](http://docs.aws.amazon.com/AmazonS3/latest/dev/crr.html). See [Replication Configuration](#replication-configuration) below for details. Terraform will only perform drift detection if a configuration value is provided.
  Use the resource [`aws_s3_bucket_replication_configuration`](s3_bucket_replication_configuration.html) instead.
* `request_payer` - (Optional, **Deprecated**) Specifies who should bear the cost of Amazon S3 data transfer.
  Can be either `BucketOwner` or `Requester`. By default, the owner of the S3 bucket would incur the costs of any data transfer.
  See [Requester Pays Buckets](http://docs.aws.amazon.com/AmazonS3/latest/dev/RequesterPaysBuckets.html) developer guide for more information.
  Terraform will only perform drift detection if a configuration value is provided.
  Use the resource [`aws_s3_bucket_request_payment_configuration`](s3_bucket_request_payment_configuration.html) instead.
* `server_side_encryption_configuration` - (Optional, **Deprecated**) A configuration of [server-side encryption configuration](http://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-encryption.html). See [Server Side Encryption Configuration](#server-side-encryption-configuration) below for details.
  Terraform will only perform drift detection if a configuration value is provided.
  Use the resource [`aws_s3_bucket_server_side_encryption_configuration`](s3_bucket_server_side_encryption_configuration.html) instead.
* `versioning` - (Optional, **Deprecated**) A configuration of the [S3 bucket versioning state](https://docs.aws.amazon.com/AmazonS3/latest/dev/Versioning.html). See [Versioning](#versioning) below for details. Terraform will only perform drift detection if a configuration value is provided. Use the resource [`aws_s3_bucket_versioning`](s3_bucket_versioning.html.markdown) instead.
* `website` - (Optional, **Deprecated**) A configuration of the [S3 bucket website](https://docs.aws.amazon.com/AmazonS3/latest/userguide/WebsiteHosting.html). See [Website](#website) below for details. Terraform will only perform drift detection if a configuration value is provided.
  Use the resource [`aws_s3_bucket_website_configuration`](s3_bucket_website_configuration.html.markdown) instead.
* `tags` - (Optional) A map of tags to assign to the bucket. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### CORS Rule

~> **NOTE:** Currently, changes to the `cors_rule` configuration of _existing_ resources cannot be automatically detected by Terraform. To manage changes of CORS rules to an S3 bucket, use the `aws_s3_bucket_cors_configuration` resource instead. If you use `cors_rule` on an `aws_s3_bucket`, Terraform will assume management over the full set of CORS rules for the S3 bucket, treating additional CORS rules as drift. For this reason, `cors_rule` cannot be mixed with the external `aws_s3_bucket_cors_configuration` resource for a given S3 bucket.

The `cors_rule` configuration block supports the following arguments:

* `allowed_headers` - (Optional) List of headers allowed.
* `allowed_methods` - (Required) One or more HTTP methods that you allow the origin to execute. Can be `GET`, `PUT`, `POST`, `DELETE` or `HEAD`.
* `allowed_origins` - (Required) One or more origins you want customers to be able to access the bucket from.
* `expose_headers` - (Optional) One or more headers in the response that you want customers to be able to access from their applications (for example, from a JavaScript `XMLHttpRequest` object).
* `max_age_seconds` - (Optional) Specifies time in seconds that browser can cache the response for a preflight request.

### Grant

~> **NOTE:** Currently, changes to the `grant` configuration of _existing_ resources cannot be automatically detected by Terraform. To manage changes of ACL grants to an S3 bucket, use the `aws_s3_bucket_acl` resource instead. If you use `grant` on an `aws_s3_bucket`, Terraform will assume management over the full set of ACL grants for the S3 bucket, treating additional ACL grants as drift. For this reason, `grant` cannot be mixed with the external `aws_s3_bucket_acl` resource for a given S3 bucket.

The `grant` configuration block supports the following arguments:

* `id` - (Optional) Canonical user id to grant for. Used only when `type` is `CanonicalUser`.
* `type` - (Required) Type of grantee to apply for. Valid values are `CanonicalUser` and `Group`. `AmazonCustomerByEmail` is not supported.
* `permissions` - (Required) List of permissions to apply for grantee. Valid values are `READ`, `WRITE`, `READ_ACP`, `WRITE_ACP`, `FULL_CONTROL`.
* `uri` - (Optional) Uri address to grant for. Used only when `type` is `Group`.

### Lifecycle Rule

~> **NOTE:** Currently, changes to the `lifecycle_rule` configuration of _existing_ resources cannot be automatically detected by Terraform. To manage changes of Lifecycle rules to an S3 bucket, use the `aws_s3_bucket_lifecycle_configuration` resource instead. If you use `lifecycle_rule` on an `aws_s3_bucket`, Terraform will assume management over the full set of Lifecycle rules for the S3 bucket, treating additional Lifecycle rules as drift. For this reason, `lifecycle_rule` cannot be mixed with the external `aws_s3_bucket_lifecycle_configuration` resource for a given S3 bucket.

~> **NOTE:** At least one of `abort_incomplete_multipart_upload_days`, `expiration`, `transition`, `noncurrent_version_expiration`, `noncurrent_version_transition` must be specified.

The `lifecycle_rule` configuration block supports the following arguments:

* `id` - (Optional) Unique identifier for the rule. Must be less than or equal to 255 characters in length.
* `prefix` - (Optional) Object key prefix identifying one or more objects to which the rule applies.
* `tags` - (Optional) Specifies object tags key and value.
* `enabled` - (Required) Specifies lifecycle rule status.
* `abort_incomplete_multipart_upload_days` (Optional) Specifies the number of days after initiating a multipart upload when the multipart upload must be completed.
* `expiration` - (Optional) Specifies a period in the object's expire. See [Expiration](#expiration) below for details.
* `transition` - (Optional) Specifies a period in the object's transitions. See [Transition](#transition) below for details.
* `noncurrent_version_expiration` - (Optional) Specifies when noncurrent object versions expire. See [Noncurrent Version Expiration](#noncurrent-version-expiration) below for details.
* `noncurrent_version_transition` - (Optional) Specifies when noncurrent object versions transitions. See [Noncurrent Version Transition](#noncurrent-version-transition) below for details.

### Expiration

The `expiration` configuration block supports the following arguments:

* `date` - (Optional) Specifies the date after which you want the corresponding action to take effect.
* `days` - (Optional) Specifies the number of days after object creation when the specific rule action takes effect.
* `expired_object_delete_marker` - (Optional) On a versioned bucket (versioning-enabled or versioning-suspended bucket), you can add this element in the lifecycle configuration to direct Amazon S3 to delete expired object delete markers. This cannot be specified with Days or Date in a Lifecycle Expiration Policy.

### Transition

The `transition` configuration block supports the following arguments:

* `date` - (Optional) Specifies the date after which you want the corresponding action to take effect.
* `days` - (Optional) Specifies the number of days after object creation when the specific rule action takes effect.
* `storage_class` - (Required) Specifies the Amazon S3 [storage class](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Transition.html#AmazonS3-Type-Transition-StorageClass) to which you want the object to transition.

### Noncurrent Version Expiration

The `noncurrent_version_expiration` configuration block supports the following arguments:

* `days` - (Required) Specifies the number of days noncurrent object versions expire.

### Noncurrent Version Transition

The `noncurrent_version_transition` configuration supports the following arguments:

* `days` - (Required) Specifies the number of days noncurrent object versions transition.
* `storage_class` - (Required) Specifies the Amazon S3 [storage class](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Transition.html#AmazonS3-Type-Transition-StorageClass) to which you want the object to transition.

### Logging

~> **NOTE:** Currently, changes to the `logging` configuration of _existing_ resources cannot be automatically detected by Terraform. To manage changes of logging parameters to an S3 bucket, use the `aws_s3_bucket_logging` resource instead. If you use `logging` on an `aws_s3_bucket`, Terraform will assume management over the full set of logging parameters for the S3 bucket, treating additional logging parameters as drift. For this reason, `logging` cannot be mixed with the external `aws_s3_bucket_logging` resource for a given S3 bucket.

The `logging` configuration block supports the following arguments:

* `target_bucket` - (Required) The name of the bucket that will receive the log objects.
* `target_prefix` - (Optional) To specify a key prefix for log objects.

### Object Lock Configuration

~> **NOTE:** You can only **enable** S3 Object Lock for **new** buckets. If you need to **enable** S3 Object Lock for an **existing** bucket, please contact AWS Support.
When you create a bucket with S3 Object Lock enabled, Amazon S3 automatically enables versioning for the bucket.
Once you create a bucket with S3 Object Lock enabled, you can't disable Object Lock or suspend versioning for the bucket.

~> **NOTE:** Currently, changes to the `object_lock_configuration` configuration of _existing_ resources cannot be automatically detected by Terraform. To manage changes of Object Lock settings to an S3 bucket, use the `aws_s3_bucket_object_lock_configuration` resource instead. If you use `object_lock_configuration` on an `aws_s3_bucket`, Terraform will assume management over the full set of Object Lock configuration parameters for the S3 bucket, treating additional Object Lock configuration parameters as drift. For this reason, `object_lock_configuration` cannot be mixed with the external `aws_s3_bucket_object_lock_configuration` resource for a given S3 bucket.

The `object_lock_configuration` configuration block supports the following arguments:

* `object_lock_enabled` - (Optional, **Deprecated**) Indicates whether this bucket has an Object Lock configuration enabled. Valid value is `Enabled`. Use the top-level argument `object_lock_enabled` instead.
* `rule` - (Optional) The Object Lock rule in place for this bucket ([documented below](#rule)).

#### Rule

The `rule` configuration block supports the following argument:

* `default_retention` - (Required) The default retention period that you want to apply to new objects placed in this bucket ([documented below](#default-retention)).

#### Default Retention

The `default_retention` configuration block supports the following arguments:

~> **NOTE:** Either `days` or `years` must be specified, but not both.

* `mode` - (Required) The default Object Lock retention mode you want to apply to new objects placed in this bucket. Valid values are `GOVERNANCE` and `COMPLIANCE`.
* `days` - (Optional) The number of days that you want to specify for the default retention period.
* `years` - (Optional) The number of years that you want to specify for the default retention period.

### Replication Configuration

~> **NOTE:** Currently, changes to the `replication_configuration` configuration of _existing_ resources cannot be automatically detected by Terraform. To manage replication configuration changes to an S3 bucket, use the `aws_s3_bucket_replication_configuration` resource instead. If you use `replication_configuration` on an `aws_s3_bucket`, Terraform will assume management over the full replication configuration for the S3 bucket, treating additional replication configuration rules as drift. For this reason, `replication_configuration` cannot be mixed with the external `aws_s3_bucket_replication_configuration` resource for a given S3 bucket.

The `replication_configuration` configuration block supports the following arguments:

* `role` - (Required) The ARN of the IAM role for Amazon S3 to assume when replicating the objects.
* `rules` - (Required) Specifies the rules managing the replication ([documented below](#rules)).

#### Rules

The `rules` configuration block supports the following arguments:

~> **NOTE:** Amazon S3's latest version of the replication configuration is V2, which includes the `filter` attribute for replication rules.
With the `filter` attribute, you can specify object filters based on the object key prefix, tags, or both to scope the objects that the rule applies to.
Replication configuration V1 supports filtering based on only the `prefix` attribute. For backwards compatibility, Amazon S3 continues to support the V1 configuration.

* `delete_marker_replication_status` - (Optional) Whether delete markers are replicated. The only valid value is `Enabled`. To disable, omit this argument. This argument is only valid with V2 replication configurations (i.e., when `filter` is used).
* `destination` - (Required) Specifies the destination for the rule ([documented below](#destination)).
* `filter` - (Optional, Conflicts with `prefix`) Filter that identifies subset of objects to which the replication rule applies ([documented below](#filter)).
* `id` - (Optional) Unique identifier for the rule. Must be less than or equal to 255 characters in length.
* `prefix` - (Optional, Conflicts with `filter`) Object keyname prefix identifying one or more objects to which the rule applies. Must be less than or equal to 1024 characters in length.
* `priority` - (Optional) The priority associated with the rule. Priority should only be set if `filter` is configured. If not provided, defaults to `0`. Priority must be unique between multiple rules.
* `source_selection_criteria` - (Optional) Specifies special object selection criteria ([documented below](#source-selection-criteria)).
* `status` - (Required) The status of the rule. Either `Enabled` or `Disabled`. The rule is ignored if status is not Enabled.

#### Filter

The `filter` configuration block supports the following arguments:

* `prefix` - (Optional) Object keyname prefix that identifies subset of objects to which the rule applies. Must be less than or equal to 1024 characters in length.
* `tags` - (Optional)  A map of tags that identifies subset of objects to which the rule applies.
  The rule applies only to objects having all the tags in its tagset.

#### Destination

~> **NOTE:** Replication to multiple destination buckets requires that `priority` is specified in the `rules` object. If the corresponding rule requires no filter, an empty configuration block `filter {}` must be specified.

The `destination` configuration block supports the following arguments:

* `bucket` - (Required) The ARN of the S3 bucket where you want Amazon S3 to store replicas of the object identified by the rule.
* `storage_class` - (Optional) The [storage class](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Destination.html#AmazonS3-Type-Destination-StorageClass) used to store the object. By default, Amazon S3 uses the storage class of the source object to create the object replica.
* `replica_kms_key_id` - (Optional) Destination KMS encryption key ARN for SSE-KMS replication. Must be used in conjunction with
  `sse_kms_encrypted_objects` source selection criteria.
* `access_control_translation` - (Optional) Specifies the overrides to use for object owners on replication. Must be used in conjunction with `account_id` owner override configuration.
* `account_id` - (Optional) The Account ID to use for overriding the object owner on replication. Must be used in conjunction with `access_control_translation` override configuration.
* `replication_time` - (Optional) Enables S3 Replication Time Control (S3 RTC) ([documented below](#replication-time)).
* `metrics` - (Optional) Enables replication metrics (required for S3 RTC) ([documented below](#metrics)).

#### Replication Time

The `replication_time` configuration block supports the following arguments:

* `status` - (Optional) The status of RTC. Either `Enabled` or `Disabled`.
* `minutes` - (Optional) Threshold within which objects are to be replicated. The only valid value is `15`.

#### Metrics

The `metrics` configuration block supports the following arguments:

* `status` - (Optional) The status of replication metrics. Either `Enabled` or `Disabled`.
* `minutes` - (Optional) Threshold within which objects are to be replicated. The only valid value is `15`.

#### Source Selection Criteria

The `source_selection_criteria` configuration block supports the following argument:

* `sse_kms_encrypted_objects` - (Optional) Match SSE-KMS encrypted objects ([documented below](#sse-kms-encrypted-objects)). If specified, `replica_kms_key_id`
  in `destination` must be specified as well.

#### SSE KMS Encrypted Objects

The `sse_kms_encrypted_objects` configuration block supports the following argument:

* `enabled` - (Required) Boolean which indicates if this criteria is enabled.

### Server Side Encryption Configuration

~> **NOTE:** Currently, changes to the `server_side_encryption_configuration` configuration of _existing_ resources cannot be automatically detected by Terraform. To manage changes in encryption of an S3 bucket, use the `aws_s3_bucket_server_side_encryption_configuration` resource instead. If you use `server_side_encryption_configuration` on an `aws_s3_bucket`, Terraform will assume management over the encryption configuration for the S3 bucket, treating additional encryption changes as drift. For this reason, `server_side_encryption_configuration` cannot be mixed with the external `aws_s3_bucket_server_side_encryption_configuration` resource for a given S3 bucket.

The `server_side_encryption_configuration` configuration block supports the following argument:

* `rule` - (Required) A single object for server-side encryption by default configuration. (documented below)

The `rule` configuration block supports the following arguments:

* `apply_server_side_encryption_by_default` - (Required) A single object for setting server-side encryption by default. (documented below)
* `bucket_key_enabled` - (Optional) Whether or not to use [Amazon S3 Bucket Keys](https://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-key.html) for SSE-KMS.

The `apply_server_side_encryption_by_default` configuration block supports the following arguments:

* `sse_algorithm` - (Required) The server-side encryption algorithm to use. Valid values are `AES256` and `aws:kms`
* `kms_master_key_id` - (Optional) The AWS KMS master key ID used for the SSE-KMS encryption. This can only be used when you set the value of `sse_algorithm` as `aws:kms`. The default `aws/s3` AWS KMS master key is used if this element is absent while the `sse_algorithm` is `aws:kms`.

### Versioning

~> **NOTE:** Currently, changes to the `versioning` configuration of _existing_ resources cannot be automatically detected by Terraform. To manage changes of versioning state to an S3 bucket, use the `aws_s3_bucket_versioning` resource instead. If you use `versioning` on an `aws_s3_bucket`, Terraform will assume management over the versioning state of the S3 bucket, treating additional versioning state changes as drift. For this reason, `versioning` cannot be mixed with the external `aws_s3_bucket_versioning` resource for a given S3 bucket.

The `versioning` configuration block supports the following arguments:

* `enabled` - (Optional) Enable versioning. Once you version-enable a bucket, it can never return to an unversioned state. You can, however, suspend versioning on that bucket.
* `mfa_delete` - (Optional) Enable MFA delete for either `Change the versioning state of your bucket` or `Permanently delete an object version`. Default is `false`. This cannot be used to toggle this setting but is available to allow managed buckets to reflect the state in AWS

### Website

~> **NOTE:** Currently, changes to the `website` configuration of _existing_ resources cannot be automatically detected by Terraform. To manage changes to the website configuration of an S3 bucket, use the `aws_s3_bucket_website_configuration` resource instead. If you use `website` on an `aws_s3_bucket`, Terraform will assume management over the configuration of the website of the S3 bucket, treating additional website configuration changes as drift. For this reason, `website` cannot be mixed with the external `aws_s3_bucket_website_configuration` resource for a given S3 bucket.

The `website` configuration block supports the following arguments:

* `index_document` - (Required, unless using `redirect_all_requests_to`) Amazon S3 returns this index document when requests are made to the root domain or any of the subfolders.
* `error_document` - (Optional) An absolute path to the document to return in case of a 4XX error.
* `redirect_all_requests_to` - (Optional) A hostname to redirect all website requests for this bucket to. Hostname can optionally be prefixed with a protocol (`http://` or `https://`) to use when redirecting requests. The default is the protocol that is used in the original request.
* `routing_rules` - (Optional) A json array containing [routing rules](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-websiteconfiguration-routingrules.html)
  describing redirect behavior and when redirects are applied.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the bucket.
* `arn` - The ARN of the bucket. Will be of format `arn:aws:s3:::bucketname`.
* `bucket_domain_name` - The bucket domain name. Will be of format `bucketname.s3.amazonaws.com`.
* `bucket_regional_domain_name` - The bucket region-specific domain name. The bucket domain name including the region name, please refer [here](https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region) for format. Note: The AWS CloudFront allows specifying S3 region-specific endpoint when creating S3 origin, it will prevent [redirect issues](https://forums.aws.amazon.com/thread.jspa?threadID=216814) from CloudFront to S3 Origin URL.
* `hosted_zone_id` - The [Route 53 Hosted Zone ID](https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_website_region_endpoints) for this bucket's region.
* `region` - The AWS region this bucket resides in.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `website_endpoint` - The website endpoint, if the bucket is configured with a website. If not, this will be an empty string.
* `website_domain` - The domain of the website endpoint, if the bucket is configured with a website. If not, this will be an empty string. This is used to create Route 53 alias records.

## Import

S3 bucket can be imported using the `bucket`, e.g.,

```
$ terraform import aws_s3_bucket.bucket bucket-name
```
