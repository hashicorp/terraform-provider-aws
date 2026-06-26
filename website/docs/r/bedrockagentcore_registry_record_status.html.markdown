---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_registry_record_status"
description: |-
  Manages the status of an AWS Bedrock AgentCore Registry Record.
---

# Resource: aws_bedrockagentcore_registry_record_status

Manages the status of an AWS Bedrock AgentCore Registry Record. Use this resource to approve, reject, or deprecate a registry record.

~> Deletion of this resource will not modify any settings, only remove the resource from state.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_registry_record_status" "example" {
  registry_id = aws_bedrockagentcore_registry_record.example.registry_id
  record_id   = aws_bedrockagentcore_registry_record.example.record_id

  status        = "APPROVED"
  status_reason = "LGTM"

  # Ensure that the registry record is in PENDING_APPROVAL state.
  lifecycle {
    action_trigger {
      events  = [before_create]
      actions = [action.aws_bedrockagentcore_submit_registry_record_for_approval.example]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `record_id` - (Required, Forces new resource) Registry record ID.
* `registry_id` - (Required, Forces new resource) Registry ID.
* `status` - (Required) Target status for the registry record. Valid values: `APPROVED`, `REJECTED`, `DEPRECATED`.
* `status_reason` - (Required) Reason for any status change, such as why the record was approved or rejected.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_registry_record_status.example
  identity = {
    registry_id = "Fx0UXvOfe4Y7iHsI"
    record_id   = "53ctXuJJIC2u"
  }
}

resource "aws_bedrockagentcore_registry_record_status" "example" {
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

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Registry Record statuses using `registry_id` and `record_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_bedrockagentcore_registry_record_status.example
  id = "Fx0UXvOfe4Y7iHsI,53ctXuJJIC2u"
}
```

Using `terraform import`, import Registry Record statuses using `registry_id` and `record_id` separated by a comma (`,`). For example:

```console
% terraform import aws_bedrockagentcore_registry_record_status.example Fx0UXvOfe4Y7iHsI,53ctXuJJIC2u
```
