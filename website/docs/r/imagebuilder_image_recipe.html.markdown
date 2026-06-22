---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image_recipe"
description: |-
    Manage an Image Builder Image Recipe
---

# Resource: aws_imagebuilder_image_recipe

Manages an Image Builder Image Recipe.

## Example Usage

```terraform
resource "aws_imagebuilder_image_recipe" "example" {
  block_device_mapping {
    device_name = "/dev/xvdb"

    ebs {
      delete_on_termination = true
      volume_size           = 100
      volume_type           = "gp2"
    }
  }

  component {
    component_arn = aws_imagebuilder_component.example.arn

    parameter {
      name  = "Parameter1"
      value = "Value1"
    }

    parameter {
      name  = "Parameter2"
      value = "Value2"
    }
  }

  name         = "example"
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.region}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
```

## Argument Reference

The following arguments are required:

* `component` - (Required) Ordered configuration block(s) with components for the image recipe. Detailed below.
* `name` - (Required) Name of the image recipe.
* `parent_image` - (Required) The image recipe uses this image as a base from which to build your customized image. The value can be the base image ARN, an AMI ID, or an SSM Parameter referencing the AMI. For an SSM Parameter, enter the prefix `ssm:`, followed by the parameter name or ARN.
* `version` - (Required) The semantic version of the image recipe, which specifies the version in the following format, with numeric values in each position to indicate a specific version: major.minor.patch. For example: 1.0.0.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `ami_tags` - (Optional)  Tags that are applied to the AMI that Image Builder creates during the Build phase prior to image distribution. Maximum of 50 tags.
* `block_device_mapping` - (Optional) Configuration block(s) with block device mappings for the image recipe. Detailed below.
* `description` - (Optional) Description of the image recipe.
* `systems_manager_agent` - (Optional) Configuration block for the Systems Manager Agent installed by default by Image Builder. Detailed below.
* `tags` - (Optional) Key-value map of resource tags for the image recipe. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_data_base64` - (Optional) Base64 encoded user data. Use this to provide commands or a command script to run when you launch your build instance.
* `working_directory` - (Optional) The working directory to be used during build and test workflows.

### `block_device_mapping`

* `device_name` - (Optional) Name of the device. For example, `/dev/sda` or `/dev/xvdb`.
* `ebs` - (Optional) Configuration block with Elastic Block Storage (EBS) block device mapping settings. Detailed below.
* `no_device` - (Optional) Set to `true` to remove a mapping from the parent image.
* `virtual_name` - (Optional) Virtual device name. For example, `ephemeral0`. Instance store volumes are numbered starting from 0.

#### `ebs`

* `delete_on_termination` - (Optional) Whether to delete the volume on termination. Defaults to unset, which is the value inherited from the parent image.
* `encrypted` - (Optional) Whether to encrypt the volume. Defaults to unset, which is the value inherited from the parent image.
* `iops` - (Optional) Number of Input/Output (I/O) operations per second to provision for an `io1` or `io2` volume.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of the Key Management Service (KMS) Key for encryption.
* `snapshot_id` - (Optional) Identifier of the EC2 Volume Snapshot.
* `throughput` - (Optional) For GP3 volumes only. The throughput in MiB/s that the volume supports.
* `volume_size` - (Optional) Size of the volume, in GiB.
* `volume_type` - (Optional) Type of the volume. For example, `gp2` or `io2`.

### `component`

* `component_arn` - (Required) Amazon Resource Name (ARN) of the Image Builder Component to associate.
* `parameter` - (Optional) Configuration block(s) for parameters to configure the component. Detailed below.

### `parameter`

* `name` - (Required) The name of the component parameter.
* `value` - (Required) The value for the named component parameter.

### `systems_manager_agent`

* `uninstall_after_build` - (Required) Whether to remove the Systems Manager Agent after the image has been built.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the image recipe.
* `arn` - Amazon Resource Name (ARN) of the image recipe.
* `date_created` - Date the image recipe was created.
* `owner` - Owner of the image recipe.
* `platform` - Platform of the image recipe.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_imagebuilder_image_recipe.example
  identity = {
    "arn" = "arn:aws:imagebuilder:us-east-1:123456789012:image-recipe/example/1.0.0"
  }
}

resource "aws_imagebuilder_image_recipe" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the Image Builder image recipe.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_imagebuilder_image_recipe` resources using the Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_imagebuilder_image_recipe.example
  id = "arn:aws:imagebuilder:us-east-1:123456789012:image-recipe/example/1.0.0"
}
```

Using `terraform import`, import `aws_imagebuilder_image_recipe` resources using the Amazon Resource Name (ARN). For example:

```console
% terraform import aws_imagebuilder_image_recipe.example arn:aws:imagebuilder:us-east-1:123456789012:image-recipe/example/1.0.0
```
