---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_instance_profile"
description: |-
  Manages an AWS DMS (Database Migration) Instance Profile.
---

# Resource: aws_dms_instance_profile

Manages an AWS DMS (Database Migration) Instance Profile. An instance profile provides network and encryption information to DMS Schema Conversion for use in connecting to a data provider.

## Example Usage

### Basic Usage

```terraform
resource "aws_dms_instance_profile" "example" {
  name = "example"
}
```

### With Networking Configuration

```terraform
resource "aws_dms_instance_profile" "example" {
  name                    = "example"
  description             = "example instance profile"
  network_type            = "IPV4"
  publicly_accessible     = false
  subnet_group_identifier = aws_dms_replication_subnet_group.example.id
  vpc_security_group_ids  = [aws_security_group.example.id]
  kms_key_arn             = aws_kms_key.example.arn

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

The following arguments are optional:

* `availability_zone` - (Optional) Availability Zone where the instance profile runs. Default is a random, system-chosen Availability Zone.
* `description` - (Optional) Description for the instance profile.
* `kms_key_arn` - (Optional) ARN of the KMS key used to encrypt the connection parameters for the instance profile. If you don't specify a value, DMS uses your default encryption key.
* `name` - (Optional) Name for the instance profile. If omitted, Terraform will assign a random, unique name.
* `network_type` - (Optional) Network type for the instance profile. Valid values are `IPV4`, `IPV6`, and `DUAL`.
* `publicly_accessible` - (Optional) Whether the instance profile is publicly accessible. Default is `true`.
* `subnet_group_identifier` - (Optional) Subnet group to associate with the instance profile.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_security_group_ids` - (Optional) VPC security group IDs to be used with the instance profile. The VPC security groups must work with the VPC containing the instance profile.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the instance profile.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_dms_instance_profile.example
  identity = {
    "arn" = "arn:aws:dms:us-east-1:123456789012:instance-profile:ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"
  }
}

resource "aws_dms_instance_profile" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the instance profile.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DMS (Database Migration) Instance Profile using the `arn`. For example:

```terraform
import {
  to = aws_dms_instance_profile.example
  id = "arn:aws:dms:us-east-1:123456789012:instance-profile:ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"
}
```

Using `terraform import`, import DMS (Database Migration) Instance Profile using the `arn`. For example:

```console
% terraform import aws_dms_instance_profile.example arn:aws:dms:us-east-1:123456789012:instance-profile:ABCDEFGHIJKLMNOPQRSTUVWXYZ123456
```
