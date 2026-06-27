---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_registry_record"
description: |-
  Manages an AWS Bedrock AgentCore Registry Record.
---

# Resource: aws_bedrockagentcore_registry_record

Manages an AWS Bedrock AgentCore Registry Record.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_registry_record" "example" {
  name        = "example-record"
  registry_id = aws_bedrockagentcore_registry.example.registry_id

  descriptor_type = "A2A"

  descriptors {
    a2a {
      agent_card {
        inline_content = <<EOF
{
  "name": "My Agent",
  "description": "Brief description of what this agent does",
  "url": "https://api.example.com/a2a",
  "version": "1.0.0",
  "protocolVersion": "0.3",
  "capabilities": {},
  "defaultInputModes": [
    "text/plain"
  ],
  "defaultOutputModes": [
    "text/plain"
  ],
  "skills": [
    {
      "id": "default-skill",
      "name": "Default Skill",
      "description": "Description of what this skill does",
      "tags": [
        "general"
      ]
    }
  ]
}
EOF
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `descriptor_type` - (Required) Descriptor type of the registry record. Valid values: `MCP`, `A2A`, `CUSTOM`, `AGENT_SKILLS`.
* `descriptors` - (Required) Descriptor-type-specific configuration containing the resource schema and metadata. See [`descriptors` Block](#descriptors-block) below.
* `name` - (Required) Name of the registry record.
* `registry_id` - (Required, Forces new resource) ID of the registry in which to create the record.

The following arguments are optional:

* `description` - (Optional) Description of the registry record.
* `record_version` - (Optional) Version of the registry record. Use this to track different versions of the record's content.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `descriptors` Block

The `descriptors` block supports the following. Exactly one of `a2a`, `agent_skills`, `custom`, or `mcp` must be configured.

* `a2a` - (Optional) Agent-to-Agent (A2A) protocol descriptor. See [`a2a` Block](#a2a-block) below.
* `agent_skills` - (Optional) Agent skills descriptor. See [`agent_skills` Block](#agent_skills-block) below.
* `custom` - (Optional) Custom descriptor for resources that don't conform to a standard protocol, such as APIs, Lambda functions, or servers. See [`custom` Block](#custom-block) below.
* `mcp` - (Optional) Model Context Protocol (MCP) descriptor. See [`mcp` Block](#mcp-block) below.

### `a2a` Block

The `a2a` block supports the following:

* `agent_card` - (Required) Agent card definition for the A2A descriptor. See [`agent_card` Block](#agent_card-block) below.

### `agent_card` Block

The `agent_card` block supports the following:

* `inline_content` - (Required) JSON string containing the agent card definition.
* `schema_version` - (Optional) Schema version of the agent card definition.

### `agent_skills` Block

The `agent_skills` block supports the following. Exactly one of `skill_definition` or `skill_md` must be configured.

* `skill_definition` - (Optional) Structured agent skill definition. See [`skill_definition` Block](#skill_definition-block) below.
* `skill_md` - (Optional) Markdown-based agent skill definition. See [`skill_md` Block](#skill_md-block) below.

### `skill_definition` Block

The `skill_definition` block supports the following:

* `inline_content` - (Required) JSON string containing the agent skill definition.
* `schema_version` - (Optional) Schema version of the agent skill definition.

### `skill_md` Block

The `skill_md` block supports the following:

* `inline_content` - (Required) Markdown content containing the agent skill definition.

### `custom` Block

The `custom` block supports the following:

* `inline_content` - (Required) JSON string containing the custom resource schema and metadata.

### `mcp` Block

The `mcp` block supports the following:

* `server` - (Required) MCP server definition. See [`server` Block](#server-block) below.
* `tools` - (Optional) MCP tools definition. See [`tools` Block](#tools-block) below.

### `server` Block

The `server` block supports the following:

* `inline_content` - (Required) JSON string containing the MCP server definition.
* `schema_version` - (Optional) Schema version of the MCP server definition.

### `tools` Block

The `tools` block supports the following:

* `inline_content` - (Required) JSON string containing the MCP tools definition.
* `schema_version` - (Optional) Schema version of the MCP tools definition.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `record_arn` - ARN of the registry record.
* `record_id` - Unique ID of the registry record.
* `status` - Status of the registry record. Valid values: `DRAFT`, `PENDING_APPROVAL`, `APPROVED`, `REJECTED`, `DEPRECATED`, `CREATING`, `UPDATING`, `CREATE_FAILED`, `UPDATE_FAILED`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_registry_record.example
  identity = {
    registry_id = "Fx0UXvOfe4Y7iHsI"
    record_id   = "53ctXuJJIC2u"
  }
}

resource "aws_bedrockagentcore_registry_record" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `record_id` (String) Registry record ID.
* `registry_id` (String) Registry ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Registry Records using `registry_id` and `record_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_bedrockagentcore_registry_record.example
  id = "Fx0UXvOfe4Y7iHsI,53ctXuJJIC2u"
}
```

Using `terraform import`, import Registry Records using `registry_id` and `record_id` separated by a comma (`,`). For example:

```console
% terraform import aws_bedrockagentcore_registry_record.example Fx0UXvOfe4Y7iHsI,53ctXuJJIC2u
```
