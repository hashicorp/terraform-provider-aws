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
resource "aws_ssmincidents_replication_set" "test" {
    regions = {
        us-west-2 = ""
        ap-south-1 = ""

    }

    tags = {
        exampleTag = "exampleValue"
    }
}
```

Adding a region (make sure only one new region is added):

```terraform
resource "aws_ssmincidents_replication_set" "test" {
    regions = {
        us-west-2 = ""
        ap-south-1 = ""
        ap-southeast-2 = ""
    }
}
```

Deleting a region (make sure only one region is deleted):

```terraform
resource "aws_ssmincidents_replication_set" "test" {
    regions = {
        us-west-2 = ""
        ap-southeast-2 = ""
    }
}
```

## Argument Reference

The following arguments are required:

~> **NOTE:** The region specified by the provider must be one of the regions within the Replication Set at all times. Make sure to keep this in mind when performing complex update operations,

~> **NOTE:** After creation, regions can only be added or deleted one at a time. Examples above demonstrate how to account for this behaviour.

~> **NOTE:** Changing the Customer Managed Key for a single region requires us to first delete that region then recreate it with an updated Customer Managed Key in separate terraform apply operations. Performing this when the Replication Set contains only one region requires deleting the Replication Set. This can be done by commenting out the Replication Set, running Terraform Apply and recreating the Replication Set with the new correct KMS keys.

~> **NOTE:** Either all regions must have a Customer Managed Key or None. Changing from using the default keys to customer managed keys or vice versa requires the Replication Set and all associated data to be destroyed and recreated.

* `regions` - (Required) Regions in the replication Set. Each region must have a KMS encryption key which is optionally customer managed. A key value of `""` represents using a default AWS owned key to manage your resources. 

The following arguments are optional:

* `tags` - (Optional) Tags associated with the Replication Set.
  
Details for region and tag constraints can be found [here](https://docs.aws.amazon.com/incident-manager/latest/APIReference/API_CreateReplicationSet.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Replication Set.
* `tags_all` - List of all tags including both the resource and provider level tags.

## Timeouts

~> **NOTE:** For extreme use cases with very large amounts of response plans and data, update and delete operations may take up to a few hours. It is highly recommended that custom timeouts are configured in these circumstances

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`) The expected time to create a new Replication Set with one Region is up to 5 minutes. Expect creation to take up to an additional 20-30 minutes for each additional region.
* `update` - (Default `40m`) Expect each update operation to take up to 20-30 minutes.
* `delete` - (Default `40m`) Expect deleting the Replication Set to take up to 20-30 minutes.

## Import

SSM Incident Manager Incidents Replication Set can be imported 

```
$ terraform import aws_ssmincidents_replication_set.example import
```
