---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_code_interpreter"
description: |-
  Manages an AWS Bedrock AgentCore Code Interpreter.
---

# Resource: aws_bedrockagentcore_code_interpreter

Manages an AWS Bedrock AgentCore Code Interpreter. Code Interpreter provides a secure environment for AI agents to execute Python code, enabling data analysis, calculations, and file processing capabilities.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_code_interpreter" "example" {
  name        = "example-code-interpreter"
  description = "Code interpreter for data analysis"

  network_configuration {
    network_mode = "PUBLIC"
  }
}
```

### Code Interpreter with Execution Role

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "bedrock-agentcore-code-interpreter-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_bedrockagentcore_code_interpreter" "example" {
  name               = "example-code-interpreter"
  description        = "Code interpreter with custom execution role"
  execution_role_arn = aws_iam_role.example.arn

  network_configuration {
    network_mode = "SANDBOX"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the code interpreter.
* `network_configuration` - (Required) Network configuration for the code interpreter. See [`network_configuration`](#network_configuration) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the code interpreter.
* `execution_role_arn` - (Optional) ARN of the IAM role that the code interpreter assumes for execution. Required when using `SANDBOX` network mode.
* `client_token` - (Optional) Unique identifier for request idempotency. If not provided, one will be generated automatically.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `network_configuration`

The `network_configuration` object supports the following:

* `network_mode` - (Required) Network mode for the code interpreter. Valid values: `PUBLIC`, `SANDBOX`, `VPC`.
* `vpc_config` - (Optional) VPC configuration. See [`vpc_config`](#vpc_config) below.

### `vpc_config`

The `vpc_config` block supports the following:

* `security_groups` - (Required) Security groups associated with the VPC configuration.
* `subnets` - (Required) Subnets associated with the VPC configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `code_interpreter_arn` - ARN of the Code Interpreter.
* `code_interpreter_id` - Unique identifier of the Code Interpreter.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Code Interpreter using the code interpreter ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_code_interpreter.example
  id = "CODEINTERPRETER1234567890"
}
```

Using `terraform import`, import Bedrock AgentCore Code Interpreter using the code interpreter ID. For example:

```console
% terraform import aws_bedrockagentcore_code_interpreter.example CODEINTERPRETER1234567890
```
