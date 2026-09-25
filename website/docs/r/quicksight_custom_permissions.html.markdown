---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_custom_permissions"
description: |-
  Manages a QuickSight custom permissions profile.
---

# Resource: aws_quicksight_custom_permissions

Manages a QuickSight custom permissions profile.

## Example Usage

```terraform
resource "aws_quicksight_custom_permissions" "example" {
  custom_permissions_name = "example-permissions"

  capabilities {
    print_reports    = "DENY"
    share_dashboards = "DENY"
  }
}
```

## Argument Reference

The following arguments are required:

* `capabilities` - (Required) Actions to include in the custom permissions profile. See [`capabilities` Block](#capabilities-block).
* `custom_permissions_name` - (Required, Forces new resource) Custom permissions profile name.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `capabilities` Block

* `add_or_run_anomaly_detection_for_analyses` - (Optional) Ability to add or run anomaly detection. Valid values: `DENY`.
* `create_and_update_dashboard_email_reports` - (Optional) Ability to create and update email reports. Valid values: `DENY`.
* `create_and_update_data_sources` - (Optional) Ability to create and update data sources. Valid values: `DENY`.
* `create_and_update_datasets` - (Optional) Ability to create and update datasets. Valid values: `DENY`.
* `create_and_update_themes` - (Optional) Ability to create and update themes. Valid values: `DENY`.
* `create_and_update_threshold_alerts` - (Optional) Ability to create and update threshold alerts. Valid values: `DENY`.
* `create_shared_folders` - (Optional) Ability to create shared folders. Valid values: `DENY`.
* `create_spice_dataset` - (Optional) Ability to create a SPICE dataset. Valid values: `DENY`.
* `export_to_csv` - (Optional) Ability to export to CSV files from the UI. Valid values: `DENY`.
* `export_to_csv_in_scheduled_reports` - (Optional) Ability to export to CSV files in scheduled email reports. Valid values: `DENY`.
* `export_to_excel` - (Optional) Ability to export to Excel files from the UI. Valid values: `DENY`.
* `export_to_excel_in_scheduled_reports` - (Optional) Ability to export to Excel files in scheduled email reports. Valid values: `DENY`.
* `export_to_pdf` - (Optional) Ability to export to PDF files from the UI. Valid values: `DENY`.
* `export_to_pdf_in_scheduled_reports` - (Optional) Ability to export to PDF files in scheduled email reports. Valid values: `DENY`.
* `include_content_in_scheduled_reports_email` - (Optional) Ability to include content in scheduled email reports. Valid values: `DENY`.
* `print_reports` - (Optional) Ability to print reports. Valid values: `DENY`.
* `rename_shared_folders` - (Optional) Ability to rename shared folders. Valid values: `DENY`.
* `share_analyses` - (Optional) Ability to share analyses. Valid values: `DENY`.
* `share_dashboards` - (Optional) Ability to share dashboards. Valid values: `DENY`.
* `share_data_sources` - (Optional) Ability to share data sources. Valid values: `DENY`.
* `share_datasets` - (Optional) Ability to share datasets. Valid values: `DENY`.
* `subscribe_dashboard_email_reports` - (Optional) Ability to subscribe to email reports. Valid values: `DENY`.
* `view_account_spice_capacity` - (Optional) Ability to view account SPICE capacity. Valid values: `DENY`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the custom permissions profile.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a QuickSight custom permissions profile using the AWS account ID and custom permissions profile name separated by a comma (`,`). For example:

```terraform
import {
  to = aws_quicksight_custom_permissions.example
  id = "123456789012,example-permissions"
}
```

Using `terraform import`, import a QuickSight custom permissions profile using the AWS account ID and custom permissions profile name separated by a comma (`,`). For example:

```console
% terraform import aws_quicksight_custom_permissions.example 123456789012,example-permissions
```
