---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_replication_configuration"
description: |-
  Provides a S3 bucket replication configuration resource.
---

# Resource: aws_s3_bucket_replication_configuration

Provides a configuration of [replication configuration](http://docs.aws.amazon.com/AmazonS3/latest/dev/crr.html) for existing s3 buckets.

## Example Usage

### Using replication configuration

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
  lifecycle {
    ignore_changes = [
      replication_configuration
    ]
  }
}

aws_s3_bucket_replication_configuration replication {
  role   = aws_iam_role.replication.arn
  bucket = aws_s3_bucket.source.id 
  rules {
    id     = "foobar"
    prefix = "foo"
    status = "Enabled"

    destination {
      bucket        = aws_s3_bucket.destination.arn
      storage_class = "STANDARD"
    }
  }
}

```

~> **NOTE:** To avoid conflicts always add the following lifecycle block to the `aws_s3_bucket` resource of the source bucket.

```
lifecycle {
  ignore_changes = [
    replication_configuration
  ]
}
```


## Argument Reference

The following arguments are supported:

The `replication_configuration` object supports the following:

* `bucket` - (Required) The ARN of the source S3 bucket where you want Amazon S3 to monitor.
* `role` - (Required) The ARN of the IAM role for Amazon S3 to assume when replicating the objects.
* `rules` - (Required) Specifies the rules managing the replication (documented below).

The `rules` object supports the following:

~> **NOTE:** Amazon S3's latest version of the replication configuration is V2, which includes the `filter` attribute for replication rules.
With the `filter` attribute, you can specify object filters based on the object key prefix, tags, or both to scope the objects that the rule applies to.
Replication configuration V1 supports filtering based on only the `prefix` attribute. For backwards compatibility, Amazon S3 continues to support the V1 configuration.

* `existing_object_replication` - (Optional) Replicate existing objects in the source bucket according to the rule configurations (documented below).
* `delete_marker_replication_status` - (Optional) Whether delete markers are replicated. The only valid value is `Enabled`. To disable, omit this argument. This argument is only valid with V2 replication configurations (i.e., when `filter` is used).
* `destination` - (Required) Specifies the destination for the rule (documented below).
* `filter` - (Optional, Conflicts with `prefix`) Filter that identifies subset of objects to which the replication rule applies (documented below).
* `id` - (Optional) Unique identifier for the rule. Must be less than or equal to 255 characters in length.
* `prefix` - (Optional, Conflicts with `filter`) Object keyname prefix identifying one or more objects to which the rule applies. Must be less than or equal to 1024 characters in length.
* `priority` - (Optional) The priority associated with the rule. Priority should only be set if `filter` is configured. If not provided, defaults to `0`. Priority must be unique between multiple rules.
* `source_selection_criteria` - (Optional) Specifies special object selection criteria (documented below).
* `status` - (Required) The status of the rule. Either `Enabled` or `Disabled`. The rule is ignored if status is not Enabled.

~> **NOTE:** Replication to multiple destination buckets requires that `priority` is specified in the `rules` object. If the corresponding rule requires no filter, an empty configuration block `filter {}` must be specified.

The `existing_object_replication` object supports the following:

* `status` - (Required) Whether the existing objects should be replicated. Either `Enabled` or `Disabled`. The object is ignored if status is not Enabled.

The `destination` object supports the following:

* `bucket` - (Required) The ARN of the S3 bucket where you want Amazon S3 to store replicas of the object identified by the rule.
* `storage_class` - (Optional) The class of storage used to store the object. Can be `STANDARD`, `REDUCED_REDUNDANCY`, `STANDARD_IA`, `ONEZONE_IA`, `INTELLIGENT_TIERING`, `GLACIER`, or `DEEP_ARCHIVE`.
* `replica_kms_key_id` - (Optional) Destination KMS encryption key ARN for SSE-KMS replication. Must be used in conjunction with
  `sse_kms_encrypted_objects` source selection criteria.
* `access_control_translation` - (Optional) Specifies the overrides to use for object owners on replication. Must be used in conjunction with `account_id` owner override configuration.
* `account_id` - (Optional) The Account ID to use for overriding the object owner on replication. Must be used in conjunction with `access_control_translation` override configuration.
* `replication_time` - (Optional) Must be used in conjunction with `metrics` (documented below).
* `metrics` - (Optional) Must be used in conjunction with `replication_time` (documented below).

The `replication_time` object supports the following:

* `status` - (Required) The status of the Replica Modifications sync. Either `Enabled` or `Disabled`. The object is ignored if status is not Enabled.

The `metrics` object supports the following:

* `status` - (Required) The status of the Replica Modifications sync. Either `Enabled` or `Disabled`. The object is ignored if status is not Enabled.

The `source_selection_criteria` object supports the following:

* `replica_modifications` - (Optional) Keep object metadata such as tags, ACLs, and Object Lock settings replicated between 
   replicas and source objects (documented below).
  
* `sse_kms_encrypted_objects` - (Optional) Match SSE-KMS encrypted objects (documented below). If specified, `replica_kms_key_id`
   in `destination` must be specified as well.

The `replica_modifications` object supports the following:

* `status` - (Required) The status of the Replica Modifications sync. Either `Enabled` or `Disabled`. The object is ignored if status is not Enabled.

The `sse_kms_encrypted_objects` object supports the following:

* `status` - (Required) The status of the SSE KMS encryption. Either `Enabled` or `Disabled`. The object is ignored if status is not Enabled.

The `filter` object supports the following:

* `prefix` - (Optional) Object keyname prefix that identifies subset of objects to which the rule applies. Must be less than or equal to 1024 characters in length.
* `tags` - (Optional)  A map of tags that identifies subset of objects to which the rule applies.
The rule applies only to objects having all the tags in its tagset.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

## Import

S3 bucket replication configuration can be imported using the `bucket`, e.g.

```
$ terraform import aws_s3_bucket_replication_configuration.replication bucket-name
```
