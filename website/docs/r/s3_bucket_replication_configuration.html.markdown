---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_replication_configuration"
description: |-
  Provides a S3 bucket replication configuration resource.
---

# Resource: aws_s3_bucket_replication_configuration

Provides an independent configuration resource for S3 bucket [replication configuration](http://docs.aws.amazon.com/AmazonS3/latest/dev/crr.html).

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

### Bi-Directional Replication

```

...

resource "aws_s3_bucket" "east" {
  bucket = "tf-test-bucket-east-12345"

  versioning {
    enabled = true
  }

  lifecycle {
    ignore_changes = [
      replication_configuration
    ]
  }
}

resource "aws_s3_bucket" "west" {
  provider = west
  bucket   = "tf-test-bucket-west-12345"

  versioning {
    enabled = true
  }

  lifecycle {
    ignore_changes = [
      replication_configuration
    ]
  }
}

aws_s3_bucket_replication_configuration "east_to_west" {
  role   = aws_iam_role.east_replication.arn
  bucket = aws_s3_bucket.east.id 
  rules {
    id     = "foobar"
    prefix = "foo"
    status = "Enabled"

    destination {
      bucket        = aws_s3_bucket.west.arn
      storage_class = "STANDARD"
    }
  }
}

aws_s3_bucket_replication_configuration "west_to_east" {
  role   = aws_iam_role.west_replication.arn
  bucket = aws_s3_bucket.west.id 
  rules {
    id     = "foobar"
    prefix = "foo"
    status = "Enabled"

    destination {
      bucket        = aws_s3_bucket.east.arn
      storage_class = "STANDARD"
    }
  }
}
```

## Usage Notes

This resource implements the same features that are provided by the `replication_configuration` object of the `aws_s3_bucket` resource.  To avoid conflicts or unexpected apply results a lifecycle configuration is needed on the `aws_s3_bucket` to ignore changes to the internal `replication_configuration` object.  Faliure to add the `lifecycle` configuation to the `aws_s3_bucket` will result in conflicting state results.

~> **NOTE:** To avoid conflicts always add the following lifecycle object to the `aws_s3_bucket` resource of the source bucket.

```
lifecycle {
  ignore_changes = [
    replication_configuration
  ]
}
```
The `aws_s3_bucket_replication_configuration` resource provides the following features that are not available in the `aws_s3_bucket` resource:

* `replica_modifications` - Added to the `source_selection_criteria` configuration object [documented below](#source_selection_criteria)
* `metrics` - Added to the `destination` configuration object [documented below](#metrics)
* `replication_time` - Added to the `destination` configuration object [documented below](#replication_time)
* `existing_object_replication` - Added to the replication rule object [documented below](#existing_object_replication)

Replication for existing objects requires activation by AWS Support.  See [userguide/replication-what-is-isnot-replicated](https://docs.aws.amazon.com/AmazonS3/latest/userguide/replication-what-is-isnot-replicated.html#existing-object-replication)


## Argument Reference

The `replication_configuration` resource supports the following:

* `bucket` - (Required) The name of the source S3 bucket you want Amazon S3 to monitor.
* `role` - (Required) The ARN of the IAM role for Amazon S3 to assume when replicating the objects.
* `rules` - (Required) Specifies the rules managing the replication [documented below](#rules).

### rules 

~> **NOTE:** Replication to multiple destination buckets requires that `priority` is specified in the `rules` object. If the corresponding rule requires no filter, an empty configuration block `filter {}` must be specified.

~> **NOTE:** Amazon S3's latest version of the replication configuration is V2, which includes the `filter` attribute for replication rules.

The `rules` object supports the following:

With the `filter` attribute, you can specify object filters based on the object key prefix, tags, or both to scope the objects that the rule applies to.  Replication configuration V1 supports filtering based on only the `prefix` attribute. For backwards compatibility, Amazon S3 continues to support the V1 configuration.

* `existing_object_replication` - (Optional) Replicate existing objects in the source bucket according to the rule configurations [documented below](#existing_object_replication).
* `delete_marker_replication` - (Optional) Whether delete markers are replicated. This argument is only valid with V2 replication configurations (i.e., when `filter` is used)[documented below](#delete_marker_replication).
* `destination` - (Required) Specifies the destination for the rule [documented below](#destination).
* `filter` - (Optional, Conflicts with `prefix`) Filter that identifies subset of objects to which the replication rule applies [documented below](#filter).
* `id` - (Optional) Unique identifier for the rule. Must be less than or equal to 255 characters in length.
* `prefix` - (Optional, Conflicts with `filter`) Object keyname prefix identifying one or more objects to which the rule applies. Must be less than or equal to 1024 characters in length.
* `priority` - (Optional) The priority associated with the rule. Priority should only be set if `filter` is configured. If not provided, defaults to `0`. Priority must be unique between multiple rules.
* `source_selection_criteria` - (Optional) Specifies special object selection criteria [documented below](#source_selection_criteria).
* `status` - (Required) The status of the rule. Either `"Enabled"` or `"Disabled"`. The rule is ignored if status is not "Enabled".

### exiting_object_replication

~> **NOTE:** Replication for existing objects requires activation by AWS Support.  See [userguide/replication-what-is-isnot-replicated](https://docs.aws.amazon.com/AmazonS3/latest/userguide/replication-what-is-isnot-replicated.html#existing-object-replication)

The `existing_object_replication` object supports the following:

```
existing_object_replication {
  status = "Enabled"
}
```
* `status` - (Required) Whether the existing objects should be replicated. Either `"Enabled"` or `"Disabled"`. The object is ignored if status is not `"Enabled"`.


### delete_marker_replication

~> **NOTE:** This configuration format differes from that of `aws_s3_bucket`.

~> **NOTE:** This argument is only available with V2 replication configurations. 

The `delete_marker_replication` object supports the following:

```
delete_marker_replication {
  status = "Enabled"
}
```
* `status` - (Required) Whether delete markers should be replicated. Either `"Enabled"` or `"Disabled"`. The object is ignored if status is not `"Enabled"`.


### destination 
The `destination` object supports the following:

* `bucket` - (Required) The ARN of the S3 bucket where you want Amazon S3 to store replicas of the object identified by the rule.
* `storage_class` - (Optional) The class of storage used to store the object. Can be `STANDARD`, `REDUCED_REDUNDANCY`, `STANDARD_IA`, `ONEZONE_IA`, `INTELLIGENT_TIERING`, `GLACIER`, or `DEEP_ARCHIVE`.
* `replica_kms_key_id` - (Optional) Destination KMS encryption key ARN for SSE-KMS replication. Must be used in conjunction with
  `sse_kms_encrypted_objects` source selection criteria.
* `access_control_translation` - (Optional) Specifies the overrides to use for object owners on replication. Must be used in conjunction with `account_id` owner override configuration.
* `account_id` - (Optional) The Account ID to use for overriding the object owner on replication. Must be used in conjunction with `access_control_translation` override configuration.
* `replication_time` - (Optional) Replication Time Control must be used in conjunction with `metrics` [documented below](#replication_time).
* `metrics` - (Optional) Metrics must be used in conjunction with `replication_time` [documented below](#metrics).

### replication_time

```
replication_time {
  status = "Enabled"
  time {
    minutes = 15
  }
}
```

The `replication_time` object supports the following:

* `status` - (Required) The status of the Replication Time Control. Either `"Enabled"` or `"Disabled"`. The object is ignored if status is not `"Enabled"`.
* `time` - (Required) The replication time `minutes` to be configured.  The `minutes` value is expected to be an integer.

### metrics

```
metrics {
  status = "Enabled"
  event_threshold {
    minutes = 15
  }
}
```

The `metrics` object supports the following:

* `status` - (Required) The status of the Destination Metrics. Either `"Enabled"` or `"Disabled"`. The object is ignored if status is not `"Enabled"`.
* `event_threshold` - (Required) The time in `minutes` specifying the operation missed threshold event.  The `minutes` value is expected to be an integer.

### source_selection_criteria

The `source_selection_criteria` object supports the following:
```
source_selection_criteria {
  replica_modification {
    status = "Enabled"
  }
  sse_kms_encrypted_objects {
    status = "Enabled"
  }
}
```

* `replica_modifications` - (Optional) Keep object metadata such as tags, ACLs, and Object Lock settings replicated between 
   replicas and source objects. The `status` value is required to be either `"Enabled"` or `"Disabled"`. The object is ignored if status is not `"Enabled"`.

* `sse_kms_encrypted_objects` - (Optional) Match SSE-KMS encrypted objects (documented below). If specified, `replica_kms_key_id`
   in `destination` must be specified as well. The `status` value is required to be either `"Enabled"` or `"Disabled"`. The object is ignored if status is not `"Enabled"`. 

  ~> **NOTE:** `sse_kms_encrypted_objects` configuration format differs here from the configuration in the `aws_s3_bucket` resource.

### filter

The `filter` object supports the following:

* `prefix` - (Optional) Object keyname prefix that identifies subset of objects to which the rule applies. Must be less than or equal to 1024 characters in length.
* `tags` - (Optional)  A map of tags that identifies subset of objects to which the rule applies.
The rule applies only to objects having all the tags in its tagset.

## Import

S3 bucket replication configuration can be imported using the `bucket`, e.g.

```
$ terraform import aws_s3_bucket_replication_configuration.replication bucket-name
```
