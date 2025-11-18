---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_prompt_routers"
description: |-
  Terraform data source for managing AWS Bedrock Prompt Routers.
---

# Data Source: aws_bedrock_prompt_routers

Terraform data source for managing AWS Bedrock Prompt Routers.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_prompt_routers" "example" {}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

- `prompt_router_summaries` - List of prompt router summary objects. See [`prompt_router_summaries`](#prompt_router_summaries).

### `prompt_router_summaries`

- `created_at` - Time at which the prompt router was created.
- `description` - Description of the prompt router.
- `fallback_model` - List of information about the fallback model. See [`fallback_model`](#fallback_model).
- `models` - List of information about each model in the prompt router. See [`models`](#models).
- `prompt_router_arn` - Amazon Resource Name (ARN) of the prompt router.
- `prompt_router_name` - Name of the prompt router.
- `routing_criteria` - Routing criteria for the prompt router. See [`routing_criteria`](#routing_criteria).
- `status` - Status of the prompt router. `AVAILABLE` means that the prompt router is ready to route requests.
- `type` - Type of the prompt router.
- `updated_at` - Time at which the prompt router was last updated.

### `fallback_model`

- `model_arn` - Amazon Resource Name (ARN) of the fallback model.

### `models`

- `model_arn` - Amazon Resource Name (ARN) of the model.

### `routing_criteria`

- `response_quality_difference` - Threshold for the difference in quality between model responses.
