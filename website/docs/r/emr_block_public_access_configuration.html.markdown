---
subcategory: "EMR"
layout: "aws"
page_title: "AWS: aws_emr_block_public_access_configuration"
description: |-
  Terraform resource for managing an AWS EMR Block Public Access Configuration.
---

# Resource: aws_emr_block_public_access_configuration

Terraform resource for managing an AWS EMR block public access configuration. This region level security configuration restricts the launch of EMR clusters that have associated security groups permitting public access on unspecified ports. See the [EMR Block Public Access Configuration](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-block-public-access.html) documentation for further information.

## Example Usage

### Basic Usage

```terraform
resource "aws_emr_block_public_access_configuration" "example" {
  block_public_security_group_rules = true
}
```

### Default Configuration

By default, each AWS region is equipped with a block public access configuration that prevents EMR clusters from being launched if they have security group rules permitting public access on any port except for port 22. The default configuration can be managed using this Terraform resource.

```terraform
resource "aws_emr_block_public_access_configuration" "example" {
  block_public_security_group_rules = true

  permitted_public_security_group_rule_range {
    min_range = 22
    max_range = 22
  }
}
```

~> **NOTE:** If an `aws_emr_block_public_access_configuration` Terraform resource is destroyed, the configuration will reset to this default configuration.

### Multiple Permitted Public Security Group Rule Ranges

The resource permits specification of multiple `permitted_public_security_group_rule_range` blocks.

```terraform
resource "aws_emr_block_public_access_configuration" "example" {
  block_public_security_group_rules = true

  permitted_public_security_group_rule_range {
    min_range = 22
    max_range = 22
  }

  permitted_public_security_group_rule_range {
    min_range = 100
    max_range = 101
  }
}
```

### Disabling Block Public Access

To permit EMR clusters to be launched in the configured region regardless of associated security group rules, the Block Public Access feature can be disabled using this Terraform resource.

```terraform
resource "aws_emr_block_public_access_configuration" "example" {
  block_public_security_group_rules = false
}
```

## Argument Reference

The following arguments are required:

* `block_public_security_group_rules` - (Required) Enable or disable EMR Block Public Access.

The following arguments are optional:

* `permitted_public_security_group_rule_range` - (Optional) Configuration block for defining permitted public security group rule port ranges. Can be defined multiple times per resource. Only valid if `block_public_security_group_rules` is set to `true`.

### `permitted_public_security_group_rule_range`

This block is used to define a range of TCP ports that should form exceptions to the Block Public Access Configuration. If an attempt is made to launch an EMR cluster in the configured region and account, with `block_public_security_group_rules = true`, the EMR cluster will be permitted to launch even if there are security group rules permitting public access to ports in this range.

* `min_range` - (Required) The first port in the range of TCP ports.
* `max_range` - (Required) The final port in the range of TCP ports.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the current EMR Block Public Access Configuration. For example:

```terraform
import {
  to = aws_emr_block_public_access_configuration.example
  id = "current"
}
```

Using `terraform import`, import the current EMR Block Public Access Configuration. For example:

```console
% terraform import aws_emr_block_public_access_configuration.example current
```
