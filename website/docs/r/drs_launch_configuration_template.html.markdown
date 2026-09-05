---
subcategory: "DRS (Elastic Disaster Recovery)"
layout: "aws"
page_title: "AWS: aws_drs_launch_configuration_template"
description: |-
  Manages an AWS DRS (Elastic Disaster Recovery) Launch Configuration Template.
---

# Resource: aws_drs_launch_configuration_template

Manages an AWS DRS (Elastic Disaster Recovery) Launch Configuration Template. Before using DRS, your account must be [initialized](https://docs.aws.amazon.com/drs/latest/userguide/getting-started-initializing.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_drs_launch_configuration_template" "example" {
  copy_private_ip             = true
  copy_tags                   = false
  launch_disposition          = "STARTED"
  launch_into_source_instance = false
  post_launch_enabled         = false
  recovery_mode               = "OPTIMAL"

  licensing {
    os_byol = true
  }

  target_instance_type_right_sizing_method = "IN_AWS"
}
```

## Argument Reference

The following arguments are optional:

* `copy_private_ip` - (Optional) Whether to copy the private IP of the Source Server to the Recovery Instance.
* `copy_tags` - (Optional) Whether to copy tags from the Source Server to EC2 resources created by the recovery job.
* `export_bucket_arn` - (Optional) ARN of the S3 bucket to export Source Network templates to.
* `launch_disposition` - (Optional) Launch disposition to use when launching a Recovery Instance. Valid values are `STOPPED` and `STARTED`.
* `launch_into_source_instance` - (Optional) Whether to launch a Recovery Instance using the instance ID of the source instance when performing a drill, recovery, or failback to the previous Region or Availability Zone.
* `licensing` - (Optional) Configuration block for licensing. [See below](#licensing).
* `post_launch_enabled` - (Optional) Whether to enable post-launch actions.
* `recovery_mode` - (Optional) Recovery mode to use when recovering the Source Server. Valid values are `FAST` and `OPTIMAL`.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `target_instance_type_right_sizing_method` - (Optional) Instance type right-sizing method to use when launching a Recovery Instance. Valid values are `NONE`, `BASIC`, and `IN_AWS`.

### `licensing`

* `os_byol` - (Optional) Whether to enable "Bring your own license" licensing for OS byol.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Launch Configuration Template.
* `id` - ID of the Launch Configuration Template.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DRS (Elastic Disaster Recovery) Launch Configuration Template using the `id`. For example:

```terraform
import {
  to = aws_drs_launch_configuration_template.example
  id = "lct-abcd1234efgh5678i"
}
```

Using `terraform import`, import DRS (Elastic Disaster Recovery) Launch Configuration Template using the `id`. For example:

```console
% terraform import aws_drs_launch_configuration_template.example lct-abcd1234efgh5678i
```
