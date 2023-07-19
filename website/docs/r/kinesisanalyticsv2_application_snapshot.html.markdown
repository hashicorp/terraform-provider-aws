---
subcategory: "Kinesis Analytics V2"
layout: "aws"
page_title: "AWS: aws_kinesisanalyticsv2_application_snapshot"
description: |-
  Manages a Kinesis Analytics v2 Application Snapshot.
---

# Resource: aws_kinesisanalyticsv2_application_snapshot

Manages a Kinesis Analytics v2 Application Snapshot.
Snapshots are the AWS implementation of [Flink Savepoints](https://ci.apache.org/projects/flink/flink-docs-release-1.11/ops/state/savepoints.html).

## Example Usage

```terraform
resource "aws_kinesisanalyticsv2_application_snapshot" "example" {
  application_name = aws_kinesisanalyticsv2_application.example.name
  snapshot_name    = "example-snapshot"
}
```

## Argument Reference

This resource supports the following arguments:

* `application_name` - (Required) The name of an existing  [Kinesis Analytics v2 Application](/docs/providers/aws/r/kinesisanalyticsv2_application.html). Note that the application must be running for a snapshot to be created.
* `snapshot_name` - (Required) The name of the application snapshot.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The application snapshot identifier.
* `application_version_id` - The current application version ID when the snapshot was created.
* `snapshot_creation_timestamp` - The timestamp of the application snapshot.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_kinesisanalyticsv2_application` using `application_name` together with `snapshot_name`. For example:

```terraform
import {
  to = aws_kinesisanalyticsv2_application_snapshot.example
  id = "example-application/example-snapshot"
}
```

Using `terraform import`, import `aws_kinesisanalyticsv2_application` using `application_name` together with `snapshot_name`. For example:

```console
% terraform import aws_kinesisanalyticsv2_application_snapshot.example example-application/example-snapshot
```
