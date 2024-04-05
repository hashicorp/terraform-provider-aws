---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_model_package_group"
description: |-
  Provides a SageMaker Model Package Group resource.
---

# Resource: aws_sagemaker_model_package_group

Provides a SageMaker Model Package Group resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_model_package_group" "example" {
  model_package_group_name = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `model_package_group_name` - (Required) The name of the model group.
* `model_package_group_description` - (Optional) A description for the model group.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the Model Package Group.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Model Package Group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker Model Package Groups using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_model_package_group.test_model_package_group
  id = "my-code-repo"
}
```

Using `terraform import`, import SageMaker Model Package Groups using the `name`. For example:

```console
% terraform import aws_sagemaker_model_package_group.test_model_package_group my-code-repo
```
