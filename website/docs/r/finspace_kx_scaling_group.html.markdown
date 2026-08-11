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

* `availability_zone_id` - (Required) Availability zone identifiers for the requested regions.
* `environment_id` - (Required) Unique identifier for the kdb environment, where you want to create the scaling group.
* `host_type` - (Required) Memory and CPU capabilities of the scaling group host on which FinSpace Managed kdb clusters will be placed.
* `name` - (Required) Unique name for the scaling group that you want to create.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level. You can add up to 50 tags to a scaling group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) identifier of the KX Scaling Group.
* `clusters` - List of Managed kdb clusters that are currently active in the given scaling group.
* `created_timestamp` - Timestamp at which the scaling group was created in FinSpace. The value is determined as epoch time in milliseconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000000.
* `last_modified_timestamp` - Last timestamp at which the scaling group was updated in FinSpace. Value determined as epoch time in seconds. For example, the value for Monday, November 1, 2021 12:00:00 PM UTC is specified as 1635768000.
* `status` - Status of scaling group (`CREATING`, `CREATE_FAILED`, `ACTIVE`, `UPDATING`, `UPDATE_FAILED`, `DELETING`, `DELETE_FAILED`, `DELETED`).
* `status_reason` - Error message when a failed state occurs.
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
