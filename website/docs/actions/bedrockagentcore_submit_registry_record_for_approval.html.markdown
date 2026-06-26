---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_submit_registry_record_for_approval"
description: |-
  Submits an AWS Bedrock AgentCore Registry Record for approval.
---

# Action: aws_bedrockagentcore_submit_registry_record_for_approval

Submits an AWS Bedrock AgentCore Registry Record for approval. This transitions the record from `DRAFT` status to `PENDING_APPROVAL` status. If the registry has auto-approval enabled, the record is automatically approved.

## Example Usage

### Basic Usage

```terraform
action "aws_bedrockagentcore_submit_registry_record_for_approval" "example" {
  config {
    registry_id = aws_bedrockagentcore_registry_record.example.registry_id
    record_id   = aws_bedrockagentcore_registry_record.example.record_id
  }
}

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

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.aws_bedrockagentcore_submit_registry_record_for_approval.example]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `record_id` - (Required) Registry record ID.
* `registry_id` - (Required) Registry ID.

The following arguments are optional:

* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).