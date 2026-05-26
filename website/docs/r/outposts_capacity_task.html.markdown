---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_capacity_task"
description: |-
  Terraform resource for managing an AWS Outposts Capacity Task.
---

# Resource: aws_outposts_capacity_task

Terraform resource for managing an AWS Outposts Capacity Task.

A capacity task redistributes the instance pools available on an Outpost rack or server to match the `instance_pool` configuration declared in the resource. Starting a capacity task is a long-running, asynchronous operation — Terraform waits for it to reach a terminal state (`COMPLETED`, `CANCELLED`, or `FAILED`) before finishing the apply.

## Example Usage

### Minimal

```terraform
data "aws_outposts_outposts" "example" {}

resource "aws_outposts_capacity_task" "example" {
  outpost_identifier = tolist(data.aws_outposts_outposts.example.arns)[0]

  instance_pool {
    instance_type = "m5.large"
    count         = 2
  }
}
```

### Multiple instance pools, excluded instances, and a specified blocking-instance action

```terraform
data "aws_outposts_assets" "example" {
  arn = "arn:aws:outposts:us-west-2:123456789012:outpost/op-1234567890abcdef"
}

resource "aws_outposts_capacity_task" "example" {
  outpost_identifier                = "op-1234567890abcdef"
  task_action_on_blocking_instances = "WAIT_FOR_EVACUATION"
  asset_id                          = data.aws_outposts_assets.example.asset_ids[0]

  instance_pool {
    instance_type = "m5.large"
    count         = 4
  }

  instance_pool {
    instance_type = "c5.xlarge"
    count         = 2
  }

  # Instance IDs the capacity task must not stop when re-balancing capacity.
  instances_to_exclude {
    instances = ["i-0123456789abcdef0", "i-0fedcba9876543210"]
  }

  timeouts {
    create = "90m"
    delete = "15m"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `outpost_identifier` - (Required) ID or ARN of the Outpost on which to run the capacity task. Both forms are accepted; the provider normalizes the value internally. Changing this value forces a new resource.
* `instance_pool` - (Required) One or more `instance_pool` blocks defining the desired instance-type layout for the Outpost. See [below](#instance_pool). At least one block is required. Changing any value forces a new resource.
* `asset_id` - (Optional) ID of a specific Outposts asset (hardware server) to target for the capacity task. If omitted, AWS selects an appropriate asset automatically. Discover valid asset IDs with the [`aws_outposts_assets`](../d/outposts_assets.html.markdown) data source. Changing this value forces a new resource.
* `instances_to_exclude` - (Optional) Single `instances_to_exclude` block specifying user-owned running instances that must not be stopped to free up capacity. See [below](#instances_to_exclude). Note: AWS does not return this value via the Get/Describe API; after import, you must add the block back to your configuration manually — see [Import](#import).
* `order_id` - (Optional) ID of the Amazon Web Services Outposts order associated with the capacity task. Changing this value forces a new resource.
* `task_action_on_blocking_instances` - (Optional) Action to take if running instances block the capacity task. Valid values are `WAIT_FOR_EVACUATION` and `FAIL_TASK`. Changing this value forces a new resource.
* `timeouts` - (Optional) Configuration block with timeouts. See [below](#timeouts).

### instance_pool

* `instance_type` - (Required) Instance type for this pool entry. Must be an instance type supported by the target Outpost. Changing this value forces a new resource.
* `count` - (Required) Number of instances of `instance_type` that should be present after the task completes. Must be at least `1`. Changing this value forces a new resource.

### instances_to_exclude

* `instances` - (Required) Set of EC2 instance IDs (of user-owned instances running on the Outpost) that the capacity task must not stop. At least one instance ID is required.

### timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `delete` - (Default `10m`)

~> **Long-running capacity tasks.** The default `create` timeout of `60m` is sufficient for most re-balancing operations on small to medium instance types. However, capacity tasks that change the configuration of bare-metal instance types (`*.metal`) or very large instance types (`24xlarge`, `48xlarge`, etc.) in the current or target state can take **8 to 12 hours** to complete, because AWS must stop, reconfigure, and re-start the underlying hardware. If your `instance_pool` configuration or the current state of the Outpost involves one of these instance types, override the `create` timeout accordingly — for example:

```terraform
timeouts {
  create = "12h"
  delete = "4h"
}
```

If the `create` waiter times out before the capacity task reaches a terminal state, Terraform will return an error and will **not** write the resource to state, even though the task continues to run in AWS. Recovering from this condition requires locating the orphaned task (for example with `aws outposts list-capacity-tasks --outpost-identifier <outpost-id>`), waiting for it to complete, and then importing it using `terraform import aws_outposts_capacity_task.example <outpost-id>/<capacity-task-id>`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `capacity_task_id` - ID assigned by AWS to the capacity task (for example, `cap-1a2b3c4d5e6f7g8h9`).
* `status` - Current status of the capacity task. One of `REQUESTED`, `IN_PROGRESS`, `WAITING_FOR_EVACUATION`, `CANCELLATION_IN_PROGRESS`, `COMPLETED`, `CANCELLED`, or `FAILED`. See the [AWS documentation](https://docs.aws.amazon.com/outposts/latest/APIReference/API_GetCapacityTask.html) for semantics.
* `creation_date` - RFC 3339 timestamp at which the capacity task was created.
* `completion_date` - RFC 3339 timestamp at which the capacity task reached a terminal state (if any).
* `failure_reason` - Human-readable reason reported by AWS when the capacity task failed. `null` unless the terminal state is `FAILED`.

## Lifecycle

Because every argument of this resource is marked as forces-new, any change to the configuration results in destroying and re-creating the capacity task. Tasks that are already in a terminal state (`COMPLETED` or `CANCELLED`) are left in place on destroy and only removed from Terraform state; tasks still in flight are cancelled and Terraform waits for them to reach `CANCELLED`. If a task reaches the terminal state `FAILED` during `delete`, the provider tolerates the "already in a terminal state" error returned by `CancelCapacityTask` and considers the resource successfully destroyed.

If a create operation produces a `FAILED` task, the resource is not written to Terraform state (the `failure_reason` is surfaced in the diagnostic instead), so no follow-up destroy is required.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_outposts_capacity_task.example
  identity = {
    outpost_identifier = "op-1234567890abcdef"
    capacity_task_id   = "cap-1a2b3c4d5e6f7g8h9"
  }
}

resource "aws_outposts_capacity_task" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `outpost_identifier` (String) Outpost identifier supplied when the task was created (ID or ARN).
* `capacity_task_id` (String) AWS-assigned capacity task ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Capacity Task using the `outpost_identifier` and `capacity_task_id` joined by a forward slash (`/`). For example:

```terraform
import {
  to = aws_outposts_capacity_task.example
  id = "op-1234567890abcdef/cap-1a2b3c4d5e6f7g8h9"
}
```

Using `terraform import`, import a Capacity Task using the same composite ID. For example:

```console
% terraform import aws_outposts_capacity_task.example op-1234567890abcdef/cap-1a2b3c4d5e6f7g8h9
```
