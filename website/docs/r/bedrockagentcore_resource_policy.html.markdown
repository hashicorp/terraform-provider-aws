---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_resource_policy"
description: |-
  Manages an AWS Bedrock AgentCore Resource Policy.
---

# Resource: aws_bedrockagentcore_resource_policy

Manages an AWS Bedrock AgentCore Resource Policy. Resource-based policies in Amazon Bedrock AgentCore allow you to control which principals (AWS accounts, IAM users, or IAM roles) can invoke and manage your Amazon Bedrock AgentCore Runtime and Gateway resources.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_agent_runtime" "example" {
  ### Configuration omitted for brevity ###
}

data "aws_iam_policy_document" "example" {
  statement {
    sid    = "AllowOAuthFromVPC"
    effect = "Allow"
    actions = [
      "bedrock-agentcore:InvokeAgentRuntime",
    ]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    resources = [
      aws_bedrockagentcore.agent_runtime.example.agent_runtime_arn
    ]
    condition {
      test     = "StringEquals"
      variable = "aws:SourceVpc"
      values   = ["vpc-1a2b3c4d"]
    }
  }
}

resource "aws_bedrockagentcore_resource_policy" "example" {
  resource_arn = aws_bedrockagentcore_agent_runtime.example.agent_runtime_arn
  policy       = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` - (Required) Amazon Resource Name (ARN) of the resource for which to create or update the resource policy.
* `policy` - (Required) Resource policy definition

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_resource_policy.example
  identity = {
    resource_arn = "<runtime or gateway ARN>"
  }
}

resource "aws_bedrockagentcore_resource_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `resource_arn` - ARN of the resource to which the Resource Policy is attached.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Resource Policy using the `resource_arn`. For example:

```terraform
import {
  to = aws_bedrockagentcore_resource_policy.example
  id = "<runtime or gateway ARN>"
}
```

Using `terraform import`, import Bedrock AgentCore Resource Policy using the `resource_arn`. For example:

```console
% terraform import aws_bedrockagentcore_resource_policy.example <runtime or gateway ARN>
```
