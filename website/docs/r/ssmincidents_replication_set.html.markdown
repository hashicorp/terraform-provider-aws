---
subcategory: "SSM Incident Manager Incidents"
layout: "aws"
page_title: "AWS: aws_ssmincidents_replication_set"
description: |-
  Terraform resource for managing an AWS SSM Incident Manager Incidents Replication Set.
---

# Resource: aws_ssmincidents_replication_set

Terraform resource for managing an AWS SSM Incident Manager Incidents Replication Set.

~> **NOTE:** Deleting a Replication Set will delete all associated Response Plans.

## Example Usage

### Basic Usage

Creating a new Replication Set:

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

Adding a region (make sure only one new region is added at once):

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

Deleting a region (make sure only one region is deleted at once):

```terraform
resource "aws_ssmincidents_replication_set" "replicationSetName" {
  region {
    name = "us-west-2"
  }
}
```

## Basic Usage with Customer Managed Keys

Creating a new Replication Set:

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

~> **NOTE:** The region specified by the provider must be one of the regions within the Replication Set at all times. Make sure to keep this in mind when performing complex update operations.

~> **NOTE:** After creation, regions can only be added or deleted one at a time.

~> **NOTE:** Incident Manager does not support updating Customer Managed Keys. To do this, you must first delete that region then recreate it with an updated Customer Managed Key in separate terraform apply operations. Performing this when the Replication Set contains only one region requires deleting the Replication Set (which will delete all associated Response Plans). This can be done by commenting out the Replication Set and all Response Plans, running Terraform Apply and recreating the Replication Set and Response Plans with the new correct KMS keys.

~> **NOTE:** Either all regions must have a Customer Managed KMS key or None. Changing from using Amazon owned keys to customer managed keys or vice versa requires the Replication Set and all associated data to be destroyed and recreated.

~> **NOTE:** If possible, create all KMS keys used by a Replication Set in the same or previous `terraform apply` command compared to when the Replication Set is created. If this is not possible and you want to delete a Replication Set, make sure to delete the Replication Set first in its own `terraform apply` command before deleting the KMS keys used by that Replication Set in a separate `Terraform apply` command. This minimises the likelihood that a KMS Key which is used by a Replication Set is accidentally deleted before attempting to delete the Replication Set, which will result in an error. If this error occurs, the deleted KMS Key must be manually re-enabled via the AWS Console before the Replication Set can be successfully deleted.

The `region` configuration block is required and supports the following arguments:

* `name` - (Required) name of region.
* `kms_key_arn` - (Optional) ARN of KMS Encryption Key. If this is not provided, AWS will manage your KMS keys for you and this will be denoted with a default value of `DefaultKey`.

The following arguments are optional:

* `tags` - (Optional) Tags associated with the Replication Set.
  
More details for maximum number of regions and tag value constraints can be found [here](https://docs.aws.amazon.com/incident-manager/latest/APIReference/API_CreateReplicationSet.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Replication Set.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `created_by` - Who created the Replication Set.
* `created_time` - When the Replication Set was created.
* `deletion_protected` - If enabled, the last region in a Replication Set cannot be deleted.
* `last_modified_by` - Who last modified the Replication Set.
* `last_modified_time` - When the Replication Set was last modified
* `status` - Overall status of a replication Set. The status will be one of: `ACTIVE`, `CREATING`, `UPDATING`, `DELETING` or `FAILED`.

In addition to the arguments above, The `region` configuration block exports the following attributes for each region:

* `status` - Region Status.
* `status_update_time` - Last time Region Status was updated.
* `status_message` - More information about the status of a region.

## Timeouts

~> **NOTE:** Update and Delete operations on Replication Sets with large amounts of response plans and data take longer to complete. It may be required for custom timeouts to be configured in these circumstances.

~> **NOTE:** Creating a new Replication Set will take additional time for each extra region in the created Replication Set beyond the first.

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `120m`)
* `update` - (Default `120m`)
* `delete` - (Default `120m`)

## Import

SSM Incident Manager Incidents Replication Set can be imported using this command:

```
$ terraform import aws_ssmincidents_replication_set.replicationSetName import
```
