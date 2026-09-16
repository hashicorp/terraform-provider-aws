---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_agent_versions"
description: |-
  Terraform data source for managing an AWS Amazon BedrockAgent Agent Versions.
---

# Data Source: aws_bedrockagent_agent_versions

Terraform data source for managing an AWS Amazon BedrockAgent Agent Versions.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrockagent_agent_versions" "test" {
  agent_id = aws_bedrockagent_agent.test.agent_id
}
```

## Argument Reference

This data source supports the following arguments:

* `agent_id` - (Required) Unique identifier of the agent.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `agent_version_summaries` - List of objects, each of which contains information about a version of the agent. See [`agent_version_summaries` Block](#agent_version_summaries-block)

### `agent_version_summaries` Block

* `agent_name` - Name of agent to which the version belongs.
* `agent_status` - Status of the agent to which the version belongs.
* `agent_version` - Version of the agent.
* `created_at` - Time at which the version was created.
* `description` - Description of the version of the agent.
* `guardrail_configuration` - Details aout the guardrail associated with the agent. See [`guardrail_configuration` Block](#guardrail_configuration-block)
* `updated_at` - Time at which the version was last updated.

### `guardrail_configuration` Block

* `guardrail_identifier` - Unique identifier of the guardrail.
* `guardrail_version` - Version of the guardrail.
