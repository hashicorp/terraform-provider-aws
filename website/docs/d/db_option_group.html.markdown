---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_option_group"
description: |-
  Terraform data source for managing an AWS RDS DB Option Group.
---

# Data Source: aws_db_option_group

Terraform data source for managing an AWS RDS DB Option Group.

## Example Usage

```terraform
data "aws_db_option_group" "example" {
    option_group_name = "option-group-test-terraform"
}
```

## Argument Reference

The following arguments are required:

* `option_group_name` - The name of the option group to describe.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `allow_vpc_and_non_vpc_instance_membership` - Indicates whether this option group can be applied to both VPC and non-VPC instances. The value true indicates the option group can be applied to both VPC and non-VPC instances.
* `copy_timestamp` - Indicates when the option group was copied.
* `engine_name` - Indicates the name of the engine that this option group can be applied to.
* `major_engine_version` - Indicates the major engine version associated with this option group.
* `option_group_arn` - Specifies the Amazon Resource Name (ARN) for the option group.
* `option_group_description` - Provides a description of the option group.
* `options` - Indicates what options are available in the option group. See the [`options` attribute reference](#options-attribute-reference) below.
* `source_account_id` - Specifies the Amazon Web Services account ID for the option group from which this option group is copied.
* `source_option_group` - Specifies the name of the option group from which this option group is copied.
* `vpc_id` - If AllowsVpcAndNonVpcInstanceMemberships is false , this field is blank. If AllowsVpcAndNonVpcInstanceMemberships is true and this field is blank, then this option group can be applied to both VPC and non-VPC instances. If this field contains a value, then this option group can only be applied to instances that are in the VPC indicated by this field.

### `options` Attribute Reference

* `option_name` - The name of the option.
* `option_description` - The description of the option.
* `persistent` - Indicates whether this option is persistent.
* `permanent` - Indicates whether this option is permanent.
* `port` - If required, the port configured for this option to use.
* `option_version` - The version of the option.
* `option_settings` - The option settings for this option. See the [`option_settings` attribute reference](#options-settings-attribute-reference) below.
* `db_security_group_membership` - If the option requires access to a port, then this DB security group allows access to the port. See the [`db_security_group_membership` attribute reference](#db-security-group-membership-attribute-reference) below.
* `vpc_security_group_memberships` - If the option requires access to a port, then this VPC security group allows access to the port. See the [`vpc_security_group_memberships` attribute reference](#vpc-security-group-memberships) below.

### `option_settings` Attribue Reference

* `name` - The name of the option that has settings that you can set.
* `value` - The current value of the option setting.
* `default_value` - The default value of the option setting.
* `description` - The description of the option setting.
* `apply_type` - The DB engine specific parameter type.
* `data_type` - The data type of the option setting.
* `allowed_values` - The allowed values of the option setting.
* `is_modifiable` - Indicates whether the option setting can be modified from the default.
* `is_collection` - Indicates whether the option setting is part of a collection.

### `db_security_group_membership` Attribute Reference

* `db_security_group_name` - The name of the DB security group.
* `status` - The status of the DB security group.

### `vpc_security_group_memberships` Attribute Reference

* `vpc_security_group_id` - The name of the VPC security group.
* `status` - The membership status of the VPC security group. 
