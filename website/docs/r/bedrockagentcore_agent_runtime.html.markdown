---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_agent_runtime"
description: |-
  Manages an AWS Bedrock AgentCore Agent Runtime.
---

# Resource: aws_bedrockagentcore_agent_runtime

Manages an AWS Bedrock AgentCore Agent Runtime. Agent Runtime provides a containerized execution environment for AI agents.

## Example Usage

### Basic Usage

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

data "aws_iam_policy_document" "ecr_permissions" {
  statement {
    actions   = ["ecr:GetAuthorizationToken"]
    effect    = "Allow"
    resources = ["*"]
  }

  statement {
    actions = [
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer"
    ]
    effect    = "Allow"
    resources = [aws_ecr_repository.example.arn]
  }
}

resource "aws_iam_role" "example" {
  name               = "bedrock-agentcore-runtime-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy" "example" {
  role   = aws_iam_role.example.id
  policy = data.aws_iam_policy_document.ecr_permissions.json
}

resource "aws_bedrockagentcore_agent_runtime" "example" {
  name     = "example-agent-runtime"
  role_arn = aws_iam_role.example.arn

  artifact {
    container_configuration {
      container_uri = "${aws_ecr_repository.example.repository_url}:latest"
    }
  }

  network_configuration = {
    network_mode = "PUBLIC"
  }
}
```

### MCP Server With Custom JWT Authorizer

```terraform
resource "aws_bedrockagentcore_agent_runtime" "example" {
  name        = "example-agent-runtime"
  description = "Agent runtime with JWT authorization"
  role_arn    = aws_iam_role.example.arn

  artifact {
    container_configuration {
      container_uri = "${aws_ecr_repository.example.repository_url}:v1.0"
    }
  }

  environment_variables = {
    LOG_LEVEL = "INFO"
    ENV       = "production"
  }

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["my-app", "mobile-app"]
      allowed_clients  = ["client-123", "client-456"]
    }
  }

  network_configuration = {
    network_mode = "PUBLIC"
  }

  protocol_configuration {
    server_protocol = "MCP"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the agent runtime.
* `role_arn` - (Required) ARN of the IAM role that the agent runtime assumes to access AWS services.
* `artifact` - (Required) Container artifact configuration. See [`artifact`](#artifact) below.
* `network_configuration` - (Required) Network configuration for the agent runtime. See [`network_configuration`](#network_configuration) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the agent runtime.
* `environment_variables` - (Optional) Map of environment variables to pass to the container.
* `client_token` - (Optional) Unique identifier for request idempotency. If not provided, one will be generated automatically.
* `authorizer_configuration` - (Optional) Authorization configuration for authenticating incoming requests. See [`authorizer_configuration`](#authorizer_configuration) below.
* `protocol_configuration` - (Optional) Protocol configuration for the agent runtime. See [`protocol_configuration`](#protocol_configuration) below.

### `artifact`

The `artifact` block supports the following:

* `container_configuration` - (Required) Container configuration block. See [`container_configuration`](#container_configuration) below.

### `container_configuration`

The `container_configuration` block supports the following:

* `container_uri` - (Required) URI of the container image in Amazon ECR.

### `authorizer_configuration`

The `authorizer_configuration` block supports the following:

* `custom_jwt_authorizer` - (Optional) JWT-based authorization configuration block. See [`custom_jwt_authorizer`](#custom_jwt_authorizer) below.

### `custom_jwt_authorizer`

The `custom_jwt_authorizer` block supports the following:

* `discovery_url` - (Required) URL used to fetch OpenID Connect configuration or authorization server metadata. Must end with `.well-known/openid-configuration`.
* `allowed_audience` - (Optional) Set of allowed audience values for JWT token validation.
* `allowed_clients` - (Optional) Set of allowed client IDs for JWT token validation.

### `network_configuration`

The `network_configuration` object supports the following:

* `network_mode` - (Required) Network mode for the agent runtime. Valid values: `PUBLIC`.

### `protocol_configuration`

The `protocol_configuration` block supports the following:

* `server_protocol` - (Optional) Server protocol for the agent runtime. Valid values: `HTTP`, `MCP`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Agent Runtime.
* `id` - Unique identifier of the Agent Runtime.
* `version` - Version of the Agent Runtime.
* `workload_identity_details` - Workload identity details for the agent runtime. See [`workload_identity_details`](#workload_identity_details) below.

### `workload_identity_details`

The `workload_identity_details` block contains the following:

* `workload_identity_arn` - ARN of the workload identity.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Agent Runtime using the agent runtime ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_agent_runtime.example
  id = "AGENT1234567890"
}
```

Using `terraform import`, import Bedrock AgentCore Agent Runtime using the agent runtime ID. For example:

```console
% terraform import aws_bedrockagentcore_agent_runtime.example AGENT1234567890
```
