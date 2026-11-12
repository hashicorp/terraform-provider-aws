---
subcategory: "Agent Registry"
layout: "aws"
page_title: "AWS: aws_agentregistry_update_registry_record_status"
description: |-
  Approves, rejects, or deprecates an AWS Agent Registry registry record.
---

# Action: aws_agentregistry_update_registry_record_status

~> **Note:** `aws_agentregistry_update_registry_record_status` is in alpha. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

!> **Warning:** Setting a record's status to `DEPRECATED` is irreversible. A deprecated record cannot be resubmitted, approved, or have its content updated; any subsequent `terraform apply` that changes the record's content will fail. The only remaining operation is deletion.

Updates the status of an AWS Agent Registry registry record as part of the registry's curation workflow. Use it to approve or reject a record that is `PENDING_APPROVAL`, or to deprecate a record so that it is no longer discoverable.

This action performs a curation decision rather than enforcing a desired state. Terraform does not reconcile the record's status afterwards: if the record's content is later updated, it returns to `DRAFT` and must be resubmitted and approved again. If the record already has the requested status, the action succeeds without calling the API, preserving the existing status reason.

The AWS API enforces the record lifecycle, and requesting a transition that is not allowed from the record's current status fails with a validation error:

* `APPROVED` and `REJECTED` are only reachable from `PENDING_APPROVAL`. A `DRAFT` record must first be submitted using the [`aws_agentregistry_submit_registry_record_for_approval`](agentregistry_submit_registry_record_for_approval.html) action.
* `DEPRECATED` is reachable from `DRAFT`, `PENDING_APPROVAL`, and `APPROVED`, and is terminal.
* A `REJECTED` record returns to `DRAFT` when its content is updated, after which it can be resubmitted.

For information about AWS Agent Registry, see the [Agent Registry section of the Amazon Bedrock AgentCore Developer Guide](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry.html). For specific information about record status transitions, see the [UpdateRegistryRecordStatus](https://docs.aws.amazon.com/agent-registry-control/latest/APIReference/API_UpdateRegistryRecordStatus.html) page in the AWS Agent Registry Control API Reference.

## Example Usage

### Curator Approval

Curation is typically performed by a different principal than the one publishing records. Declare the decision as an action and invoke it explicitly once the record has been reviewed.

```terraform
data "aws_agentregistry_registry_record" "candidate" {
  registry_id = var.registry_id
  record_id   = var.record_id
}

action "aws_agentregistry_update_registry_record_status" "approve" {
  config {
    registry_id   = var.registry_id
    record_id     = var.record_id
    status        = "APPROVED"
    status_reason = "Security review completed"
  }
}
```

```console
% terraform apply -invoke=action.aws_agentregistry_update_registry_record_status.approve
```

### Rejection

```terraform
action "aws_agentregistry_update_registry_record_status" "reject" {
  config {
    registry_id   = var.registry_id
    record_id     = var.record_id
    status        = "REJECTED"
    status_reason = "Missing ownership metadata"
  }
}
```

### Deprecation

```terraform
action "aws_agentregistry_update_registry_record_status" "retire" {
  config {
    registry_id   = aws_agentregistry_registry.example.registry_id
    record_id     = aws_agentregistry_registry_record.example.record_id
    status        = "DEPRECATED"
    status_reason = "Replaced by example-record-v2"
  }
}
```

### Submit and Approve in a Single Apply

When the same principal both publishes and curates records, submission and approval can be chained on the record's lifecycle. Actions in an `action_trigger` run in order. Consider configuring `auto_approval_rules` on the registry instead, which achieves the same result with a single action.

```terraform
action "aws_agentregistry_submit_registry_record_for_approval" "submit" {
  config {
    registry_id = aws_agentregistry_registry_record.example.registry_id
    record_id   = aws_agentregistry_registry_record.example.record_id
  }
}

action "aws_agentregistry_update_registry_record_status" "approve" {
  config {
    registry_id   = aws_agentregistry_registry_record.example.registry_id
    record_id     = aws_agentregistry_registry_record.example.record_id
    status        = "APPROVED"
    status_reason = "Approved by Terraform"
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

  lifecycle {
    action_trigger {
      events = [after_create, after_update]
      actions = [
        action.aws_agentregistry_submit_registry_record_for_approval.submit,
        action.aws_agentregistry_update_registry_record_status.approve,
      ]
    }
  }
}
```

## Argument Reference

This action supports the following arguments:

* `record_id` - (Required) Identifier of the registry record. Accepts a record ID or ARN.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `registry_id` - (Required) Identifier of the registry containing the record. Accepts a registry ID or ARN.
* `status` - (Required) Target status for the record. Valid values: `APPROVED`, `REJECTED`, `DEPRECATED`.
* `status_reason` - (Required) Reason for the status change, for example why the record was approved, rejected, or deprecated. Maximum length of 255 characters.
* `timeout` - (Optional) Timeout in seconds to wait for a record that is still being created or updated to settle before changing its status. Default: `600`.
