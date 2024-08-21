---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_option_group"
description: |-
  Provides an RDS DB option group resource.
---

# Resource: aws_db_option_group

Provides an RDS DB option group resource. Documentation of the available options for various RDS engines can be found at:

* [MariaDB Options](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.MariaDB.Options.html)
* [Microsoft SQL Server Options](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.SQLServer.Options.html)
* [MySQL Options](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.MySQL.Options.html)
* [Oracle Options](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.Oracle.Options.html)

## Example Usage

```terraform
resource "aws_db_option_group" "example" {
  name                     = "option-group-test-terraform"
  option_group_description = "Terraform Option Group"
  engine_name              = "sqlserver-ee"
  major_engine_version     = "11.00"

  option {
    option_name = "Timezone"

    option_settings {
      name  = "TIME_ZONE"
      value = "UTC"
    }
  }

  option {
    option_name = "SQLSERVER_BACKUP_RESTORE"

    option_settings {
      name  = "IAM_ROLE_ARN"
      value = aws_iam_role.example.arn
    }
  }

  option {
    option_name = "TDE"
  }
}
```

~> **Note:** Any modifications to the `aws_db_option_group` are set to happen immediately as we default to applying immediately.

~> **WARNING:** You can perform a destroy on a `aws_db_option_group`, as long as it is not associated with any Amazon RDS resource. An option group can be associated with a DB instance, a manual DB snapshot, or an automated DB snapshot.

If you try to delete an option group that is associated with an Amazon RDS resource, an error similar to the following is returned:

> An error occurred (InvalidOptionGroupStateFault) when calling the DeleteOptionGroup operation: The option group 'optionGroupName' cannot be deleted because it is in use.

More information about this can be found [here](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_WorkingWithOptionGroups.html#USER_WorkingWithOptionGroups.Delete).

## Argument Reference

This resource supports the following arguments:

* `name` - (Optional, Forces new resource) Name of the option group. If omitted, Terraform will assign a random, unique name. Must be lowercase, to match as it is stored in AWS.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`. Must be lowercase, to match as it is stored in AWS.
* `option_group_description` - (Optional) Description of the option group. Defaults to "Managed by Terraform".
* `engine_name` - (Required) Specifies the name of the engine that this option group should be associated with.
* `major_engine_version` - (Required) Specifies the major version of the engine that this option group should be associated with.
* `option` - (Optional) The options to apply. See [`option` Block](#option-block) below for more details.
* `skip_destroy` - (Optional) Set to true if you do not wish the option group to be deleted at destroy time, and instead just remove the option group from the Terraform state.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `option` Block

The `option` blocks support the following arguments:

* `option_name` - (Required) Name of the option (e.g., MEMCACHED).
* `option_settings` - (Optional) The option settings to apply. See [`option_settings` Block](#option_settings-block) below for more details.
* `port` - (Optional) Port number when connecting to the option (e.g., 11211). Leaving out or removing `port` from your configuration does not remove or clear a port from the option in AWS. AWS may assign a default port. Not including `port` in your configuration means that the AWS provider will ignore a previously set value, a value set by AWS, and any port changes.
* `version` - (Optional) Version of the option (e.g., 13.1.0.0). Leaving out or removing `version` from your configuration does not remove or clear a version from the option in AWS. AWS may assign a default version. Not including `version` in your configuration means that the AWS provider will ignore a previously set value, a value set by AWS, and any version changes.
* `db_security_group_memberships` - (Optional) List of DB Security Groups for which the option is enabled.
* `vpc_security_group_memberships` - (Optional) List of VPC Security Groups for which the option is enabled.

#### `option_settings` Block

The `option_settings` blocks support the following arguments:

* `name` - (Optional) Name of the setting.
* `value` - (Optional) Value of the setting.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - DB option group name.
* `arn` - ARN of the DB option group.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `delete` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DB option groups using the `name`. For example:

```terraform
import {
  to = aws_db_option_group.example
  id = "mysql-option-group"
}
```

Using `terraform import`, import DB option groups using the `name`. For example:

```console
% terraform import aws_db_option_group.example mysql-option-group
```
