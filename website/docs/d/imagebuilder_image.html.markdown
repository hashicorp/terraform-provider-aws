---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image"
description: |-
    Provides details about an Image Builder Image
---

# Data Source: aws_imagebuilder_image

Provides details about an Image Builder Image.

## Example Usage

### Latest

```terraform
data "aws_imagebuilder_image" "example" {
  arn = "arn:aws:imagebuilder:us-west-2:aws:image/amazon-linux-2-x86/x.x.x"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `arn` - (Required) ARN of the image. The suffix can either be specified with wildcards (`x.x.x`) to fetch the latest build version or a full build version (e.g., `2020.11.26/1`) to fetch an exact version.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `build_version_arn` - Build version ARN of the image. This will always have the `#.#.#/#` suffix.
* `container_recipe_arn` - ARN of the container recipe.
* `date_created` - Date the image was created.
* `distribution_configuration_arn` - ARN of the Image Builder Distribution Configuration.
* `enhanced_image_metadata_enabled` - Whether additional information about the image being created is collected.
* `image_recipe_arn` - ARN of the image recipe.
* `image_scanning_configuration` - List of an object with image scanning configuration fields.
    * `image_scanning_enabled` - Indicates whether Image Builder keeps a snapshot of the vulnerability scans that Amazon Inspector runs against the build instance when you create a new image.
    * `ecr_configuration` - Configuration block with ECR configuration.
        * `repository_name` - The name of the container repository that Amazon Inspector scans to identify findings for your container images.
        * `container_tags` - Set of tags for Image Builder to apply to the output container image that that Amazon Inspector scans.
* `image_tests_configuration` - List of an object with image tests configuration.
    * `image_tests_enabled` - Whether image tests are enabled.
    * `timeout_minutes` - Number of minutes before image tests time out.
* `infrastructure_configuration_arn` - ARN of the Image Builder Infrastructure Configuration.
* `name` - Name of the image.
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
* `tags` - Key-value map of resource tags for the image.
* `version` - Version of the image.
