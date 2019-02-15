---
layout: "aws"
page_title: "AWS: aws_sagemaker_lifecycle_config"
sidebar_current: "docs-aws-resource-sagemaker-lifecycle-configuration"
description: |-
  Provides a lifecycle configuration for SageMaker Notebook instances.
---

# aws_sagemaker_lifecycle_config

Provides a lifecycle configuration for SageMaker Notebook instances.

## Example Usage

Usage:

```hcl
resource "aws_sagemaker_lifecycle_config" "lc" {
  name = "foo"
  on_create = "${base64encode('echo foo')}"
  on_start = "${base64encode('echo bar')}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the model (must be unique). If omitted, Terraform will assign a random, unique name.
* `on_create` - (Optional) A shell script (base64-encoded) that runs only once when the notebook instance is created.
* `on_start` - (Required) A shell script (base64-encoded) that runs every time the notebook is started including the time it's created.

## Attributes Reference

The following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this model.

## Import

Models can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_lifecycle_config.lc foo
```
