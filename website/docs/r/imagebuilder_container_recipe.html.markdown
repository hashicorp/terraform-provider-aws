---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_container_recipe"
description: |-
    Manage an Image Builder Container Recipe
---

# Resource: aws_imagebuilder_container_recipe

Manages an Image Builder Container Recipe.

## Example Usage

```terraform
resource "aws_imagebuilder_container_recipe" "example" {
  name    = "example"
  version = "1.0.0"

  container_type = "DOCKER"
  parent_image   = "arn:aws:imagebuilder:eu-central-1:aws:image/amazon-linux-x86-latest/x.x.x"

  target_repository {
    repository_name = aws_ecr_repository.example.name
    service         = "ECR"
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

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF
}
```

## Argument Reference

The following arguments are required:

* `component` - (Required) Ordered configuration block(s) with components for the container recipe. Detailed below.
* `container_type` - (Required) The type of the container to create. Valid values: `DOCKER`.
* `name` - (Required) The name of the container recipe.
* `parent_image` (Required) The base image for the container recipe.
* `target_repository` (Required) The destination repository for the container image. Detailed below.
* `version` (Required) Version of the container recipe.

The following attributes are optional:

* `description` - (Optional) The description of the container recipe.
* `dockerfile_template_data` - (Optional) The Dockerfile template used to build the image as an inline data blob.
* `dockerfile_template_uri` - (Optional) The Amazon S3 URI for the Dockerfile that will be used to build the container image.
* `instance_configuration` - (Optional) Configuration block used to configure an instance for building and testing container images. Detailed below.
* `kms_key_id` - (Optional) The KMS key used to encrypt the container image.
* `platform_override` - (Optional) Specifies the operating system platform when you use a custom base image.
* `skip_destroy` - (Optional) Whether to retain the old version when the resource is destroyed or replacement is necessary. Defaults to `false`.
* `tags` - (Optional) Key-value map of resource tags for the container recipe. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `working_directory` - (Optional) The working directory to be used during build and test workflows.

### component

The `component` block supports the following arguments:

* `component_arn` - (Required) Amazon Resource Name (ARN) of the Image Builder Component to associate.
* `parameter` - (Optional) Configuration block(s) for parameters to configure the component. Detailed below.

### parameter

The following arguments are required:

* `name` - (Required) The name of the component parameter.
* `value` - (Required) The value for the named component parameter.

### target_repository

The following arguments are required:

* `repository_name` - (Required) The name of the container repository where the output container image is stored. This name is prefixed by the repository location.
* `service` - (Required) The service in which this image is registered. Valid values: `ECR`.

### instance_configuration

The following arguments are optional:

* `block_device_mapping` - (Optional) Configuration block(s) with block device mappings for the container recipe. Detailed below.
* `image` - (Optional) The AMI ID to use as the base image for a container build and test instance. If not specified, Image Builder will use the appropriate ECS-optimized AMI as a base image.

### block_device_mapping

The following arguments are optional:

* `device_name` - (Optional) Name of the device. For example, `/dev/sda` or `/dev/xvdb`.
* `ebs` - (Optional) Configuration block with Elastic Block Storage (EBS) block device mapping settings. Detailed below.
* `no_device` - (Optional) Set to `true` to remove a mapping from the parent image.
* `virtual_name` - (Optional) Virtual device name. For example, `ephemeral0`. Instance store volumes are numbered starting from 0.

#### ebs

The following arguments are optional:

* `delete_on_termination` - (Optional) Whether to delete the volume on termination. Defaults to unset, which is the value inherited from the parent image.
* `encrypted` - (Optional) Whether to encrypt the volume. Defaults to unset, which is the value inherited from the parent image.
* `iops` - (Optional) Number of Input/Output (I/O) operations per second to provision for an `io1` or `io2` volume.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of the Key Management Service (KMS) Key for encryption.
* `snapshot_id` - (Optional) Identifier of the EC2 Volume Snapshot.
* `throughput` - (Optional) For GP3 volumes only. The throughput in MiB/s that the volume supports.
* `volume_size` - (Optional) Size of the volume, in GiB.
* `volume_type` - (Optional) Type of the volume. For example, `gp2` or `io2`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - (Required) Amazon Resource Name (ARN) of the container recipe.
* `date_created` - Date the container recipe was created.
* `encrypted` - A flag that indicates if the target container is encrypted.
* `owner` - Owner of the container recipe.
* `platform` - Platform of the container recipe.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_imagebuilder_container_recipe` resources using the Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_imagebuilder_container_recipe.example
  id = "arn:aws:imagebuilder:us-east-1:123456789012:container-recipe/example/1.0.0"
}
```

Using `terraform import`, import `aws_imagebuilder_container_recipe` resources using the Amazon Resource Name (ARN). For example:

```console
% terraform import aws_imagebuilder_container_recipe.example arn:aws:imagebuilder:us-east-1:123456789012:container-recipe/example/1.0.0
```
