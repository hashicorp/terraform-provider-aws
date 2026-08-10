---
subcategory: "FinSpace"
layout: "aws"
page_title: "AWS: aws_finspace_kx_volume"
description: |-
  Terraform resource for managing an AWS FinSpace Kx Volume.
---

# Resource: aws_finspace_kx_volume

Terraform resource for managing an AWS FinSpace Kx Volume.

## Example Usage

### Basic Usage

```terraform
resource "aws_finspace_kx_volume" "example" {
  name               = "my-tf-kx-volume"
  environment_id     = aws_finspace_kx_environment.example.id
  availability_zones = ["use1-az2"]
  az_mode            = "SINGLE"
  type               = "NAS_1"
  nas1_configuration {
    size = 1200
    type = "SSD_250"
  }
}
```

## Argument Reference

The following arguments are required:

* `availability_zones` - (Required) Identifier of the AWS Availability Zone IDs.
* `az_mode` - (Required) Number of availability zones you want to assign per volume. Currently, FinSpace only supports `SINGLE` for volumes, which assigns one availability zone per volume.
* `environment_id` - (Required) Unique identifier for the kdb environment, whose clusters can attach to the volume.
* `name` - (Required) Unique name for the volume that you want to create.
* `type` - (Required) Type of file system volume. Currently, FinSpace only supports the `NAS_1` volume type. When you select the `NAS_1` volume type, you must also provide `nas1_configuration`.

The following arguments are optional:

* `description` - (Optional) Description of the volume.
* `nas1_configuration` - (Optional) Configuration for the Network attached storage (`NAS_1`) file system volume. This parameter is required when `volume_type` is `NAS_1`. See [`nas1_configuration` Block](#nas1_configuration-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value pairs to label the volume. You can add up to 50 tags to a volume.

### `nas1_configuration` Block

The `nas1_configuration` block supports the following arguments:

* `size` - (Required) Size of the network attached storage.
* `type` - (Required) Type of the network attached storage.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifier of the KX volume.
* `attached_clusters` - Clusters attached to the volume. See [`attached_clusters` Block](#attached_clusters-block) below.
* `created_timestamp` - Timestamp at which the volume was created in FinSpace. The value is determined as epoch time in milliseconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000000.
* `last_modified_timestamp` - Last timestamp at which the volume was updated in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `status` - Status of volume creation. Values are `CREATING` (volume creation is in progress), `CREATE_FAILED` (volume creation has failed), `ACTIVE` (volume is active), `UPDATING` (volume is in the process of being updated), `UPDATE_FAILED` (update action failed), `UPDATED` (volume is successfully updated), `DELETING` (volume is in the process of being deleted), `DELETE_FAILED` (system failed to delete the volume), and `DELETED` (volume is successfully deleted).
* `status_reason` - Error message when a failed state occurs.

### `attached_clusters` Block

* `cluster_name` - Name of the KX cluster.
* `cluster_status` - Status of the KX cluster.
* `cluster_type` - Type of the KX cluster.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `45m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS FinSpace Kx Volume using the `id` (environment ID and volume name, comma-delimited). For example:

```terraform
import {
  to = aws_finspace_kx_volume.example
  id = "n3ceo7wqxoxcti5tujqwzs,my-tf-kx-volume"
}
```

Using `terraform import`, import an AWS FinSpace Kx Volume using the `id` (environment ID and volume name, comma-delimited). For example:

```console
% terraform import aws_finspace_kx_volume.example n3ceo7wqxoxcti5tujqwzs,my-tf-kx-volume
```
