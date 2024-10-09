---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image"
description: |-
    Manages an Image Builder Image
---

# Resource: aws_imagebuilder_image

Manages an Image Builder Image.

## Example Usage

```terraform
resource "aws_imagebuilder_image" "example" {
  distribution_configuration_arn   = aws_imagebuilder_distribution_configuration.example.arn
  image_recipe_arn                 = aws_imagebuilder_image_recipe.example.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.example.arn
}
```

## Argument Reference

The following arguments are required:

* `infrastructure_configuration_arn` - (Required) Amazon Resource Name (ARN) of the Image Builder Infrastructure Configuration.

The following arguments are optional:

* `container_recipe_arn` - (Optional) - Amazon Resource Name (ARN) of the container recipe.
* `distribution_configuration_arn` - (Optional) Amazon Resource Name (ARN) of the Image Builder Distribution Configuration.
* `enhanced_image_metadata_enabled` - (Optional) Whether additional information about the image being created is collected. Defaults to `true`.
* `execution_role` - (Optional) Amazon Resource Name (ARN) of the service-linked role to be used by Image Builder to [execute workflows](https://docs.aws.amazon.com/imagebuilder/latest/userguide/manage-image-workflows.html).
* `image_recipe_arn` - (Optional) Amazon Resource Name (ARN) of the image recipe.
* `image_tests_configuration` - (Optional) Configuration block with image tests configuration. Detailed below.
* `image_scanning_configuration` - (Optional) Configuration block with image scanning configuration. Detailed below.
* `workflow` - (Optional) Configuration block with the workflow configuration. Detailed below.
* `tags` - (Optional) Key-value map of resource tags for the Image Builder Image. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### image_tests_configuration

The following arguments are optional:

* `image_tests_enabled` - (Optional) Whether image tests are enabled. Defaults to `true`.
* `timeout_minutes` - (Optional) Number of minutes before image tests time out. Valid values are between `60` and `1440`. Defaults to `720`.

### image_scanning_configuration

The following arguments are optional:

* `image_scanning_enabled` - (Optional) Indicates whether Image Builder keeps a snapshot of the vulnerability scans that Amazon Inspector runs against the build instance when you create a new image. Defaults to `false`.
* `ecr_configuration` - (Optional) Configuration block with ECR configuration. Detailed below.

### ecr_configuration

The following arguments are optional:

* `repository_name` - (Optional) The name of the container repository that Amazon Inspector scans to identify findings for your container images.
* `container_tags` - (Optional) Set of tags for Image Builder to apply to the output container image that that Amazon Inspector scans.

### workflow

The following arguments are required:

* `workflow_arn` - (Required) Amazon Resource Name (ARN) of the Image Builder Workflow.

The following arguments are optional:

* `on_failure` - (Optional) The action to take if the workflow fails. Must be one of `CONTINUE` or `ABORT`.
* `parallel_group` - (Optional) The parallel group in which to run a test Workflow.
* `parameter` - (Optional) Configuration block for the workflow parameters. Detailed below.

### parameter

The following arguments are required:

* `name` - (Required) The name of the Workflow parameter.
* `value` - (Required) The value of the Workflow parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the image.
* `date_created` - Date the image was created.
* `platform` - Platform of the image.
* `os_version` - Operating System version of the image.
* `output_resources` - List of objects with resources created by the image.
    * `amis` - Set of objects with each Amazon Machine Image (AMI) created.
        * `account_id` - Account identifier of the AMI.
        * `description` - Description of the AMI.
        * `image` - Identifier of the AMI.
        * `name` - Name of the AMI.
        * `region` - Region of the AMI.
    * `containers` - Set of objects with each container image created and stored in the output repository.
        * `image_uris` - Set of URIs for created containers.
        * `region` - Region of the container image.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version` - Version of the image.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_imagebuilder_image` resources using the Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_imagebuilder_image.example
  id = "arn:aws:imagebuilder:us-east-1:123456789012:image/example/1.0.0/1"
}
```

Using `terraform import`, import `aws_imagebuilder_image` resources using the Amazon Resource Name (ARN). For example:

```console
% terraform import aws_imagebuilder_image.example arn:aws:imagebuilder:us-east-1:123456789012:image/example/1.0.0/1
```
