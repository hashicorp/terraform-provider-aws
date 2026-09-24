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

* `application_name` - (Required) Name of an existing [Kinesis Analytics v2 Application](/docs/providers/aws/r/kinesisanalyticsv2_application.html). Note that the application must be running for a snapshot to be created.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `snapshot_name` - (Required) Name of the application snapshot.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `application_version_id` - Current application version ID when the snapshot was created.
* `id` - Application snapshot identifier.
* `snapshot_creation_timestamp` - Timestamp of the application snapshot.

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
