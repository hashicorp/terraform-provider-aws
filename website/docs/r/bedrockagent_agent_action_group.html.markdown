---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_agent_action_group"
description: |-
  Terraform resource for managing an AWS Agents for Amazon Bedrock Agent Action Group.
---
# Resource: aws_bedrockagent_agent_action_group

Terraform resource for managing an AWS Agents for Amazon Bedrock Agent Action Group.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagent_agent_action_group" "example" {
  action_group_name          = "example"
  agent_id                   = "GGRRAED6JP"
  agent_version              = "DRAFT"
  skip_resource_in_use_check = true
  action_group_executor {
    lambda = "arn:aws:lambda:us-west-2:123456789012:function:example-function"
  }
  api_schema {
    payload = file("path/to/schema.yaml")
  }
}
```

### API Schema in S3 Bucket

```terraform
resource "aws_bedrockagent_agent_action_group" "example" {
  action_group_name          = "example"
  agent_id                   = "GGRRAED6JP"
  agent_version              = "DRAFT"
  skip_resource_in_use_check = true
  action_group_executor {
    lambda = "arn:aws:lambda:us-west-2:123456789012:function:example-function"
  }
  api_schema {
    s3 {
      s3_bucket_name = "example-bucket"
      s3_object_key  = "path/to/schema.json"
    }
  }
}
```

### Function Schema (Simplified Schema)

```terraform
resource "aws_bedrockagent_agent_action_group" "example" {
  action_group_name          = "example"
  agent_id                   = "GGRRAED6JP"
  agent_version              = "DRAFT"
  skip_resource_in_use_check = true
  action_group_executor {
    lambda = "arn:aws:lambda:us-west-2:123456789012:function:example-function"
  }
  function_schema {
    member_functions {
      functions {
        name        = "example-function"
        description = "Example function"
        parameters {
          map_block_key = "param1"
          type          = "string"
          description   = "The first parameter"
          required      = true
        }
        parameters {
          map_block_key = "param2"
          type          = "integer"
          description   = "The second parameter"
          required      = false
        }
      }
    }
  }
}
```

### Return of Control

```terraform
resource "aws_bedrockagent_agent_action_group" "example" {
  action_group_name          = "example"
  agent_id                   = "GGRRAED6JP"
  agent_version              = "DRAFT"
  skip_resource_in_use_check = true
  action_group_executor {
    custom_control = "RETURN_CONTROL"
  }
  api_schema {
    payload = file("path/to/schema.yaml")
  }
}
```

## Argument Reference

The following arguments are required:

* `action_group_name` - (Required) Name of the action group.
* `agent_id` - (Required) The unique identifier of the agent for which to create the action group.
* `agent_version` - (Required) Version of the agent for which to create the action group. Valid values: `DRAFT`.
* `action_group_executor` - (Required) ARN of the Lambda function containing the business logic that is carried out upon invoking the action or custom control method for handling the information elicited from the user. See [`action_group_executor` Block](#action_group_executor-block) for details.

The following arguments are optional:

* `action_group_state` - (Optional) Whether the action group is available for the agent to invoke or not when sending an [InvokeAgent](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_agent-runtime_InvokeAgent.html) request. Valid values: `ENABLED`, `DISABLED`.
* `api_schema` - (Optional) Either details about the S3 object containing the OpenAPI schema for the action group or the JSON or YAML-formatted payload defining the schema. For more information, see [Action group OpenAPI schemas](https://docs.aws.amazon.com/bedrock/latest/userguide/agents-api-schema.html). See [`api_schema` Block](#api_schema-block) for details.
* `description` - (Optional) Description of the action group.
* `function_schema` - (Optional) Describes the function schema for the action group.
  Each function represents an action in an action group.
  See [`function_schema` Block](#function_schema-block) for details.
* `parent_action_group_signature` - (Optional) To allow your agent to request the user for additional information when trying to complete a task, set this argument to `AMAZON.UserInput`. You must leave the `description`, `api_schema`, and `action_group_executor` arguments blank for this action group. Valid values: `AMAZON.UserInput`.
* `skip_resource_in_use_check` - (Optional) Whether the in-use check is skipped when deleting the action group.

### `action_group_executor` Block

The `action_group_executor` configuration block supports the following arguments:

* `custom_control` - (Optional) Custom control method for handling the information elicited from the user. Valid values: `RETURN_CONTROL`.
  To skip using a Lambda function and instead return the predicted action group, in addition to the parameters and information required for it, in the `InvokeAgent` response, specify `RETURN_CONTROL`.
  Only one of `custom_control` or `lambda` can be specified.
* `lambda` - (Optional) ARN of the Lambda function containing the business logic that is carried out upon invoking the action.
  Only one of `lambda` or `custom_control` can be specified.

### `api_schema` Block

The `api_schema` configuration block supports the following arguments:

* `payload` - (Optional) JSON or YAML-formatted payload defining the OpenAPI schema for the action group.
  Only one of `payload` or `s3` can be specified.
* `s3` - (Optional) Details about the S3 object containing the OpenAPI schema for the action group. See [`s3` Block](#s3-block) for details.
  Only one of `s3` or `payload` can be specified.

### `s3` Block

The `s3` configuration block supports the following arguments:

* `s3_bucket_name` - (Optional) Name of the S3 bucket.
* `s3_object_key` - (Optional) S3 object key containing the resource.

### `function_schema` Block

The `function_schema` configuration block supports the following arguments:

* `member_functions` - (Optional) Contains a list of functions.
  Each function describes and action in the action group.
  See [`member_functions` Block](#member_functions-block) for details.

### `member_functions` Block

The `member_functions` configuration block supports the following arguments:

* `functions` - (Optional) Functions that each define an action in the action group. See [`functions` Block](#functions-block) for details.

### `functions` Block

The `functions` configuration block supports the following arguments:

* `name` - (Required) Name for the function.
* `description` - (Optional) Description of the function and its purpose.
* `parameters` - (Optional) Parameters that the agent elicits from the user to fulfill the function. See [`parameters` Block](#parameters-block) for details.

### `parameters` Block

The `parameters` configuration block supports the following arguments:

* `map_block_key` - (Required) Name of the parameter.

  **Note:** The argument name `map_block_key` may seem out of context, but is necessary for backward compatibility reasons in the provider.
* `type` - (Required) Data type of the parameter. Valid values: `string`, `number`, `integer`, `boolean`, `array`.
* `description` - (Optional) Description of the parameter. Helps the foundation model determine how to elicit the parameters from the user.
* `required` - (Optional) Whether the parameter is required for the agent to complete the function for action group invocation.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `action_group_id` - Unique identifier of the action group.
- `id` - Action group ID, agent ID, and agent version separated by `,`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `delete` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Agent Action Group using the action group ID, the agent ID, and the agent version separated by `,`. For example:

```terraform
import {
  to = aws_bedrockagent_agent_action_group.example
  id = "MMAUDBZTH4,GGRRAED6JP,DRAFT"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Agent Action Group the action group ID, the agent ID, and the agent version separated by `,`. For example:

```console
% terraform import aws_bedrockagent_agent_action_group.example MMAUDBZTH4,GGRRAED6JP,DRAFT
```
