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

* `name` - (Required) Name of the registry record.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `record_arn` - ARN of the registry record.
* `record_id` - Unique ID of the registry record.
* `status` - Status of the registry record.

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

- `record_id` (String) Registry record ID.
- `registry_id` (String) Registry ID.

#### Optional

* `account_id` (String) AWS account ID for this resource.
* `region` (String) AWS Region for this resource.

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
