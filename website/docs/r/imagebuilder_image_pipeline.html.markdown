---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image_pipeline"
description: |-
    Manages an Image Builder Image Pipeline
---

# Resource: aws_imagebuilder_image_pipeline

Manages an Image Builder Image Pipeline.

~> **NOTE:** Starting with version `5.74.0`, lifecycle meta-argument [`replace_triggered_by`](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#replace_triggered_by) must be used in order to prevent a dependency error on destroy.

## Example Usage

```terraform
resource "aws_imagebuilder_image_pipeline" "example" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.example.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.example.arn
  name                             = "example"

  schedule {
    schedule_expression = "cron(0 0 * * ? *)"
  }

  lifecycle {
    replace_triggered_by = [
      aws_imagebuilder_image_recipe.example
    ]
  }
}

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

* `infrastructure_configuration_arn` - (Required) Amazon Resource Name (ARN) of the Image Builder Infrastructure Configuration.
* `name` - (Required) Name of the image pipeline.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `container_recipe_arn` - (Optional) Amazon Resource Name (ARN) of the container recipe.
* `description` - (Optional) Description of the image pipeline.
* `distribution_configuration_arn` - (Optional) Amazon Resource Name (ARN) of the Image Builder Distribution Configuration.
* `enhanced_image_metadata_enabled` - (Optional) Whether additional information about the image being created is collected. Defaults to `true`.
* `execution_role` - (Optional) Amazon Resource Name (ARN) of the service-linked role to be used by Image Builder to [execute workflows](https://docs.aws.amazon.com/imagebuilder/latest/userguide/manage-image-workflows.html).
* `image_recipe_arn` - (Optional) Amazon Resource Name (ARN) of the image recipe.
* `image_scanning_configuration` - (Optional) Configuration block with image scanning configuration. Detailed below.
* `image_tests_configuration` - (Optional) Configuration block with image tests configuration. Detailed below.
* `logging_configuration` - (Optional) Configuration block with logging configuration. Detailed below.
* `schedule` - (Optional) Configuration block with schedule settings. Detailed below.
* `status` - (Optional) Status of the image pipeline. Valid values are `DISABLED` and `ENABLED`. Defaults to `ENABLED`.
* `workflow` - (Optional) Configuration block with the workflow configuration. Detailed below.
* `tags` - (Optional) Key-value map of resource tags for the image pipeline. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### image_scanning_configuration

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `image_scanning_enabled` - (Optional) Whether image scans are enabled. Defaults to `false`.
* `ecr_configuration` - (Optional) Configuration block with ECR configuration for image scanning. Detailed below.

### ecr_configuration

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `container tags` - (Optional) list of tags to apply to scanned images
* `repository_name` - (Optional) The name of the repository to scan

### image_tests_configuration

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `image_tests_enabled` - (Optional) Whether image tests are enabled. Defaults to `true`.
* `timeout_minutes` - (Optional) Number of minutes before image tests time out. Valid values are between `60` and `1440`. Defaults to `720`.

### logging_configuration

The following arguments are optional:

* `image_log_group_name` - (Optional) Name of the CloudWatch Log Group to send image logs to.
* `pipeline_log_group_name` - (Optional) Name of the CloudWatch Log Group to send pipeline logs to.

### schedule

The following arguments are required:

* `schedule_expression` - (Required) Cron expression of how often the pipeline start condition is evaluated. For example, `cron(0 0 * * ? *)` is evaluated every day at midnight UTC. Configurations using the five field syntax that was previously accepted by the API, such as `cron(0 0 * * *)`, must be updated to the six field syntax. For more information, see the [Image Builder User Guide](https://docs.aws.amazon.com/imagebuilder/latest/userguide/cron-expressions.html).

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `pipeline_execution_start_condition` - (Optional) Condition when the pipeline should trigger a new image build. Valid values are `EXPRESSION_MATCH_AND_DEPENDENCY_UPDATES_AVAILABLE` and `EXPRESSION_MATCH_ONLY`. Defaults to `EXPRESSION_MATCH_AND_DEPENDENCY_UPDATES_AVAILABLE`.

* `timezone` - (Optional) The timezone that applies to the scheduling expression. For example, "Etc/UTC", "America/Los_Angeles" in the [IANA timezone format](https://www.joda.org/joda-time/timezones.html). If not specified this defaults to UTC.

### workflow

The following arguments are required:

* `workflow_arn` - (Required) Amazon Resource Name (ARN) of the Image Builder Workflow.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `on_failure` - (Optional) The action to take if the workflow fails. Must be one of `CONTINUE` or `ABORT`.
* `parallel_group` - (Optional) The parallel group in which to run a test Workflow.
* `parameter` - (Optional) Configuration block for the workflow parameters. Detailed below.

### parameter

The following arguments are required:

* `name` - (Required) The name of the Workflow parameter.
* `value` - (Required) The value of the Workflow parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the image pipeline.
* `arn` - Amazon Resource Name (ARN) of the image pipeline.
* `date_created` - Date the image pipeline was created.
* `date_last_run` - Date the image pipeline was last run.
* `date_next_run` - Date the image pipeline will run next.
* `date_updated` - Date the image pipeline was updated.
* `platform` - Platform of the image pipeline.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_imagebuilder_image_pipeline.example
  identity = {
    "arn" = "arn:aws:imagebuilder:us-east-1:123456789012:image-pipeline/example"
  }
}

resource "aws_imagebuilder_image_pipeline" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the Image Builder image pipeline.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_imagebuilder_image_pipeline` resources using the Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_imagebuilder_image_pipeline.example
  id = "arn:aws:imagebuilder:us-east-1:123456789012:image-pipeline/example"
}
```

Using `terraform import`, import `aws_imagebuilder_image_pipeline` resources using the Amazon Resource Name (ARN). For example:

```console
% terraform import aws_imagebuilder_image_pipeline.example arn:aws:imagebuilder:us-east-1:123456789012:image-pipeline/example
```
