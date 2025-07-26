---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image_pipeline"
description: |-
    Provides details about an Image Builder Image Pipeline
---

# Data Source: aws_imagebuilder_image_pipeline

Provides details about an Image Builder Image Pipeline.

## Example Usage

```terraform
data "aws_imagebuilder_image_pipeline" "example" {
  arn = "arn:aws:imagebuilder:us-west-2:aws:image-pipeline/example"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `arn` - (Required) ARN of the image pipeline.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `container_recipe_arn` - ARN of the container recipe.
* `date_created` - Date the image pipeline was created.
* `date_last_run` - Date the image pipeline was last run.
* `date_next_run` - Date the image pipeline will run next.
* `date_updated` - Date the image pipeline was updated.
* `description` - Description of the image pipeline.
* `distribution_configuration_arn` - ARN of the Image Builder Distribution Configuration.
* `enhanced_image_metadata_enabled` - Whether additional information about the image being created is collected.
* `image_recipe_arn` - ARN of the image recipe.
* `image_tests_configuration` - List of an object with image scanning configuration.
    * `image_scanning_enabled`  - Whether image scanning is enabled.
    * `ecr_configuration`  - List if an object with ecr configuration for image scanning
        * `container_tags`  - Tags that are added to the output containers that are scanned
        * `repository_name`  - The name of the container repository that Amazon Inspector scans
* `image_tests_configuration` - List of an object with image tests configuration.
    * `image_tests_enabled` - Whether image tests are enabled.
    * `timeout_minutes` - Number of minutes before image tests time out.
* `infrastructure_configuration_arn` - ARN of the Image Builder Infrastructure Configuration.
* `name` - Name of the image pipeline.
* `platform` - Platform of the image pipeline.
* `schedule` - List of an object with schedule settings.
    * `pipeline_execution_start_condition` - Condition when the pipeline should trigger a new image build.
    * `schedule_expression` - Cron expression of how often the pipeline start condition is evaluated.
* `status` - Status of the image pipeline.
* `tags` - Key-value map of resource tags for the image pipeline.
