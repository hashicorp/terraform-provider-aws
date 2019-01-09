---
layout: "aws"
page_title: "AWS: aws_sagemaker_endpoint_configuration"
sidebar_current: "docs-aws-resource-sagemaker-endpoint-configuration"
description: |-
  Provides a SageMaker endpoint configuration resource.
---

# aws_sagemaker_endpoint_configuration

Provides a SageMaker endpoint configuration resource.

## Example Usage


Basic usage:

```hcl
resource "aws_sagemaker_endpoint_configuration" "ec" {
    name = "my-endpoint-config"

    production_variant {
        variant_name            = "variant-1"
        model_name              = "${aws_sagemaker_model.m.name}"
        initial_instance_count  = 1
        instance_type           = "ml.t2.medium"
    }

    tags {
        Name = "foo"
    }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the endpoint configuration. If omitted, Terraform will assign a random, unique name.
* `production_variants` - (Required) Fields are documented below.
* `kms_key_id` - (Optional) KMS key to encrypt the model data.
* `tags` - (Optional) A mapping of tags to assign to the resource.

The `production_variant` block supports:

* `variant_name` - (Optional) The name of the variant. If omitted, Terraform will assign a random, unique name.
* `model_name` - (Required) The name of the model to use.
* `initial_instance_count` - (Required) Initial number of instances used for auto-scaling.
* `instance_type` (Required) - The type of instance to start.
* `initial_variant_weight` (Optional) - Determines initial traffic distribution among all of the models that you specify in the endpoint configuration. If unspecified, it defaults to 1.0.
* `accelerator_type` (Optional) - The size of the Elastic Inference (EI) instance to use for the production variant.
## Attributes Reference

The following attributes are exported:

* `name` - The name of the endpoint configuration.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this endpoint configuration.

## Import

Endpoint configurations can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_endpoint_configuration.test_endpoint_config endpoint-config-foo
```
