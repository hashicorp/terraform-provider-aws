---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image_recipe"
description: |-
    Provides details about an Image Builder Image Recipe
---

# Data Source: aws_imagebuilder_image_recipe

Provides details about an Image Builder Image Recipe.

## Example Usage

```terraform
data "aws_imagebuilder_image_recipe" "example" {
  arn = "arn:aws:imagebuilder:us-east-1:aws:image-recipe/example/1.0.0"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `arn` - (Required) ARN of the image recipe.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `block_device_mapping` - Set of objects with block device mappings for the image recipe.
    * `device_name` - Name of the device. For example, `/dev/sda` or `/dev/xvdb`.
    * `ebs` - Single list of object with Elastic Block Storage (EBS) block device mapping settings.
        * `delete_on_termination` - Whether to delete the volume on termination. Defaults to unset, which is the value inherited from the parent image.
        * `encrypted` - Whether to encrypt the volume. Defaults to unset, which is the value inherited from the parent image.
        * `iops` - Number of Input/Output (I/O) operations per second to provision for an `io1` or `io2` volume.
        * `kms_key_id` - ARN of the Key Management Service (KMS) Key for encryption.
        * `snapshot_id` - Identifier of the EC2 Volume Snapshot.
        * `throughput` - For GP3 volumes only. The throughput in MiB/s that the volume supports.
        * `volume_size` - Size of the volume, in GiB.
        * `volume_type` - Type of the volume. For example, `gp2` or `io2`.
    * `no_device` - Whether to remove a mapping from the parent image.
    * `virtual_name` - Virtual device name. For example, `ephemeral0`. Instance store volumes are numbered starting from 0.
* `component` - List of objects with components for the image recipe.
    * `component_arn` - ARN of the Image Builder Component.
    * `parameter` - Set of parameters that are used to configure the component.
        * `name` - Name of the component parameter.
        * `value` - Value of the component parameter.
* `date_created` - Date the image recipe was created.
* `description` - Description of the image recipe.
* `name` - Name of the image recipe.
* `owner` - Owner of the image recipe.
* `parent_image` - Base image of the image recipe.
* `platform` - Platform of the image recipe.
* `tags` - Key-value map of resource tags for the image recipe.
* `user_data_base64` - Base64 encoded contents of user data. Commands or a command script to run when build instance is launched.
* `version` - Version of the image recipe.
* `working_directory` - Working directory used during build and test workflows.
