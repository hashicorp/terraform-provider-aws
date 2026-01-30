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
  agent_runtime_name = "example_agent_runtime"
  role_arn           = aws_iam_role.example.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = "${aws_ecr_repository.example.repository_url}:latest"
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}
```

### MCP Server With Custom JWT Authorizer

```terraform
resource "aws_bedrockagentcore_agent_runtime" "example" {
  agent_runtime_name = "example_agent_runtime"
  description        = "Agent runtime with JWT authorization"
  role_arn           = aws_iam_role.example.arn

  agent_runtime_artifact {
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

  network_configuration {
    network_mode = "PUBLIC"
  }

  protocol_configuration {
    server_protocol = "MCP"
  }
}
```

### Agent runtime artifact from S3 with Code Configuration

```terraform
resource "aws_bedrockagentcore_agent_runtime" "example" {
  agent_runtime_name = "example_agent_runtime"
  role_arn           = aws_iam_role.example.arn

  agent_runtime_artifact {
    code_configuration {
      entry_point = ["main.py"]
      runtime     = "PYTHON_3_13"
      code {
        s3 {
          bucket = "example-bucket"
          prefix = "example-agent-runtime-code.zip"
        }
      }
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}
```

## Argument Reference

The following arguments are required:

* `agent_runtime_name` - (Required) Name of the agent runtime.
* `role_arn` - (Required) ARN of the IAM role that the agent runtime assumes to access AWS services.
* `agent_runtime_artifact` - (Required) Container artifact configuration. See [`agent_runtime_artifact`](#agent-runtime-artifact) below.
* `network_configuration` - (Required) Network configuration for the agent runtime. See [`network_configuration`](#network_configuration) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the agent runtime.
* `environment_variables` - (Optional) Map of environment variables to pass to the container.
* `authorizer_configuration` - (Optional) Authorization configuration for authenticating incoming requests. See [`authorizer_configuration`](#authorizer_configuration) below.
* `lifecycle_configuration` - (Optional) Runtime session and resource lifecycle configuration for the agent runtime. See [`lifecycle_configuration`](#lifecycle_configuration) below.
* `protocol_configuration` - (Optional) Protocol configuration for the agent runtime. See [`protocol_configuration`](#protocol_configuration) below.
* `request_header_configuration` - (Optional) Configuration for HTTP request headers that will be passed through to the runtime. See [`request_header_configuration`](#request_header_configuration) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `agent_runtime_artifact`

The `agent_runtime_artifact` block supports the following:

* `code_configuration` - (Optional) Code configuration block for the agent runtime artifact, including the source code location and execution settings. Exactly one of `code_configuration` or `container_configuration` must be specified. See [`code_configuration`](#code_configuration) below.
* `container_configuration` - (Optional) Container configuration block for the agent artifact. Exactly one of `code_configuration` or `container_configuration` must be specified. See [`container_configuration`](#container_configuration) below.

### `code_configuration`

The `code_configuration` block supports the following:

* `code` - (Required) Configuration block for the source code location and configuration details. See [`code`](#code) below.
* `entry_point` - (Required) Array specifying the entry point for code execution, indicating the function or method to invoke when the code runs. The array must contain 1 or 2 elements. Examples: `["main.py"]`, `["opentelemetry-instrument", "main.py"]`.
* `runtime` - (Required) Runtime environment used to execute the code. Valid values: `PYTHON_3_10`, `PYTHON_3_11`, `PYTHON_3_12`, `PYTHON_3_13`.

### `code`

The `code` block supports the following:

* `s3` - (Required) Configuration block for the Amazon S3 object that contains the source code for the agent runtime. See [`s3`](#s3) below.

### `s3`

The `s3` block supports the following:

* `bucket` - (Required) Name of the Amazon S3 bucket.
* `prefix` - (Required) Key of the object containing the ZIP file of the source code for the agent runtime in the Amazon S3 bucket.
* `version_id` - (Optional) Version ID of the Amazon S3 object. If not specified, the latest version of the object is used.

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

### `lifecycle_configuration`

The `lifecycle_configuration` block supports the following:

* `idle_runtime_session_timeout` - (Optional) Timeout in seconds for idle runtime sessions.
* `max_lifetime` - (Optional) Maximum lifetime for the instance in seconds.

### `network_configuration`

The `network_configuration` block supports the following:

* `network_mode` - (Required) Network mode for the agent runtime. Valid values: `PUBLIC`, `VPC`.
* `network_mode_config` - (Optional) Network mode configuration. See [`network_mode_config`](#network_mode_config) below.

### `network_mode_config`

The `network_mode_config` block supports the following:

* `security_groups` - (Required) Security groups associated with the VPC configuration.
* `subnets` - (Required) Subnets associated with the VPC configuration.

### `protocol_configuration`

The `protocol_configuration` block supports the following:

* `server_protocol` - (Optional) Server protocol for the agent runtime. Valid values: `HTTP`, `MCP`, `A2A`.

### `request_header_configuration`

The `request_header_configuration` block supports the following:

* `request_header_allowlist` - (Optional) A list of HTTP request headers that are allowed to be passed through to the runtime.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `agent_runtime_arn` - ARN of the Agent Runtime.
* `agent_runtime_id` - Unique identifier of the Agent Runtime.
* `agent_runtime_version` - Version of the Agent Runtime.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
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

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Agent Runtime using `agent_runtime_id`. For example:

```terraform
import {
  to = aws_bedrockagentcore_agent_runtime.example
  id = "agent-runtime-12345"
}
```

Using `terraform import`, import Bedrock AgentCore Agent Runtime using `agent_runtime_id`. For example:

```console
% terraform import aws_bedrockagentcore_agent_runtime.example agent-runtime-12345
```
