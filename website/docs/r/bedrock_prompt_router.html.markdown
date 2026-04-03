---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_prompt_router"
description: |-
  Manages an AWS Bedrock Prompt Router.
---

# Resource: aws_bedrock_prompt_router

Manages an AWS Bedrock Prompt Router that routes requests between multiple foundation models based on routing criteria.

## Example Usage

### Basic Usage

```terraform
data "aws_region" "current" {}

resource "aws_bedrock_prompt_router" "example" {
  prompt_router_name = "example-router"
  description        = "Example prompt router for intelligent routing"

  fallback_model {
    model_arn = "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/anthropic.claude-3-5-sonnet-20241022-v2:0"
  }

  models {
    model_arn = "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/anthropic.claude-3-5-sonnet-20241022-v2:0"
  }

  models {
    model_arn = "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/anthropic.claude-3-5-haiku-20241022-v1:0"
  }

  routing_criteria {
    response_quality_difference = 25
  }

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `prompt_router_name` - (Required) Name of the prompt router. Must be unique within your AWS account in the current region. Must be between 1 and 64 characters.
* `fallback_model` - (Required) Default model to use when the routing criteria is not met. See [`fallback_model`](#fallback_model).
* `models` - (Required) List of foundation models that the prompt router can route requests to. At least one model must be specified. See [`models`](#models).
* `routing_criteria` - (Required) Criteria used to determine how incoming requests are routed to different models. See [`routing_criteria`](#routing_criteria).

The following arguments are optional:

* `description` - (Optional) Description of the prompt router to help identify its purpose. Must be between 1 and 200 characters.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value mapping of resource tags for the prompt router.

### `fallback_model`

* `model_arn` - (Required) Amazon Resource Name (ARN) of the fallback model.

### `models`

* `model_arn` - (Required) Amazon Resource Name (ARN) of the model.

### `routing_criteria`

* `response_quality_difference` - (Required) Threshold for the difference in quality between model responses. Must be a value between 0 and 100, and must be a multiple of 5 (e.g., 0, 5, 10, 15, ..., 100).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `prompt_router_arn` - Amazon Resource Name (ARN) of the prompt router.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Prompt Router using the `prompt_router_arn`. For example:

```terraform
import {
  to = aws_bedrock_prompt_router.example
  id = "arn:aws:bedrock:us-east-1:123456789012:default-prompt-router/example-router"
}
```

Using `terraform import`, import Bedrock Prompt Router using the `prompt_router_arn`. For example:

```console
% terraform import aws_bedrock_prompt_router.example arn:aws:bedrock:us-east-1:123456789012:default-prompt-router/example-router
```
