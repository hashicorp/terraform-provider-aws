---
subcategory: "Agents for Amazon Bedrock"
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
  agent_id                   = "ABDJFOWER1"
  agent_version              = "DRAFT"
  skip_resource_in_use_check = true
  action_group_executor {
    lambda = "arn:aws:lambda:us-east-1:123456789012:function:example-function"
  }
  api_schema {
    s3 {
      s3_bucket_name = "example-bucket"
      s3_object_key  = "path/to/schema.json"
    }
  }
}

```

## Argument Reference

The following arguments are required:

* `action_group_name` - (Required) Name of the Agent Action Group.
* `agent_id` - (Required) Id of the Agent for the Action Group.
* `agent_version` - (Required) Version of the Agent to attach the Action Group to.
* `action_group_executor` - (Required) Configuration of the executor for the Action Group.
* `api_schema` - (Required) Configuration of the API Schema for the Action Group.

### action_group_executor

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

The following arguments are required:

* `lambda` - (Required) ARN of the Lambda that defines the business logic for the action group.

### api_schema

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

The following arguments are optional:

* `payload` - (Optional) YAML or JSON OpenAPI Schema.
* `s3` - (Optional) Configuration of S3 schema location

### s3

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

The following arguments are optional:

* `s3_bucket_name` - (Required) The S3 bucket name that contains the OpenAPI Schema.
* `s3_object_key` - (Required) The S3 Object Key for the OpenAPI Schema in the S3 Bucket.

The following arguments are optional:

* `action_group_state` - (Optional) `ENABLED` or `DISABLED`
* `description` - (Optional) Description of the Agent Action Group.
* `skip_resource_in_use_check` - (Optional) Set to true to skip the in-use check when deleting.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Agent Action Group using the `ABDJFOWER1,HSKTNKANI4,DRAFT`. For example:

```terraform
import {
  to = aws_bedrockagent_agent_action_group.example
  id = "ABDJFOWER1,HSKTNKANI4,DRAFT"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Agent Action Group using the `example_id_arg`. For example:

```console
% terraform import aws_bedrockagent_agent_action_group.example ABDJFOWER1,HSKTNKANI4,DRAFT
```
