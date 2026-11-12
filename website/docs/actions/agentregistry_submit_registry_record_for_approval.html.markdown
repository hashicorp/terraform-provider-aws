---
subcategory: "Agent Registry"
layout: "aws"
page_title: "AWS: aws_agentregistry_submit_registry_record_for_approval"
description: |-
  Submits an AWS Agent Registry registry record for approval.
---

# Action: aws_agentregistry_submit_registry_record_for_approval

~> **Note:** `aws_agentregistry_submit_registry_record_for_approval` is in alpha. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Submits an AWS Agent Registry registry record for approval, moving it into the registry's approval workflow. Depending on the registry's `approval_configuration`, the record is either approved immediately or set to `PENDING_APPROVAL` for a curator to approve or reject.

Registry records are created in `DRAFT` status and are not discoverable until approved. Creating a record with the `aws_agentregistry_registry_record` resource does not submit it, and auto-approval rules on the registry do not bypass submission: a record in a registry configured with `APPROVE_ALL` still remains in `DRAFT` until it is submitted. Updating a record's content also returns it to `DRAFT`, so records must be resubmitted after every change.

The action's behavior depends on the record's current status:

* `DRAFT` - The record is submitted. The resulting status is `APPROVED` if the registry auto-approves it, otherwise `PENDING_APPROVAL`.
* `PENDING_APPROVAL`, `APPROVED` - No change is made; the action succeeds.
* `CREATING`, `UPDATING` - The action waits for the record to settle, then proceeds based on the resulting status.
* `REJECTED` - The action fails. A rejected record cannot be resubmitted directly; update its content to return it to `DRAFT`, then submit it again.
* `DEPRECATED` - The action fails. Deprecation is a terminal state; the record cannot be submitted or modified.

For information about AWS Agent Registry, see the [Agent Registry section of the Amazon Bedrock AgentCore Developer Guide](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/registry.html). For specific information about submitting records, see the [SubmitRegistryRecordForApproval](https://docs.aws.amazon.com/agent-registry-control/latest/APIReference/API_SubmitRegistryRecordForApproval.html) page in the AWS Agent Registry Control API Reference.

## Example Usage

### Auto-Approved Registry

Records in a registry with `auto_approval_rules = ["APPROVE_ALL"]` become discoverable on every apply. Triggering on both `after_create` and `after_update` ensures a record is resubmitted whenever Terraform changes its content.

~> **Note:** In HCP Terraform Stacks using Terraform 1.16 and later, the `caller` symbol can reference the triggering resource's attributes (for example `caller.record_id`), allowing one action block to serve many records. It is not available in the Terraform CLI, so the examples below reference the record resource directly.

```terraform
resource "aws_agentregistry_registry" "example" {
  name = "example-registry"

  approval_configuration {
    auto_approval_rules = ["APPROVE_ALL"]
  }

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}

action "aws_agentregistry_submit_registry_record_for_approval" "example" {
  config {
    registry_id = aws_agentregistry_registry_record.example.registry_id
    record_id   = aws_agentregistry_registry_record.example.record_id
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
      events  = [after_create, after_update]
      actions = [action.aws_agentregistry_submit_registry_record_for_approval.example]
    }
  }
}
```

### Manual Approval Workflow

In a registry without auto-approval rules, submission moves the record to `PENDING_APPROVAL`. A curator then approves or rejects it, typically from a separate configuration using the [`aws_agentregistry_update_registry_record_status`](agentregistry_update_registry_record_status.html) action.

```terraform
resource "aws_agentregistry_registry" "example" {
  name = "example-registry"

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}

action "aws_agentregistry_submit_registry_record_for_approval" "example" {
  config {
    registry_id = aws_agentregistry_registry_record.example.registry_id
    record_id   = aws_agentregistry_registry_record.example.record_id
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
      events  = [after_create, after_update]
      actions = [action.aws_agentregistry_submit_registry_record_for_approval.example]
    }
  }
}
```

### Invoke From the CLI

```terraform
action "aws_agentregistry_submit_registry_record_for_approval" "example" {
  config {
    registry_id = "registry-id-12345678"
    record_id   = "record-id-1234"
  }
}
```

```console
% terraform apply -invoke=action.aws_agentregistry_submit_registry_record_for_approval.example
```

## Argument Reference

This action supports the following arguments:

* `record_id` - (Required) Identifier of the registry record to submit. Accepts a record ID or ARN.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `registry_id` - (Required) Identifier of the registry containing the record. Accepts a registry ID or ARN.
* `timeout` - (Optional) Timeout in seconds to wait for a record that is still being created or updated to settle before submitting it. Default: `600`.
