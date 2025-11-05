---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_cloud_exadata_infrastructure"
description: |-
  Terraform resource for managing exadata infrastructure resource for Oracle Database@AWS.
---

# Resource: aws_odb_cloud_exadata_infrastructure

Terraform resource for managing exadata infrastructure resource in AWS for Oracle Database@AWS.

## Example Usage

### Basic Usage

```terraform

resource "aws_odb_cloud_exadata_infrastructure" "example" {
  display_name                     = "my-exa-infra"
  shape                            = "Exadata.X11M"
  storage_count                    = 3
  compute_count                    = 2
  availability_zone_id             = "use1-az6"
  customer_contacts_to_send_to_oci = [{ email = "abc@example.com" }, { email = "def@example.com" }]
  database_server_type             = "X11M"
  storage_server_type              = "X11M-HC"
  maintenance_window {
    custom_action_timeout_in_mins    = 16
    days_of_week                     = [{ name = "MONDAY" }, { name = "TUESDAY" }]
    hours_of_day                     = [11, 16]
    is_custom_action_timeout_enabled = true
    lead_time_in_weeks               = 3
    months                           = [{ name = "FEBRUARY" }, { name = "MAY" }, { name = "AUGUST" }, { name = "NOVEMBER" }]
    patching_mode                    = "ROLLING"
    preference                       = "CUSTOM_PREFERENCE"
    weeks_of_month                   = [2, 4]
  }
  tags = {
    "env" = "dev"
  }

}

resource "aws_odb_cloud_exadata_infrastructure" "example" {
  display_name         = "my_exa_X9M"
  shape                = "Exadata.X9M"
  storage_count        = 3
  compute_count        = 2
  availability_zone_id = "use1-az6"
  maintenance_window {
    custom_action_timeout_in_mins    = 16
    is_custom_action_timeout_enabled = true
    patching_mode                    = "ROLLING"
    preference                       = "NO_PREFERENCE"
  }
}
```

## Argument Reference

The following arguments are required:

* `display_name` - (Required) The user-friendly name for the Exadata infrastructure. Changing this will force terraform to create a new resource.
* `shape` - (Required) The model name of the Exadata infrastructure. Changing this will force terraform to create new resource.
* `storage_count` - (Required) The number of storage servers that are activated for the Exadata infrastructure. Changing this will force terraform to create new resource.
* `compute_count` - (Required) The number of compute instances that the Exadata infrastructure is located. Changing this will force terraform to create new resource.
* `availability_zone_id` - (Required) The AZ ID of the AZ where the Exadata infrastructure is located. Changing this will force terraform to create new resource.

The following arguments are optional:

* `customer_contacts_to_send_to_oci` - (Optional) The email addresses of contacts to receive notification from Oracle about maintenance updates for the Exadata infrastructure. Changing this will force terraform to create new resource.
* `availability_zone`: (Optional) The name of the Availability Zone (AZ) where the Exadata infrastructure is located. Changing this will force terraform to create new resource.
* `database_server_type` - (Optional) The database server model type of the Exadata infrastructure. For the list of valid model names, use the ListDbSystemShapes operation. This is a mandatory parameter for Exadata.X11M system shape. Changing this will force terraform to create new resource.
* `storage_server_type` - (Optional) The storage server model type of the Exadata infrastructure. For the list of valid model names, use the ListDbSystemShapes operation. This is a mandatory parameter for Exadata.X11M system shape. Changing this will force terraform to create new resource.
* `tags` - (Optional) A map of tags to assign to the exadata infrastructure. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### maintenance_window

* `custom_action_timeout_in_mins` - (Required) The custom action timeout in minutes for the maintenance window.
* `is_custom_action_timeout_enabled` - (Required) ndicates whether custom action timeout is enabled for the maintenance window.
* `patching_mode` - (Required) The patching mode for the maintenance window.
* `preference` - (Required) The preference for the maintenance window scheduling.
* `days_of_week` - (Optional) The days of the week when maintenance can be performed.
* `hours_of_day` - (Optional) The hours of the day when maintenance can be performed.
* `lead_time_in_weeks` - (Optional) The lead time in weeks before the maintenance window.
* `months` - (Optional) The months when maintenance can be performed.
* `weeks_of_month` - (Optional) The weeks of the month when maintenance can be performed.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the Exadata infrastructure.
* `arn` - Amazon Resource Name (ARN) of the Exadata infrastructure.
* `activated_storage_count` - The number of storage servers requested for the Exadata infrastructure.
* `additional_storage_count` - The number of storage servers requested for the Exadata infrastructure.
* `available_storage_size_in_gbs` - The amount of available storage, in gigabytes (GB), for the Exadata infrastructure.
* `cpu_count` - The total number of CPU cores that are allocated to the Exadata infrastructure.
* `data_storage_size_in_tbs` - The size of the Exadata infrastructure's data disk group, in terabytes (TB).
* `db_node_storage_size_in_gbs` - The size of the Exadata infrastructure's local node storage, in gigabytes (GB).
* `db_server_version` - The software version of the database servers (dom0) in the Exadata infrastructure.
* `last_maintenance_run_id` - The Oracle Cloud Identifier (OCID) of the last maintenance run for the Exadata infrastructure.
* `max_cpu_count` -  The total number of CPU cores available on the Exadata infrastructure.
* `max_data_storage_in_tbs` - The total amount of data disk group storage, in terabytes (TB), that's available on the Exadata infrastructure.
* `max_db_node_storage_size_in_gbs` - The total amount of local node storage, in gigabytes (GB), that's available on the Exadata infrastructure.
* `max_memory_in_gbs` - The total amount of memory in gigabytes (GB) available on the Exadata infrastructure.
* `monthly_db_server_version` - The monthly software version of the database servers in the Exadata infrastructure.
* `monthly_storage_server_version` - The monthly software version of the storage servers installed on the Exadata infrastructure.
* `next_maintenance_run_id` - The OCID of the next maintenance run for the Exadata infrastructure.
* `ocid` - The OCID of the Exadata infrastructure.
* `oci_resource_anchor_name` - The name of the OCI resource anchor for the Exadata infrastructure.
* `percent_progress` - The amount of progress made on the current operation on the Exadata infrastructure, expressed as a percentage.
* `status` - The current status of the Exadata infrastructure.
* `status_reason` - Additional information about the status of the Exadata infrastructure.
* `storage_server_version` - The software version of the storage servers on the Exadata infrastructure.
* `total_storage_size_in_gbs` - The total amount of storage, in gigabytes (GB), on the Exadata infrastructure.
* `created_at` - The time when the Exadata infrastructure was created.
* `compute_model` - The OCI model compute model used when you create or clone an instance: ECPU or OCPU.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `24h`)
* `update` - (Default `24h`)
* `delete` - (Default `24h`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch Ingestion Pipeline using the `id`. For example:

```terraform
import {
  to = aws_odb_cloud_exadata_infrastructure.example
  id = "example"
}
```

Using `terraform import`, import Exadata Infrastructure using the `id`. For example:

```console
% terraform import aws_odb_cloud_exadata_infrastructure.example example
```
