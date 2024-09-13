---
subcategory: "SSM Incident Manager Incidents"
layout: "aws"
page_title: "AWS: aws_ssmincidents_replication_set"
description: |-
  Terraform data source for managing an incident replication set in AWS Systems Manager Incident Manager.
---


# Data Source: aws_ssmincidents_replication_set

~> **NOTE:** The AWS Region specified by a Terraform provider must always be one of the Regions specified for the replication set.

Use this Terraform data source to manage a replication set in AWS Systems Manager Incident Manager.

## Example Usage

### Basic Usage

```terraform
data "aws_ssmincidents_replication_set" "example" {}
```

## Argument Reference

No arguments are required.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resouce Name (ARN) of the replication set.
* `tags` - All tags applied to the replication set.
* `created_by` - The ARN of the user who created the replication set.
* `deletion_protected` - If `true`, the last remaining Region in a replication set can’t be deleted.
* `last_modified_by` - The ARN of the user who last modified the replication set.
* `status` - The overall status of a replication set.
    * Valid Values: `ACTIVE` | `CREATING` | `UPDATING` | `DELETING` | `FAILED`

The `region` configuration block exports the following attributes for each Region:

* `name` - The name of the Region.
* `kms_key_arn` - The ARN of the AWS Key Management Service (AWS KMS) encryption key.
* `status` - The current status of the Region.
    * Valid Values: `ACTIVE` | `CREATING` | `UPDATING` | `DELETING` | `FAILED`
* `status_message` - More information about the status of a Region.
