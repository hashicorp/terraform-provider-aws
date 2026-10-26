---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_replication_configuration"
description: |-
  Terraform data source for managing an AWS S3 (Simple Storage) Bucket Replication Configuration.
---

# Data Source: aws_s3_bucket_replication_configuration

Terraform data source for managing an AWS S3 (Simple Storage) Bucket Replication Configuration.

## Example Usage

### Basic Usage

```terraform
data "aws_s3_bucket_replication_configuration" "example" {
  bucket = "example-bucket"
}
```

## Argument Reference

This data source supports the following arguments:

* `bucket` - (Required) Name of the bucket to get the replication configuration for.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `role` - ARN of the IAM role that Amazon S3 assumes when replicating objects.
* `rule` - List of configuration blocks that define the rules managing replication. See [`rule` Block](#rule-block) below.

### `rule` Block

* `delete_marker_replication` - Configuration block that specifies whether delete markers are replicated. See [`delete_marker_replication` Block](#delete_marker_replication-block) below.
* `destination` - Configuration block that specifies the destination for the rule. See [`destination` Block](#destination-block) below.
* `existing_object_replication` - Configuration block that specifies replication of existing objects. See [`existing_object_replication` Block](#existing_object_replication-block) below.
* `filter` - Configuration block that identifies the subset of objects to which the rule applies. See [`filter` Block](#filter-block) below.
* `id` - Unique identifier for the rule.
* `prefix` - Object key name prefix that identifies one or more objects to which the rule applies.
* `priority` - Priority associated with the rule.
* `source_selection_criteria` - Configuration block that specifies special object selection criteria. See [`source_selection_criteria` Block](#source_selection_criteria-block) below.
* `status` - Status of the rule.

### `delete_marker_replication` Block

* `status` - Whether delete markers are replicated.

### `destination` Block

* `access_control_translation` - Configuration block that specifies the overrides to use for object owners on replication. See [`access_control_translation` Block](#access_control_translation-block) below.
* `account` - Account ID used to specify the replica ownership.
* `bucket` - ARN of the bucket where Amazon S3 stores the results.
* `encryption_configuration` - Configuration block that provides information about encryption. See [`encryption_configuration` Block](#encryption_configuration-block) below.
* `metrics` - Configuration block that specifies replication metrics-related settings. See [`metrics` Block](#metrics-block) below.
* `replication_time` - Configuration block that specifies S3 Replication Time Control (S3 RTC). See [`replication_time` Block](#replication_time-block) below.
* `storage_class` - Storage class used to store the object.

### `access_control_translation` Block

* `owner` - Replica ownership.

### `encryption_configuration` Block

* `replica_kms_key_id` - ID (Key ARN or Alias ARN) of the customer managed AWS KMS key stored in KMS for the destination bucket.

### `metrics` Block

* `event_threshold` - Configuration block that specifies the time threshold for emitting the `s3:Replication:OperationMissedThreshold` event. See [`event_threshold` Block](#event_threshold-block) below.
* `status` - Status of the Destination Metrics.

### `event_threshold` Block

* `minutes` - Time in minutes.

### `replication_time` Block

* `status` - Status of the Replication Time Control.
* `time` - Configuration block that specifies the time by which replication should be complete for all objects and operations on objects. See [`time` Block](#time-block) below.

### `time` Block

* `minutes` - Time in minutes.

### `existing_object_replication` Block

* `status` - Whether existing objects are replicated.

### `filter` Block

* `and` - Configuration block for specifying rule filters. See [`and` Block](#and-block) below.
* `prefix` - Object key name prefix that identifies the subset of objects to which the rule applies.
* `tag` - Configuration block for specifying a tag key and value. See [`tag` Block](#tag-block) below.

### `and` Block

* `prefix` - Object key name prefix that identifies the subset of objects to which the rule applies.
* `tag` - List of tags that identify a subset of objects to which the rule applies. See [`tag` Block](#tag-block) below.

### `tag` Block

* `key` - Name of the object key.
* `value` - Value of the tag.

### `source_selection_criteria` Block

* `replica_modifications` - Configuration block for selections for modifications on replicas. See [`replica_modifications` Block](#replica_modifications-block) below.
* `sse_kms_encrypted_objects` - Configuration block for filter information for the selection of Amazon S3 objects encrypted with AWS KMS. See [`sse_kms_encrypted_objects` Block](#sse_kms_encrypted_objects-block) below.

### `replica_modifications` Block

* `status` - Whether Amazon S3 replicates modifications on replicas.

### `sse_kms_encrypted_objects` Block

* `status` - Whether Amazon S3 replicates objects created with server-side encryption using an AWS KMS key stored in KMS.
