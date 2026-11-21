---
subcategory: "Kinesis Analytics V2"
layout: "aws"
page_title: "AWS: aws_kinesisanalyticsv2_snapshot"
description: |-
  Manages a Managed Service for Apache Flink enhanced snapshot.
---

# Resource: aws_kinesisanalyticsv2_snapshot

Manages a Managed Service for Apache Flink enhanced snapshot of a running application.

An enhanced snapshot is a standalone resource with its own ARN. Any application in enhanced snapshot mode in the same account and Region can restore it. Unlike an [`aws_kinesisanalyticsv2_application_snapshot`](/docs/providers/aws/r/kinesisanalyticsv2_application_snapshot.html), it is not deleted with its application. For more information, see [Snapshots](https://docs.aws.amazon.com/managed-flink/latest/java/how-snapshots.html) in the *Managed Service for Apache Flink Developer Guide*.

~> **NOTE:** The application must use `snapshot_mode = "ENHANCED_SNAPSHOTS"` in its `application_snapshot_configuration` and must be running when the snapshot is created.

~> **NOTE:** Enhanced snapshots, including the system snapshots that Managed Service for Apache Flink creates on stop and update, remain after the application is deleted and keep accruing storage charges until deleted. Terraform deletes only the snapshots it manages. A running application does not prevent deletion of a snapshot it would restore from.

## Example Usage

### Basic Usage

```terraform
resource "aws_kinesisanalyticsv2_snapshot" "example" {
  application_name = aws_kinesisanalyticsv2_application.example.name
  name             = "example-snapshot"
}
```

### Restore Into a Different Application

```terraform
resource "aws_kinesisanalyticsv2_snapshot" "blue" {
  application_name = aws_kinesisanalyticsv2_application.blue.name
  name             = "blue-for-green"
}

resource "aws_kinesisanalyticsv2_application" "green" {
  name                   = "green"
  runtime_environment    = "FLINK-1_20"
  service_execution_role = aws_iam_role.example.arn

  application_configuration {
    # ... other configuration ...

    application_snapshot_configuration {
      snapshots_enabled = true
      snapshot_mode     = "ENHANCED_SNAPSHOTS"
    }

    run_configuration {
      application_restore_configuration {
        application_restore_type = "RESTORE_FROM_CUSTOM_SNAPSHOT"
        snapshot_arn             = aws_kinesisanalyticsv2_snapshot.blue.arn
      }
    }
  }

  start_application = true
}
```

The caller that starts `green` needs `kinesisanalytics:RestoreFromSnapshot` on the snapshot ARN.

## Argument Reference

The following arguments are required:

* `application_name` - (Required) Name of the [application](/docs/providers/aws/r/kinesisanalyticsv2_application.html) to snapshot.
* `name` - (Required) Name of the snapshot. Must be unique among the snapshots in the account and Region. Names beginning with `SYSTEMSNAPSHOT` are reserved.

The following arguments are optional:

* `description` - (Optional) Summary description of the snapshot.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the snapshot. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `application_version_id` - Version ID of the application when the snapshot was created.
* `arn` - ARN of the snapshot.
* `creation_time` - Time the snapshot was created, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.8).
* `runtime_environment` - Runtime environment of the application when the snapshot was created.
* `size_in_bytes` - Size of the snapshot, in bytes.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_kinesisanalyticsv2_snapshot.example
  identity = {
    name = "example-snapshot"
  }
}

resource "aws_kinesisanalyticsv2_snapshot" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` - Name of the snapshot.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import enhanced snapshots using the `name`. For example:

```terraform
import {
  to = aws_kinesisanalyticsv2_snapshot.example
  id = "example-snapshot"
}
```

Using `terraform import`, import enhanced snapshots using the `name`. For example:

```console
% terraform import aws_kinesisanalyticsv2_snapshot.example example-snapshot
```
