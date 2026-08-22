---
subcategory: "Kinesis Analytics V2"
layout: "aws"
page_title: "AWS: aws_kinesisanalyticsv2_application_maintenance_configuration"
description: |-
  Manages an AWS Kinesis Analytics V2 Application Maintenance Configuration.
---

# Resource: aws_kinesisanalyticsv2_application_maintenance_configuration

Manages an AWS Kinesis Analytics V2 Application Maintenance Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_kinesisanalyticsv2_application_maintenance_configuration" "example" {
  application_name = aws_kinesisanalyticsv2_application.example.name
  application_maintenance_window_start_time = "02:00"
}
```

## Argument Reference

The following arguments are required:

* `application_name` - (Required) The name of an existing  [Kinesis Analytics v2 Application](/docs/providers/aws/r/kinesisanalyticsv2_application.html). Note that the application must be running for a snapshot to be created.
* `application_maintenance_window_start_time` - (Required) The starting time (in UTC) of the custom maintenance window. Note that the end time will be automatically set as application_maintenance_window_start_time + 8 hours. Please refer to [AWS Documentation](/https://docs.aws.amazon.com/managed-flink/latest/java/maintenance.html?icmpid=docs_console_unmapped) for the full documentation.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The application snapshot identifier.
* `application_maintenance_window_start_time` - The starting time (in UTC) of the custom maintenance window.
* `snapshot_creation_timestamp` - The original starting time (in UTC) of the managed Flink application

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Kinesis Analytics V2 Application Maintenance Configuration using the `example_id_arg`. For example:

```terraform
import {
  to = aws_kinesisanalyticsv2_application_maintenance_configuration.example
  id = "application_maintenance_configuration-id-12345678"
}
```

Using `terraform import`, import Kinesis Analytics V2 Application Maintenance Configuration using the `example_id_arg`. For example:

```console
% terraform import aws_kinesisanalyticsv2_application_maintenance_configuration.example application_maintenance_configuration-id-12345678
```
