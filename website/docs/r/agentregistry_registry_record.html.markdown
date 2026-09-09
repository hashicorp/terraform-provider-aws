---
subcategory: "Agent Registry"
layout: "aws"
page_title: "AWS: aws_agentregistry_registry_record"
description: |-
  Manages an AWS Agent Registry registry record.
---

# Resource: aws_agentregistry_registry_record

Manages an AWS Agent Registry registry record.

A registry record represents a reusable agentic component in a registry, such as an MCP server, an A2A agent card, an agent skills definition, or a custom component.

## Example Usage

### Basic Usage

```terraform
resource "aws_agentregistry_registry" "example" {
  name = "example-registry"

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}

resource "aws_agentregistry_registry_record" "example" {
  registry_id = aws_agentregistry_registry.example.registry_id
  name        = "example-record"
  record_type = "CUSTOM"

  descriptors {
    custom {
      data = jsonencode({
        name = "example-component"
      })
    }
  }
}
```

### MCP Server Record

```terraform
resource "aws_agentregistry_registry_record" "example" {
  registry_id = aws_agentregistry_registry.example.registry_id
  name        = "example-mcp-server"
  record_type = "MCP"

  descriptors {
    mcp_server {
      data = jsonencode({
        name        = "example-mcp-server"
        description = "Example MCP server"
        remotes = [{
          type = "streamable-http"
          url  = "https://mcp.example.com/mcp"
        }]
      })
      data_schema_version = "2025-07-09"
    }
  }
}
```

### MCP Server Record Synchronized From a URL

```terraform
resource "aws_agentregistry_registry_record" "example" {
  registry_id = aws_agentregistry_registry.example.registry_id
  name        = "example-mcp-server"
  record_type = "MCP"

  descriptors {
    mcp_server {
      source {
        from_url {
          url = "https://mcp.example.com/.well-known/mcp.json"

          credential_provider_configuration {
            credential_provider_type = "IAM"

            credential_provider {
              iam_credential_provider {
                role_arn = aws_iam_role.example.arn
                service  = "execute-api"
              }
            }
          }
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `descriptors` - (Required) Typed descriptor content for the registry record. [See below](#descriptors-block).
* `name` - (Required) Name of the registry record. Names are unique within a registry.
* `record_type` - (Required) Type of the registry record, which determines the descriptor format. Valid values: `MCP`, `AGENT`, `SKILL`, `CUSTOM`.
* `registry_id` - (Required, Forces new resource) Identifier of the registry in which to create the record.

The following arguments are optional:

* `description` - (Optional) Description of the registry record.
* `display_name` - (Optional) Human-readable display name of the registry record.
* `record_version` - (Optional) Version of the registry record.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `descriptors` Block

The `descriptors` configuration block supports exactly one of the following arguments, matching the configured `record_type`:

* `a2a_agent_card` - (Optional) A2A agent card descriptor, used when `record_type` is `AGENT`. [See below](#a2a_agent_card-block).
* `agent_skills_definition` - (Optional) Agent skills definition descriptor, used when `record_type` is `SKILL`. [See below](#agent_skills_definition-block).
* `custom` - (Optional) Custom descriptor, used when `record_type` is `CUSTOM`. [See below](#custom-block).
* `mcp_server` - (Optional) MCP server descriptor, used when `record_type` is `MCP`. [See below](#mcp_server-block).

### `a2a_agent_card` Block

The `a2a_agent_card` configuration block supports the following arguments:

* `data` - (Optional) A2A agent card content, serialized as descriptor payload data.
* `data_schema_version` - (Optional) Schema version of the descriptor payload.
* `source` - (Optional) Source configuration used to synchronize the descriptor content. [See below](#source-block).

### `agent_skills_definition` Block

The `agent_skills_definition` configuration block supports the following arguments:

* `additional_data` - (Optional) Additional data associated with the agent skills definition. [See below](#agent_skills_definition-additional_data-block).
* `data` - (Optional) Agent skills definition content, serialized as descriptor payload data.
* `data_schema_version` - (Optional) Schema version of the descriptor payload.

### `agent_skills_definition` `additional_data` Block

The `additional_data` configuration block of `agent_skills_definition` supports the following arguments:

* `skill_md` - (Optional) Markdown skill content associated with the agent skills definition. [See below](#skill_md-block).

### `skill_md` Block

The `skill_md` configuration block supports the following arguments:

* `data` - (Optional) Agent skills markdown content, serialized as descriptor payload data.
* `data_schema_version` - (Optional) Schema version of the descriptor payload.
* `source` - (Optional) Source configuration used to synchronize the markdown content. [See below](#source-block).

### `custom` Block

The `custom` configuration block supports the following arguments:

* `data` - (Optional) Descriptor payload data.

### `mcp_server` Block

The `mcp_server` configuration block supports the following arguments:

* `additional_data` - (Optional) Additional data associated with the MCP server descriptor. [See below](#mcp_server-additional_data-block).
* `data` - (Optional) MCP server descriptor content, serialized as descriptor payload data.
* `data_schema_version` - (Optional) Schema version of the descriptor payload.
* `source` - (Optional) Source configuration used to synchronize the descriptor content. [See below](#source-block).

### `mcp_server` `additional_data` Block

The `additional_data` configuration block of `mcp_server` supports the following arguments:

* `tools` - (Optional) MCP tools descriptor containing tool definitions. [See below](#tools-block).

### `tools` Block

The `tools` configuration block supports the following arguments:

* `data` - (Optional) Descriptor payload data.
* `data_schema_version` - (Optional) Schema version of the descriptor payload.

### `source` Block

The `source` configuration block supports the following arguments:

* `from_url` - (Required) URL-based descriptor source. [See below](#from_url-block).

### `from_url` Block

The `from_url` configuration block supports the following arguments:

* `credential_provider_configuration` - (Optional) Credential providers used to authenticate when fetching descriptor content from the source URL. [See below](#credential_provider_configuration-block).
* `url` - (Required) URL from which the descriptor content is retrieved.

### `credential_provider_configuration` Block

The `credential_provider_configuration` configuration block supports the following arguments:

* `credential_provider` - (Required) Credential provider details corresponding to the specified credential provider type. [See below](#credential_provider-block).
* `credential_provider_type` - (Required) Type of credential provider. Valid values: `OAUTH`, `IAM`.

### `credential_provider` Block

The `credential_provider` configuration block supports exactly one of the following arguments:

* `iam_credential_provider` - (Optional) IAM role credential provider details. [See below](#iam_credential_provider-block).
* `oauth_credential_provider` - (Optional) OAuth 2.0 credential provider details. [See below](#oauth_credential_provider-block).

### `iam_credential_provider` Block

The `iam_credential_provider` configuration block supports the following arguments:

* `region` - (Optional) AWS Region to use for request signing. If not specified, the Region is derived from the source URL hostname, falling back to the Region of the registry.
* `role_arn` - (Optional) ARN of the IAM role to assume for request signing.
* `service` - (Optional) Service name to use for request signing, such as `execute-api`.

### `oauth_credential_provider` Block

The `oauth_credential_provider` configuration block supports the following arguments:

* `custom_parameters` - (Optional) Additional parameters to include in the OAuth 2.0 token request.
* `grant_type` - (Optional) OAuth 2.0 grant type used to obtain access tokens. Valid values: `CLIENT_CREDENTIALS`.
* `provider_arn` - (Required) ARN of the OAuth 2.0 credential provider resource in Amazon Bedrock AgentCore Identity.
* `scopes` - (Optional) OAuth 2.0 scopes to request when obtaining access tokens.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `record_arn` - ARN of the registry record.
* `record_id` - Unique identifier of the registry record.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_agentregistry_registry_record.example
  identity = {
    registry_id = "registry-id-12345678"
    record_id   = "record-id-12345678"
  }
}

resource "aws_agentregistry_registry_record" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `record_id` (String) Record ID.
* `registry_id` (String) Registry ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Agent Registry Registry Record using the registry ID and record ID, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_agentregistry_registry_record.example
  id = "registry-id-12345678,record-id-12345678"
}
```

Using `terraform import`, import an Agent Registry Registry Record using the registry ID and record ID, separated by a comma (`,`). For example:

```console
% terraform import aws_agentregistry_registry_record.example registry-id-12345678,record-id-12345678
```
