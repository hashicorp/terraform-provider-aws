---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_capacity_manager_settings"
description: |-
  Manages EC2 Capacity Manager settings for the current AWS account.
---

# Resource: aws_ec2_capacity_manager_settings

Manages [EC2 Capacity Manager](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/capacity-manager.html) settings for the current AWS account. EC2 Capacity Manager monitors, analyzes, and helps manage EC2 capacity usage across On-Demand Instances, Spot Instances, and Capacity Reservations.

Capacity Manager can be enabled in only one Region per account (the home Region), where it aggregates capacity data from all Regions. After enabling, initial data ingestion may take several hours to complete.

~> **Note:** Destroying this resource disables EC2 Capacity Manager, which also disables Organizations access.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_capacity_manager_settings" "example" {
  enabled = true
}
```

### Organizations Access

To aggregate data from all accounts in an AWS Organization, enable Organizations access from the management account (or a delegated administrator account):

```terraform
resource "aws_ec2_capacity_manager_settings" "example" {
  enabled              = true
  organizations_access = true
}
```

### Organizations Access with a Delegated Administrator

```terraform
data "aws_caller_identity" "delegated" {
  provider = aws.delegated
}

resource "aws_organizations_delegated_administrator" "example" {
  account_id        = data.aws_caller_identity.delegated.account_id
  service_principal = "ec2.capacitymanager.amazonaws.com"
}

resource "aws_ec2_capacity_manager_settings" "example" {
  provider = aws.delegated

  enabled              = true
  organizations_access = true

  depends_on = [aws_organizations_delegated_administrator.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `enabled` - (Required) Whether to enable EC2 Capacity Manager for the account in the configured Region.
* `organizations_access` - (Optional) Whether to enable cross-account data aggregation for AWS Organizations. Requires `enabled` to be `true`. Can only be enabled from the Organizations management account or a delegated administrator account. Defaults to `false`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_capacity_manager_settings.example
  identity = {
    region = "us-west-2"
  }
}

resource "aws_ec2_capacity_manager_settings" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Capacity Manager settings using the region name. For example:

```terraform
import {
  to = aws_ec2_capacity_manager_settings.example
  id = "us-west-2"
}
```

Using `terraform import`, import EC2 Capacity Manager settings using the region name. For example:

```console
% terraform import aws_ec2_capacity_manager_settings.example us-west-2
```
