---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket"
description: |-
  Provides a S3 bucket resource.
---

# Resource: aws_s3_bucket

Provides a S3 bucket resource.

-> This functionality is for managing S3 in an AWS Partition. To manage [S3 on Outposts](https://docs.aws.amazon.com/AmazonS3/latest/dev/S3onOutposts.html), see the [`aws_s3control_bucket`](/docs/providers/aws/r/s3control_bucket.html) resource.

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

The `website` argument is read-only as of version 4.0 of the Terraform AWS Provider.
See the [`aws_s3_bucket_website_configuration` resource](s3_bucket_website_configuration.html.markdown) for configuration details.

### Using CORS

The `cors_rule` argument is read-only as of version 4.0 of the Terraform AWS Provider.
See the [`aws_s3_bucket_cors_configuration` resource](s3_bucket_cors_configuration.html.markdown) for configuration details.

### Using versioning

```terraform
resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.b.id
  acl    = "private"
}
```

### Enable Logging

The `logging` argument is read-only as of version 4.0 of the Terraform AWS Provider.
See the [`aws_s3_bucket_logging` resource](s3_bucket_logging.html.markdown) for configuration details.

### Using object lifecycle

```terraform
resource "aws_s3_bucket" "bucket" {
  bucket = "my-bucket"

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

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}

resource "aws_s3_bucket" "versioning_bucket" {
  bucket = "my-versioning-bucket"

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

resource "aws_s3_bucket_acl" "versioning_bucket_acl" {
  bucket = aws_s3_bucket.versioning_bucket.id
  acl    = "private"
}
```

### Using replication configuration

The `replication_configuration` argument is read-only as of version 4.0 of the Terraform AWS Provider.
See the [`aws_s3_bucket_replication_configuration` resource](s3_bucket_replication_configuration.html.markdown) for configuration details.

### Enable Default Server Side Encryption

The `server_side_encryption_configuration` argument is read-only as of version 4.0.
See the [`aws_s3_bucket_server_side_encryption_configuration` resource](s3_bucket_server_side_encryption_configuration.html.markdown) for configuration details.

### Using ACL policy grants

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
* `acl` - (Optional) The [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl) to apply. Valid values are `private`, `public-read`, `public-read-write`, `aws-exec-read`, `authenticated-read`, and `log-delivery-write`. Defaults to `private`.  Conflicts with `grant`.
* `grant` - (Optional) An [ACL policy grant](https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#sample-acl) (documented below). Conflicts with `acl`.
* `policy` - (Optional) A valid [bucket policy](https://docs.aws.amazon.com/AmazonS3/latest/dev/example-bucket-policies.html) JSON document. Note that if the policy document is not specific enough (but still valid), Terraform may view the policy as constantly changing in a `terraform plan`. In this case, please make sure you use the verbose/specific version of the policy. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).

* `tags` - (Optional) A map of tags to assign to the bucket. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `force_destroy` - (Optional, Default:`false`) A boolean that indicates all objects (including any [locked objects](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock-overview.html)) should be deleted from the bucket so that the bucket can be destroyed without error. These objects are *not* recoverable.
* `versioning` - (Optional) A state of [versioning](https://docs.aws.amazon.com/AmazonS3/latest/dev/Versioning.html) (documented below)
* `lifecycle_rule` - (Optional) A configuration of [object lifecycle management](http://docs.aws.amazon.com/AmazonS3/latest/dev/object-lifecycle-mgmt.html) (documented below).
* `object_lock_configuration` - (Optional) A configuration of [S3 object locking](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock.html) (documented below)

The `versioning` object supports the following:

* `enabled` - (Optional) Enable versioning. Once you version-enable a bucket, it can never return to an unversioned state. You can, however, suspend versioning on that bucket.
* `mfa_delete` - (Optional) Enable MFA delete for either `Change the versioning state of your bucket` or `Permanently delete an object version`. Default is `false`. This cannot be used to toggle this setting but is available to allow managed buckets to reflect the state in AWS

The `lifecycle_rule` object supports the following:

* `id` - (Optional) Unique identifier for the rule. Must be less than or equal to 255 characters in length.
* `prefix` - (Optional) Object key prefix identifying one or more objects to which the rule applies.
* `tags` - (Optional) Specifies object tags key and value.
* `enabled` - (Required) Specifies lifecycle rule status.
* `abort_incomplete_multipart_upload_days` (Optional) Specifies the number of days after initiating a multipart upload when the multipart upload must be completed.
* `expiration` - (Optional) Specifies a period in the object's expire (documented below).
* `transition` - (Optional) Specifies a period in the object's transitions (documented below).
* `noncurrent_version_expiration` - (Optional) Specifies when noncurrent object versions expire (documented below).
* `noncurrent_version_transition` - (Optional) Specifies when noncurrent object versions transitions (documented below).

At least one of `abort_incomplete_multipart_upload_days`, `expiration`, `transition`, `noncurrent_version_expiration`, `noncurrent_version_transition` must be specified.

The `expiration` object supports the following

* `date` (Optional) Specifies the date after which you want the corresponding action to take effect.
* `days` (Optional) Specifies the number of days after object creation when the specific rule action takes effect.
* `expired_object_delete_marker` (Optional) On a versioned bucket (versioning-enabled or versioning-suspended bucket), you can add this element in the lifecycle configuration to direct Amazon S3 to delete expired object delete markers. This cannot be specified with Days or Date in a Lifecycle Expiration Policy.

The `transition` object supports the following

* `date` (Optional) Specifies the date after which you want the corresponding action to take effect.
* `days` (Optional) Specifies the number of days after object creation when the specific rule action takes effect.
* `storage_class` (Required) Specifies the Amazon S3 [storage class](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Transition.html#AmazonS3-Type-Transition-StorageClass) to which you want the object to transition.

The `noncurrent_version_expiration` object supports the following

* `days` (Required) Specifies the number of days noncurrent object versions expire.

The `noncurrent_version_transition` object supports the following

* `days` (Required) Specifies the number of days noncurrent object versions transition.
* `storage_class` (Required) Specifies the Amazon S3 [storage class](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Transition.html#AmazonS3-Type-Transition-StorageClass) to which you want the object to transition.

The `grant` object supports the following:

* `id` - (optional) Canonical user id to grant for. Used only when `type` is `CanonicalUser`.
* `type` - (required) - Type of grantee to apply for. Valid values are `CanonicalUser` and `Group`. `AmazonCustomerByEmail` is not supported.
* `permissions` - (required) List of permissions to apply for grantee. Valid values are `READ`, `WRITE`, `READ_ACP`, `WRITE_ACP`, `FULL_CONTROL`.
* `uri` - (optional) Uri address to grant for. Used only when `type` is `Group`.

The `object_lock_configuration` object supports the following:

* `object_lock_enabled` - (Required) Indicates whether this bucket has an Object Lock configuration enabled. Valid value is `Enabled`.
* `rule` - (Optional) The Object Lock rule in place for this bucket.

The `rule` object supports the following:

* `default_retention` - (Required) The default retention period that you want to apply to new objects placed in this bucket.

The `default_retention` object supports the following:

* `mode` - (Required) The default Object Lock retention mode you want to apply to new objects placed in this bucket. Valid values are `GOVERNANCE` and `COMPLIANCE`.
* `days` - (Optional) The number of days that you want to specify for the default retention period.
* `years` - (Optional) The number of years that you want to specify for the default retention period.

Either `days` or `years` must be specified, but not both.

~> **NOTE on `object_lock_configuration`:** You can only enable S3 Object Lock for new buckets. If you need to turn on S3 Object Lock for an existing bucket, please contact AWS Support.
When you create a bucket with S3 Object Lock enabled, Amazon S3 automatically enables versioning for the bucket.
Once you create a bucket with S3 Object Lock enabled, you can't disable Object Lock or suspend versioning for the bucket.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the bucket.
* `acceleration_status` - (Optional) The accelerate configuration status of the bucket. Not available in `cn-north-1` or `us-gov-west-1`.
* `arn` - The ARN of the bucket. Will be of format `arn:aws:s3:::bucketname`.
* `bucket_domain_name` - The bucket domain name. Will be of format `bucketname.s3.amazonaws.com`.
* `bucket_regional_domain_name` - The bucket region-specific domain name. The bucket domain name including the region name, please refer [here](https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region) for format. Note: The AWS CloudFront allows specifying S3 region-specific endpoint when creating S3 origin, it will prevent [redirect issues](https://forums.aws.amazon.com/thread.jspa?threadID=216814) from CloudFront to S3 Origin URL.
* `cors_rule` - Set of origins and methods ([cross-origin](https://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html) access allowed).
    * `allowed_headers` - Set of headers that are specified in the Access-Control-Request-Headers header.
    * `allowed_methods` - Set of HTTP methods that the origin is allowed to execute.
    * `allowed_origins` - Set of origins customers are able to access the bucket from.
    * `expose_headers` - Set of headers in the response that customers are able to access from their applications.
    * `max_age_seconds` The time in seconds that browser can cache the response for a preflight request.
* `hosted_zone_id` - The [Route 53 Hosted Zone ID](https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_website_region_endpoints) for this bucket's region.
* `logging` - The [logging parameters](https://docs.aws.amazon.com/AmazonS3/latest/UG/ManagingBucketLogging.html) for the bucket.
    * `target_bucket` - The name of the bucket that receives the log objects.
    * `target_prefix` - The prefix for all log object keys/
* `region` - The AWS region this bucket resides in.
* `replication_configuration` - The [replication configuration](http://docs.aws.amazon.com/AmazonS3/latest/dev/crr.html).
    * `role` - The ARN of the IAM role for Amazon S3 assumed when replicating the objects.
    * `rules` - The rules managing the replication.
        * `delete_marker_replication_status` - Whether delete markers are replicated.
        * `destination` - The destination for the rule.
            * `access_control_translation` - The overrides to use for object owners on replication.
                * `owner` - The override value for the owner on replicated objects.
            * `account_id` - The Account ID to use for overriding the object owner on replication.
            * `bucket` - The ARN of the S3 bucket where Amazon S3 stores replicas of the object identified by the rule.
            * `metrics` - Replication metrics.
                * `status` - The status of replication metrics.
                * `minutes` - Threshold within which objects are replicated.
            * `storage_class` - The [storage class](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Destination.html#AmazonS3-Type-Destination-StorageClass) used to store the object.
            * `replica_kms_key_id` - Destination KMS encryption key ARN for SSE-KMS replication.
            * `replication_time` - S3 Replication Time Control (S3 RTC).
                * `status` - The status of RTC.
                * `minutes` - Threshold within which objects are to be replicated.
        * `filter` - Filter that identifies subset of objects to which the replication rule applies.
            * `prefix` - Object keyname prefix that identifies subset of objects to which the rule applies.
            * `tags` - Map of tags that identifies subset of objects to which the rule applies.
        * `id` - Unique identifier for the rule.
        * `prefix` - Object keyname prefix identifying one or more objects to which the rule applies
        * `priority` - The priority associated with the rule.
        * `source_selection_criteria` - The special object selection criteria.
            * `sse_kms_encrypted_objects` - Matched SSE-KMS encrypted objects.
                * `enabled` - Whether this criteria is enabled.
        * `status` - The status of the rule.
* `request_payer` - Either `BucketOwner` or `Requester` that pays for the download and request fees.
* `server_side_encryption_configuration` - The [server-side encryption configuration](http://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-encryption.html).
    * `rule` - (required) Information about a particular server-side encryption configuration rule.
        * `apply_server_side_encryption_by_default` - The default server-side encryption applied to new objects in the bucket.
            * `kms_master_key_id` - (optional) The AWS KMS master key ID used for the SSE-KMS encryption.
            * `sse_algorithm` - (required) The server-side encryption algorithm used.
        * `bucket_key_enabled` - (Optional) Whether an [Amazon S3 Bucket Key](https://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-key.html) is used for SSE-KMS.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `website` - The website configuration, if configured.
    * `error_document` - The name of the error document for the website.
    * `index_document` - The name of the index document for the website.
    * `redirect_all_requests_to` - The redirect behavior for every request to this bucket's website endpoint.
    * `routing_rules` - (Optional) The rules that define when a redirect is applied and the redirect behavior.
* `website_endpoint` - The website endpoint, if the bucket is configured with a website. If not, this will be an empty string.
* `website_domain` - The domain of the website endpoint, if the bucket is configured with a website. If not, this will be an empty string. This is used to create Route 53 alias records.

## Import

S3 bucket can be imported using the `bucket`, e.g.,

```
$ terraform import aws_s3_bucket.bucket bucket-name
```

The `policy` argument is not imported and will be deprecated in a future version 3.x of the Terraform AWS Provider for removal in version 4.0. Use the [`aws_s3_bucket_policy` resource](/docs/providers/aws/r/s3_bucket_policy.html) to manage the S3 Bucket Policy instead.
