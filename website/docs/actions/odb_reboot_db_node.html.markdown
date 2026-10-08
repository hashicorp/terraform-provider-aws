---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_reboot_db_node"
description: |-
  Reboots an Oracle Database@AWS DB node.
---

# Action: aws_odb_reboot_db_node

Reboots an existing Oracle Database@AWS DB node and waits for its status to return to `AVAILABLE`. This action operates on one node in a cloud VM cluster and does not recreate the node or its cluster.

This action requires Terraform v1.14.0 or later and an AWS provider build containing this action. The first released provider version has not yet been assigned.

See the AWS [RebootDbNode API](https://docs.aws.amazon.com/odb/latest/APIReference/API_RebootDbNode.html) for service behavior. This action supports cloud VM clusters; Exascale VM clusters and Autonomous VM clusters are not supported by this action.

~> **Note:** Rebooting a DB node interrupts workloads on that node. Coordinate the operation with database maintenance and availability requirements.

**Operation Behavior:**

The selected node must initially be `AVAILABLE`. The action submits a reboot request and reports progress while observing the node's lifecycle status. To avoid treating its pre-reboot status as completion, the action requires an observed transition followed by two consecutive `AVAILABLE` status checks. A failed operation, unexpected state, or timeout produces an error.

AWS does not return a reboot operation identifier or a last-reboot timestamp. If no transition is observed, the action times out without claiming success, even if AWS completed the reboot between observations. Inspect the node before requesting another reboot after any failure, timeout, or cancellation. AWS may continue a request it has already accepted; canceling Terraform does not cancel the reboot.

The action does not manage persistent desired state and exports no attributes. Refresh the node data source to read its current status.

**Permissions:**

The caller requires `odb:GetDbNode` and `odb:RebootDbNode`. For resource-scoped permissions, authorize both the cloud VM cluster ARN and the DB node ARN. Discovery through `aws_odb_db_nodes` additionally requires `odb:ListDbNodes`. See the [Oracle Database@AWS service authorization reference](https://docs.aws.amazon.com/service-authorization/latest/reference/list_odb.html).

## Example Usage

### Basic Usage

Replace the example identifiers with the AWS identifiers of an existing cloud VM cluster and one of its DB nodes. Obtain node identifiers from the [`aws_odb_db_nodes` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/odb_db_nodes). Use the node's `id`, rather than its ARN or Oracle Cloud identifier (OCID).

```terraform
action "aws_odb_reboot_db_node" "example" {
  config {
    cloud_vm_cluster_id = "example-cloud-vm-cluster"
    db_node_id          = "example-db-node"
  }
}
```

Preview and explicitly invoke the action:

```console
% terraform plan -invoke=action.aws_odb_reboot_db_node.example
% terraform apply -invoke=action.aws_odb_reboot_db_node.example
```

Every explicit `-invoke` requests another reboot. A plan or refresh does not reboot the node, and declaring an action alone does not invoke it on ordinary applies.

### Repeatable Maintenance Request

Add this resource alongside the action above to request a reboot when the resource is first created and whenever its maintenance revision changes:

```terraform
resource "terraform_data" "example" {
  input = "maintenance-001"

  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_odb_reboot_db_node.example]
    }
  }
}
```

Change `input` to `"maintenance-002"` and apply to explicitly request the next reboot. An unchanged apply after a successful operation does not update this resource and does not repeat the reboot. Use a stable revision value; a value that changes on every plan would request repeated operations. Replacing or recreating the trigger resource also requests a reboot.

## Argument Reference

The following arguments are required:

* `cloud_vm_cluster_id` - (Required) AWS identifier of the cloud VM cluster containing the node. Must contain 6–64 letters, digits, underscores, tildes, periods, or hyphens.
* `db_node_id` - (Required) AWS identifier of the DB node to reboot. Must contain 6–64 letters, digits, underscores, tildes, periods, or hyphens.

The following arguments are optional:

* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `timeout` - (Optional) Maximum duration in seconds for the operation, including status checks. Must be between `1` and `86400`. Default: `3600`.
