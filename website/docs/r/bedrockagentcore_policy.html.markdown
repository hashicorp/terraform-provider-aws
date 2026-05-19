---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_policy"
description: |-
  Manages an AWS Bedrock AgentCore Policy.
---

# Resource: aws_bedrockagentcore_policy

Manages an AWS Bedrock AgentCore Policy. A Policy attaches Cedar authorization rules to a Policy Engine, which evaluates them at runtime to control agent access to resources.

## Example Usage

### Cedar Policy

```terraform
resource "aws_bedrockagentcore_policy" "example" {
  name             = "example_policy"
  policy_engine_id = "ExamplePolicyEngine-abc1234567"
  description      = "Allow read access to example resources"

  definition {
    cedar {
      statement = <<-EOT
        permit(principal, action == Action::"Read", resource);
      EOT
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `definition` - (Required) Policy definition. See [`definition` Block](#definition-block) for details.
* `name` - (Required) Name of the policy. Must be 1-48 characters and match the pattern `^[A-Za-z][A-Za-z0-9_]*$`. Changing this forces a new resource to be created.
* `policy_engine_id` - (Required) Identifier of the Policy Engine that owns this policy. Changing this forces a new resource to be created.

The following arguments are optional:

* `description` - (Optional) Description of the policy.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `validation_mode` - (Optional) Controls whether validation findings cause policy creation or update to fail. Valid values: `FAIL_ON_ANY_FINDINGS`, `IGNORE_ALL_FINDINGS`. Defaults to `FAIL_ON_ANY_FINDINGS`.

### `definition` Block

The `definition` configuration block supports the following arguments:

* `cedar` - (Required) Inline Cedar policy. See [`cedar` Block](#cedar-block) for details.

### `cedar` Block

The `cedar` configuration block supports the following arguments:

* `statement` - (Required) Cedar policy statement.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Timestamp when the Policy was created.
* `policy_arn` - ARN of the Policy.
* `policy_id` - Identifier of the Policy.
* `status` - Current lifecycle status of the Policy.
* `status_reasons` - Reasons accompanying the current status, if any.
* `updated_at` - Timestamp when the Policy was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Policy using the `policy_engine_id` and `policy_id` separated by a comma. For example:

```terraform
import {
  to = aws_bedrockagentcore_policy.example
  id = "ExamplePolicyEngine-abc1234567,Policy-0123456789"
}
```

Using `terraform import`, import Bedrock AgentCore Policy using the `policy_engine_id` and `policy_id` separated by a comma. For example:

```console
% terraform import aws_bedrockagentcore_policy.example ExamplePolicyEngine-abc1234567,Policy-0123456789
```
