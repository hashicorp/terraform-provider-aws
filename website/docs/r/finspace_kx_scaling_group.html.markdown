---
subcategory: "FinSpace"
layout: "aws"
page_title: "AWS: aws_finspace_kx_scaling_group"
description: |-
  Terraform resource for managing an AWS FinSpace Kx Scaling Group.
---

# Resource: aws_finspace_kx_scaling_group

Terraform resource for managing an AWS FinSpace Kx Scaling Group.

## Example Usage

### Basic Usage

```terraform
resource "aws_finspace_kx_scaling_group" "example" {
  name                 = "my-tf-kx-scalinggroup"
  environment_id       = aws_finspace_kx_environment.example.id
  availability_zone_id = "use1-az2"
  host_type            = "kx.sg.4xlarge"
}
```

## Argument Reference

The following arguments are required:

* `availability_zone_id` - (Required) The availability zone identifiers for the requested regions.
* `environment_id` - (Required) A unique identifier for the kdb environment, where you want to create the scaling group.
* `name` - (Required) Unique name for the scaling group that you want to create.
* `host_type` - (Required) The memory and CPU capabilities of the scaling group host on which FinSpace Managed kdb clusters will be placed.

The following arguments are optional:

* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level. You can add up to 50 tags to a scaling group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifier of the KX Scaling Group.
* `clusters` - The list of Managed kdb clusters that are currently active in the given scaling group.
* `created_timestamp` - The timestamp at which the scaling group was created in FinSpace. The value is determined as epoch time in milliseconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000000.
* `last_modified_timestamp` - Last timestamp at which the scaling group was updated in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `status` - The status of scaling group.
    * `CREATING` – The scaling group creation is in progress.
    * `CREATE_FAILED` – The scaling group creation has failed.
    * `ACTIVE` – The scaling group is active.
    * `UPDATING` – The scaling group is in the process of being updated.
    * `UPDATE_FAILED` – The update action failed.
    * `DELETING` – The scaling group is in the process of being deleted.
    * `DELETE_FAILED` – The system failed to delete the scaling group.
    * `DELETED` – The scaling group is successfully deleted.
* `status_reason` - The error message when a failed state occurs.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `4h`)
* `update` - (Default `30m`)
* `delete` - (Default `4h`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS FinSpace Kx scaling group using the `id` (environment ID and scaling group name, comma-delimited). For example:

```terraform
import {
  to = aws_finspace_kx_scaling_group.example
  id = "n3ceo7wqxoxcti5tujqwzs,my-tf-kx-scalinggroup"
}
```

Using `terraform import`, import an AWS FinSpace Kx Scaling Group using the `id` (environment ID and scaling group name, comma-delimited). For example:

```console
% terraform import aws_finspace_kx_scaling_group.example n3ceo7wqxoxcti5tujqwzs,my-tf-kx-scalinggroup
```
