---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_configuration_bundle"
description: |-
  Manages an AWS Bedrock AgentCore Configuration Bundle.
---

# Resource: aws_bedrockagentcore_configuration_bundle

Manages an AWS Bedrock AgentCore Configuration Bundle. A configuration bundle stores versioned component configurations for agent evaluation workflows. Each component is keyed by its identifier ARN (the component type is inferred from the ARN) and carries a free-form JSON configuration document. Updates are versioned git-style — each change to the components produces a new bundle version.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_configuration_bundle" "example" {
  bundle_name = "example_bundle"

  component {
    component_identifier = "arn:aws:bedrock-agentcore:::evaluator/Builtin.Helpfulness"
    configuration = jsonencode({
      threshold = 0.7
    })
  }
}
```

### Multiple Components

```terraform
resource "aws_bedrockagentcore_configuration_bundle" "example" {
  bundle_name = "example_bundle"
  description = "Evaluation components for the support agent"

  component {
    component_identifier = "arn:aws:bedrock-agentcore:::evaluator/Builtin.Helpfulness"
    configuration        = jsonencode({ threshold = 0.7 })
  }

  component {
    component_identifier = "arn:aws:bedrock-agentcore:::evaluator/Builtin.GoalSuccessRate"
    configuration        = jsonencode({ threshold = 0.5 })
  }
}
```

## Argument Reference

The following arguments are required:

* `bundle_name` - (Required) Name of the configuration bundle. Must be unique within your account. Up to 100 characters; must start with a letter and contain only alphanumeric characters and underscores.
* `component` - (Required) One or more [`component`](#component) blocks describing the versioned component configurations in the bundle. See [Component](#component) below.

The following arguments are optional:

* `branch_name` - (Optional) Branch name for the bundle version lineage. Between 1 and 128 characters.
* `commit_message` - (Optional) Commit message recorded for the bundle version. Required by the service when updating components; a default is supplied on update if omitted.
* `created_by` - (Optional) [`created_by`](#created_by) block identifying the source that created the version. See [created_by](#created_by) below.
* `description` - (Optional) Description of the configuration bundle. Between 1 and 500 characters.
* `kms_key_arn` - (Optional) ARN of the KMS key used to encrypt the bundle. Can be changed in place.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### component

* `component_identifier` - (Required) Identifier ARN of the component (for example, a built-in evaluator such as `arn:aws:bedrock-agentcore:::evaluator/Builtin.Helpfulness`). The component type is inferred from the ARN. Up to 2048 characters.
* `configuration` - (Required) Component configuration as a JSON document. Must be non-empty; its schema depends on the component type.

### created_by

* `name` - (Required) Name of the source that created the version.
* `arn` - (Optional) ARN of the source that created the version.

### timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `bundle_arn` - ARN of the configuration bundle.
* `bundle_id` - Identifier of the configuration bundle.
* `created_at` - Creation timestamp (RFC3339).
* `lineage_metadata` - Version lineage metadata. A single-element list with `branch_name`, `commit_message`, `parent_version_ids`, and a `created_by` block (`name`, `arn`).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `updated_at` - Last-updated timestamp (RFC3339).
* `version_id` - Identifier of the current bundle version. Advances each time the components are updated.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a configuration bundle using the `bundle_id`. For example:

```terraform
import {
  to = aws_bedrockagentcore_configuration_bundle.example
  id = "example_bundle-abc1234567"
}
```

Using `terraform import`, import a configuration bundle using the `bundle_id`. For example:

```console
% terraform import aws_bedrockagentcore_configuration_bundle.example example_bundle-abc1234567
```
