---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_lifecycle_policy"
description: |-
  Manages an Image Builder Lifecycle Policy
---

# Resource: aws_imagebuilder_lifecycle_policy

Manages an Image Builder Lifecycle Policy.

## Example Usage

```terraform
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role" "example" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "imagebuilder.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
  name = "example"
}

resource "aws_iam_role_policy_attachment" "example" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/EC2ImageBuilderLifecycleExecutionPolicy"
  role       = aws_iam_role.example.name
}

resource "aws_imagebuilder_lifecycle_policy" "example" {
  name           = "name"
  description    = "Example description"
  execution_role = aws_iam_role.example.arn
  resource_type  = "AMI_IMAGE"
  policy_details {
    action {
      type = "DELETE"
    }
    filter {
      type            = "AGE"
      value           = 6
      retain_at_least = 10
      unit            = "YEARS"
    }
  }
  resource_selection {
    tag_map = {
      "key1" = "value1"
      "key2" = "value2"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.example]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the lifecycle policy to create.
* `resource_type` - (Required) The type of Image Builder resource that the lifecycle policy applies to. Valid values: `AMI_IMAGE` or `CONTAINER_IMAGE`.
* `execution_role` - (Required) The name or Amazon Resource Name (ARN) for the IAM role you create that grants Image Builder access to run lifecycle actions. More information about this role can be found [`here`](https://docs.aws.amazon.com/imagebuilder/latest/userguide/image-lifecycle-prerequisites.html#image-lifecycle-prereq-role).
* `policy_details` - (Required) Configuration block with policy details. Detailed below.
* `resource_selection` - (Required) Selection criteria for the resources that the lifecycle policy applies to. Detailed below.

The following arguments are optional:

* `description` - (Optional) description for the lifecycle policy.
* `tags` - (Optional) Key-value map of resource tags for the Image Builder Lifecycle Policy. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### policy_details

The following arguments are required:

* `action` - (Required) Configuration details for the policy action.
* `filter` - (Required) Specifies the resources that the lifecycle policy applies to.

The following arguments are optional:

* `exclusion_rules` - (Optional) Additional rules to specify resources that should be exempt from policy actions.

### action

The following arguments are required:

* `type` - (Required) Specifies the lifecycle action to take. Valid values: `DELETE`, `DEPRECATE` or `DISABLE`.

The following arguments are optional:

* `include_resources` - (Optional) Specifies the resources that the lifecycle policy applies to. Detailed below.

### include_resources

The following arguments are optional:

* `amis` - (Optional) Specifies whether the lifecycle action should apply to distributed AMIs.
* `containers` - (Optional) Specifies whether the lifecycle action should apply to distributed containers.
* `snapshots` - (Optional) Specifies whether the lifecycle action should apply to snapshots associated with distributed AMIs.

### filter

The following arguments are required:

* `type` - (Required) Filter resources based on either age or count. Valid values: `AGE` or `COUNT`.
* `value` - (Required) The number of units for the time period or for the count. For example, a value of 6 might refer to six months or six AMIs.

The following arguments are optional:

* `retain_at_least` - (Optional) For age-based filters, this is the number of resources to keep on hand after the lifecycle DELETE action is applied. Impacted resources are only deleted if you have more than this number of resources. If you have fewer resources than this number, the impacted resource is not deleted.
* `unit` - (Optional) Defines the unit of time that the lifecycle policy uses to determine impacted resources. This is required for age-based rules. Valid values: `DAYS`, `WEEKS`, `MONTHS` or `YEARS`.

### exclusion_rules

The following arguments are optional:

* `amis` - (Optional) Lists configuration values that apply to AMIs that Image Builder should exclude from the lifecycle action. Detailed below.
* `tag_map` - (Optional) Contains a list of tags that Image Builder uses to skip lifecycle actions for Image Builder image resources that have them.

### amis

The following arguments are optional:

* `is_public` - (Optional) Configures whether public AMIs are excluded from the lifecycle action.
* `last_launched` - (Optional) Specifies configuration details for Image Builder to exclude the most recent resources from lifecycle actions. Detailed below.
* `regions` - (Optional) Configures AWS Regions that are excluded from the lifecycle action.
* `shared_accounts` - Specifies AWS accounts whose resources are excluded from the lifecycle action.
* `tag_map` - (Optional) Lists tags that should be excluded from lifecycle actions for the AMIs that have them.

### last_launched

The following arguments are required:

* `unit` - (Required) Defines the unit of time that the lifecycle policy uses to calculate elapsed time since the last instance launched from the AMI. For example: days, weeks, months, or years. Valid values: `DAYS`, `WEEKS`, `MONTHS` or `YEARS`.
* `value` - (Required) The integer number of units for the time period. For example 6 (months).

### resource_selection

The following arguments are optional:

* `recipes` - (Optional) A list of recipes that are used as selection criteria for the output images that the lifecycle policy applies to. Detailed below.
* `tag_map` - (Optional) A list of tags that are used as selection criteria for the Image Builder image resources that the lifecycle policy applies to.

### recipes

The following arguments are required:

* `name` - (Required) The name of an Image Builder recipe that the lifecycle policy uses for resource selection.
* `semantic_version` - (Required) The version of the Image Builder recipe specified by the name field.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the lifecycle policy.
* `arn` - Amazon Resource Name (ARN) of the lifecycle policy.
* `status` - The status of the lifecycle policy.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_imagebuilder_lifecycle_policy` using the Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_imagebuilder_lifecycle_policy.example
  id = "arn:aws:imagebuilder:us-east-1:123456789012:lifecycle-policy/example"
}
```

Using `terraform import`, import `aws_imagebuilder_lifecycle_policy` using the Amazon Resource Name (ARN). For example:

```console
% terraform import aws_imagebuilder_lifecycle_policy.example arn:aws:imagebuilder:us-east-1:123456789012:lifecycle-policy/example
```
