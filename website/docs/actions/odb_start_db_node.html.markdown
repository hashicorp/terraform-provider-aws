---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_start_db_node"
description: |-
  Starts an Oracle Database@AWS DB node.
---

# Action: aws_odb_start_db_node

Starts an existing Oracle Database@AWS DB node and waits for its status to become `AVAILABLE`. This action operates on one node in a cloud VM cluster and does not recreate the node or its cluster. A node that is already `AVAILABLE` requires no start request.

This action requires Terraform v1.14.0 or later and an AWS provider build containing this action. The first released provider version has not yet been assigned.

See the AWS [StartDbNode API](https://docs.aws.amazon.com/odb/latest/APIReference/API_StartDbNode.html) for service behavior. This action supports cloud VM clusters; Exascale VM clusters and Autonomous VM clusters are not supported by this action.

**Operation Behavior:**

The action checks the selected node before sending a start request. A `STOPPED` node is eligible to start; an `AVAILABLE` node succeeds without a start request. Other initial states produce an error. After starting the node, the action reports progress while waiting for `AVAILABLE`, and returns an error for a failed operation, an unexpected state, or a timeout.

The action does not manage persistent desired state and exports no attributes. Refresh the node data source to read its current status. Canceling or timing out Terraform does not cancel a request AWS has already accepted; inspect the node before invoking the action again.

**Permissions:**

The caller requires `odb:GetDbNode` and `odb:StartDbNode`. For resource-scoped permissions, authorize both the cloud VM cluster ARN and the DB node ARN. Discovery through `aws_odb_db_nodes` additionally requires `odb:ListDbNodes`. See the [Oracle Database@AWS service authorization reference](https://docs.aws.amazon.com/service-authorization/latest/reference/list_odb.html).

## Example Usage

### Basic Usage

Replace the example identifiers with the AWS identifiers of an existing cloud VM cluster and one of its DB nodes. Obtain node identifiers from the [`aws_odb_db_nodes` data source](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/odb_db_nodes). Use the node's `id`, rather than its ARN or Oracle Cloud identifier (OCID).

```terraform
action "aws_odb_start_db_node" "example" {
  config {
    cloud_vm_cluster_id = "example-cloud-vm-cluster"
    db_node_id          = "example-db-node"
  }
}
```

Preview and explicitly invoke the action:

```console
% terraform plan -invoke=action.aws_odb_start_db_node.example
% terraform apply -invoke=action.aws_odb_start_db_node.example
```

A plan or refresh does not start the node. Declaring an action alone does not invoke it on ordinary applies. An explicit `-invoke` command requests the operation each time it runs. For invocation from a resource lifecycle event, see the [reboot action example](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/actions/odb_reboot_db_node#repeatable-maintenance-request).

## Argument Reference

The following arguments are required:

* `cloud_vm_cluster_id` - (Required) AWS identifier of the cloud VM cluster containing the node. Must contain 6–64 letters, digits, underscores, tildes, periods, or hyphens.
* `db_node_id` - (Required) AWS identifier of the DB node to start. Must contain 6–64 letters, digits, underscores, tildes, periods, or hyphens.

The following arguments are optional:

* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `timeout` - (Optional) Maximum duration in seconds for the operation, including status checks. Must be between `1` and `86400`. Default: `3600`.
