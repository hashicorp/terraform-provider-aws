---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_notebook_instance_lifecycle_configuration"
description: |-
  Provides a lifecycle configuration for SageMaker AI Notebook Instances.
---

# Resource: aws_sagemaker_notebook_instance_lifecycle_configuration

Provides a lifecycle configuration for SageMaker AI Notebook Instances.

## Example Usage

Usage:

```terraform
resource "aws_sagemaker_notebook_instance_lifecycle_configuration" "lc" {
  name      = "foo"
  on_create = base64encode("echo foo")
  on_start  = base64encode("echo bar")
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Optional) The name of the lifecycle configuration (must be unique). If omitted, Terraform will assign a random, unique name.
* `on_create` - (Optional) A shell script (base64-encoded) that runs only once when the SageMaker AI Notebook Instance is created.
* `on_start` - (Optional) A shell script (base64-encoded) that runs every time the SageMaker AI Notebook Instance is started including the time it's created.
* `tags` - (Optional) A mapping of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this lifecycle configuration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import models using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_notebook_instance_lifecycle_configuration.lc
  id = "foo"
}
```

Using `terraform import`, import models using the `name`. For example:

```console
% terraform import aws_sagemaker_notebook_instance_lifecycle_configuration.lc foo
```
