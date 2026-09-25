---
subcategory: "Agent Registry"
layout: "aws"
page_title: "AWS: aws_agentregistry_registry_record"
description: |-
  Provides details about an AWS Agent Registry registry record.
---

# Data Source: aws_agentregistry_registry_record

Provides details about an AWS Agent Registry registry record.

## Example Usage

### Basic Usage

```terraform
data "aws_agentregistry_registry_record" "example" {
  registry_id = "registry-id-12345678"
  record_id   = "record-id-12345678"
}
```

### Lookup by ARN

```terraform
data "aws_agentregistry_registry_record" "example" {
  registry_id = "arn:aws:agent-registry:us-west-2:123456789012:registry/registry-id-12345678"
  record_id   = "arn:aws:agent-registry:us-west-2:123456789012:registry/registry-id-12345678/record/record-id-12345678"
}
```

## Argument Reference

The following arguments are required:

* `record_id` - (Required) Identifier of the registry record to retrieve. Accepts a record ID or ARN.
* `registry_id` - (Required) Identifier of the registry containing the record. Accepts a registry ID or ARN.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `created_at` - Timestamp when the registry record was created.
* `description` - Description of the registry record.
* `descriptors` - Typed descriptors that define the content of the registry record. [See below](#descriptors-block).
* `display_name` - Human-readable display name of the registry record.
* `name` - Name of the registry record.
* `record_arn` - ARN of the registry record.
* `record_type` - Type of the registry record. Valid values: `MCP`, `AGENT`, `SKILL`, `CUSTOM`.
* `record_version` - Version identifier of the registry record.
* `registry_arn` - ARN of the parent registry that owns the record.
* `status` - Lifecycle status of the registry record. Valid values: `DRAFT`, `PENDING_APPROVAL`, `APPROVED`, `REJECTED`, `DEPRECATED`, `CREATING`, `UPDATING`, `CREATE_FAILED`, `UPDATE_FAILED`.
* `status_reason` - Reason for the current status. Typically populated when the status indicates a failure state.
* `tags` - Map of tags assigned to the registry record.
* `updated_at` - Timestamp when the registry record was last updated.

### `descriptors` Block

* `a2a_agent_card` - A2A agent card descriptor, populated when `record_type` is `AGENT`. [See below](#a2a_agent_card-block).
* `agent_skills_definition` - Agent skills definition descriptor, populated when `record_type` is `SKILL`. [See below](#agent_skills_definition-block).
* `custom` - Custom descriptor, populated when `record_type` is `CUSTOM`. [See below](#custom-block).
* `mcp_server` - MCP server descriptor, populated when `record_type` is `MCP`. [See below](#mcp_server-block).

### `a2a_agent_card` Block

* `data` - A2A agent card content, serialized as descriptor payload data.
* `data_schema_version` - Schema version of the descriptor payload.
* `source` - Source configuration used to synchronize the descriptor content. [See below](#source-block).

### `agent_skills_definition` Block

* `additional_data` - Additional data associated with the agent skills definition. [See below](#agent_skills_definition-additional_data-block).
* `data` - Agent skills definition content, serialized as descriptor payload data.
* `data_schema_version` - Schema version of the descriptor payload.

### `agent_skills_definition` `additional_data` Block

* `skill_md` - Markdown skill content associated with the agent skills definition. [See below](#skill_md-block).

### `skill_md` Block

* `data` - Agent skills markdown content, serialized as descriptor payload data.
* `data_schema_version` - Schema version of the descriptor payload.
* `source` - Source configuration used to synchronize the markdown content. [See below](#source-block).

### `custom` Block

* `data` - Descriptor payload data.

### `mcp_server` Block

* `additional_data` - Additional data associated with the MCP server descriptor. [See below](#mcp_server-additional_data-block).
* `data` - MCP server descriptor content, serialized as descriptor payload data.
* `data_schema_version` - Schema version of the descriptor payload.
* `source` - Source configuration used to synchronize the descriptor content. [See below](#source-block).

### `mcp_server` `additional_data` Block

* `tools` - MCP tools descriptor containing tool definitions. [See below](#tools-block).

### `tools` Block

* `data` - Descriptor payload data.
* `data_schema_version` - Schema version of the descriptor payload.

### `source` Block

* `from_url` - URL-based descriptor source. [See below](#from_url-block).

### `from_url` Block

* `credential_provider_configuration` - Credential providers used to authenticate when fetching descriptor content from the source URL. [See below](#credential_provider_configuration-block).
* `url` - URL from which the descriptor content is retrieved.

### `credential_provider_configuration` Block

* `credential_provider` - Credential provider details corresponding to the credential provider type. [See below](#credential_provider-block).
* `credential_provider_type` - Type of credential provider. Valid values: `OAUTH`, `IAM`.

### `credential_provider` Block

* `iam_credential_provider` - IAM role credential provider details. [See below](#iam_credential_provider-block).
* `oauth_credential_provider` - OAuth 2.0 credential provider details. [See below](#oauth_credential_provider-block).

### `iam_credential_provider` Block

* `region` - AWS Region used for request signing.
* `role_arn` - ARN of the IAM role assumed for request signing.
* `service` - Service name used for request signing.

### `oauth_credential_provider` Block

* `custom_parameters` - Additional parameters included in the OAuth 2.0 token request.
* `grant_type` - OAuth 2.0 grant type used to obtain access tokens. Valid values: `CLIENT_CREDENTIALS`.
* `provider_arn` - ARN of the OAuth 2.0 credential provider resource in Amazon Bedrock AgentCore Identity.
* `scopes` - OAuth 2.0 scopes requested when obtaining access tokens.
