---
subcategory: "SSM Incident Manager Incidents"
layout: "aws"
page_title: "AWS: aws_ssmincidents_replication_set"
description: |-
  Terraform resource for managing an incident replication set for AWS Systems Manager Incident Manager.
---

# Resource: aws_ssmincidents_replication_set

Provides a resource for managing a replication set in AWS Systems Manager Incident Manager.

~> **NOTE:** Deleting a replication set also deletes all Incident Manager related data including response plans, incident records, contacts and escalation plans.

## Example Usage

### Basic Usage

Create a replication set.

```terraform
resource "aws_ssmincidents_replication_set" "replicationSetName" {
  region {
    name = "us-west-2"
  }

  tags = {
    exampleTag = "exampleValue"
  }
}
```

Add a Region to a replication set. (You can add only one Region at a time.)

```terraform
resource "aws_ssmincidents_replication_set" "replicationSetName" {
  region {
    name = "us-west-2"
  }

  region {
    name = "ap-southeast-2"
  }
}
```

Delete a Region from a replication set. (You can delete only one Region at a time.)

```terraform
resource "aws_ssmincidents_replication_set" "replicationSetName" {
  region {
    name = "us-west-2"
  }
}
```

## Basic Usage with an AWS Customer Managed Key

Create a replication set with an AWS Key Management Service (AWS KMS) customer manager key:

```terraform

resource "aws_kms_key" "example_key" {}

resource "aws_ssmincidents_replication_set" "replicationSetName" {
  region {
    name        = "us-west-2"
    kms_key_arn = aws_kms_key.example_key.arn
  }

  tags = {
    exampleTag = "exampleValue"
  }
}
```

## Argument Reference

~> **NOTE:** The Region specified by a Terraform provider must always be one of the Regions specified for the replication set. This is especially important when you perform complex update operations.

~> **NOTE:** After a replication set is created, you can add or delete only one Region at a time.

~> **NOTE:** Incident Manager does not support updating the customer managed key associated with a replication set. Instead, for a replication set with multiple Regions, you must first delete a Region from the replication set, then re-add it with a different customer managed key in separate `terraform apply` operations. For a replication set with only one Region, the entire replication set must be deleted and recreated. To do this, comment out the replication set and all response plans, and then run the `terraform apply` command to recreate the replication set with the new customer managed key.

~> **NOTE:** You must either use AWS-owned keys on all regions of a replication set, or customer managed keys. To change between an AWS owned key and a customer managed key, a replication set and it associated data must be deleted and recreated.

~> **NOTE:** If possible, create all the customer managed keys you need (using the `terraform apply` command) before you create the replication set, or create the keys and replication set in the same `terraform apply` command. Otherwise, to delete a replication set, you must run one `terraform apply` command to delete the replication set and another to delete the AWS KMS keys used by the replication set. Deleting the AWS KMS keys before deleting the replication set results in an error. In that case, you must manually reenable the deleted key using the AWS Management Console before you can delete the replication set.

The `region` configuration block is required and supports the following arguments:

* `name` - (Required) The name of the Region, such as `ap-southeast-2`.
* `kms_key_arn` - (Optional) The Amazon Resource name (ARN) of the customer managed key. If omitted, AWS manages the AWS KMS keys for you, using an AWS owned key, as indicated by a default value of `DefaultKey`.

The following arguments are optional:

* `tags` - Tags applied to the replication set.

For information about the maximum allowed number of Regions and tag value constraints, see [CreateReplicationSet in the *AWS Systems Manager Incident Manager API Reference*](https://docs.aws.amazon.com/incident-manager/latest/APIReference/API_CreateReplicationSet.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the replication set.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `created_by` - The ARN of the user who created the replication set.
* `created_time` - A timestamp showing when the replication set was created.
* `deletion_protected` - If `true`, the last region in a replication set cannot be deleted.
* `last_modified_by` - A timestamp showing when the replication set was last modified.
* `last_modified_time` - When the replication set was last modified
* `status` - The overall status of a replication set.
    * Valid Values: `ACTIVE` | `CREATING` | `UPDATING` | `DELETING` | `FAILED`

In addition to the preceding arguments, the `region` configuration block exports the following attributes for each Region:

* `status` - The current status of the Region.
    * Valid Values: `ACTIVE` | `CREATING` | `UPDATING` | `DELETING` | `FAILED`
* `status_update_time` - A timestamp showing when the Region status was last updated.
* `status_message` - More information about the status of a Region.

## Timeouts

~> **NOTE:** `Update` and `Delete` operations applied to replication sets with large numbers of response plans and data take longer to complete. We recommend that you configure custom timeouts for this situation.

~> **NOTE:** Each additional Region included when you create a replication set increases the amount of time required to complete the `create` operation.

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `120m`)
* `update` - (Default `120m`)
* `delete` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Incident Manager replication. For example:

```terraform
import {
  to = aws_ssmincidents_replication_set.replicationSetName
  id = "import"
}
```

Using `terraform import`, import an Incident Manager replication. For example:

```console
% terraform import aws_ssmincidents_replication_set.replicationSetName import
```
