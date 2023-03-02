---
subcategory: "AWS Systems Manager Incident Manager incidents"
layout: "aws"
page_title: "AWS: aws_ssmincidents_replication_set"
description: |-
Terraform resource for managing an incident replication set for AWS Systems Manager Incident Manager.
---

# Resource: aws_ssmincidents_replication_set

Provides a resource for managing a replication set in AWS Systems Manager Incident Manager.

## Example Usage

~> **NOTE:** When you delete a replication set, Incident Manager deletes all data associated with the replication set. This includes response plans, incident records, contacts, and escalation plans.

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

Add an AWS Region to a replication set. (You can add only one Region at a time.)

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

Delete an AWS Region from a replication set. (You can delete only one Region at a time.)

```terraform
resource "aws_ssmincidents_replication_set" "replicationSetName" {
  region {
    name = "us-west-2"
  }
}
```

## Using an AWS customer managed key with a replication set

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

~> **NOTE:** You must use one of the AWS Regions specified for the replication set in the specification for a Terraform provider. This is important when you perform complex update operations.

~> **NOTE:** After you create a replication set, you can only add or delete one Region at a time.

~> **NOTE:** Incident Manager doesn’t support updates to the customer managed key associated with a replication set. Instead, for a replication set with multiple Regions, you must first delete the Region from the replication set. Then, you add the Region back to the replication set with a different customer managed key in a separate `terraform apply` operation. For a replication set with only one Region, you must delete and recreate the entire replication set. To do this, comment out the replication set and all associated response plans. Then run the `terraform apply` command to recreate the replication set with the new customer managed key.

~> **NOTE:** You must either associate all Regions in a replication set with either an AWS owned key or a customer managed key. To change between an AWS owned key and a customer managed key, you must delete and recreate the replication set and all of its associated data.

~> **NOTE:** We recommend that you create the customer managed keys you need with the `terraform apply` command before you create the replication set. You can also create the keys and replication set at the same time with the same `terraform apply` command. Otherwise, to delete a replication set, you must run separate `terraform apply` commands to first delete the replication set and then the AWS KMS keys used by that replication set. If these recommendations are not followed, Terraform may accidentally delete the AWS KMS keys before the replication set is deleted, which will cause an error to be reported. In that case, you must manually restore the deleted key from the AWS Management Console before you can delete the replication set.

The `region` configuration block is required and supports the following arguments:

* `name` - (Required) The name of the Region, such as `ap-southeast-2`.
* `kms_key_arn` - (Optional) The Amazon Resource Name (ARN) for the customer managed key. If omitted, AWS manages the AWS KMS keys for you with an AWS owned key. This is indicated by a default value of `DefaultKey`.

The following arguments are optional:

* `tags` - Tags applied to the replication set.

For information about the maximum allowed number of Regions and tag value constraints, see [CreateReplicationSet](https://docs.aws.amazon.com/incident-manager/latest/APIReference/API_CreateReplicationSet.html) in the *AWS Systems Manager Incident Manager API Reference*.

## Attributes Reference

The following attributes are exported in addition to the previously mentioned arguments.

* `arn` - The ARN of the replication set.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `created_by` - The ARN of the user who created the replication set.
* `deletion_protected` - If `true`, the last region in a replication set cannot be deleted.
* `last_modified_by` - The ARN of the user who last modified the replication set.
* `status` - The overall status of a replication set.
    * Valid Values: `ACTIVE` | `CREATING` | `UPDATING` | `DELETING` | `FAILED`

The `region` configuration block also exports the following attributes for each Region:

* `status` - The current status of the replication set in a Region.
    * Valid Values: `ACTIVE` | `CREATING` | `UPDATING` | `DELETING` | `FAILED`
* `status_message` - More information about the status of a replication set.

## Timeouts

~> **NOTE:** When you create or delete replication sets with large numbers of response plans and data, the operation can take longer to complete. We recommend that you configure custom timeouts for larger replication sets.

~> **NOTE:** Each additional Region that you include when you create a replication set increases the amount of time required to complete the operation.

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts)

The default time for each of the following is 120m:

* `create`
* `update`
* `delete`

## Import

Use the following command to import an Incident Manager replication set:

```
$ terraform import aws_ssmincidents_replication_set.replicationSetName import
```
