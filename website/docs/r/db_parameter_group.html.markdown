---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_parameter_group"
description: |-
  Provides an RDS DB parameter group resource.
---

# Resource: aws_db_parameter_group

Provides an RDS DB parameter group resource. Documentation of the available parameters for various RDS engines can be found at:

* [Aurora MySQL Parameters](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraMySQL.Reference.html)
* [Aurora PostgreSQL Parameters](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraPostgreSQL.Reference.html)
* [MariaDB Parameters](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.MariaDB.Parameters.html)
* [Oracle Parameters](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_ModifyInstance.Oracle.html#USER_ModifyInstance.Oracle.sqlnet)
* [PostgreSQL Parameters](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.PostgreSQL.CommonDBATasks.html#Appendix.PostgreSQL.CommonDBATasks.Parameters)

> **Hands-on:** For an example of the `aws_db_parameter_group` in use, follow the [Manage AWS RDS Instances](https://learn.hashicorp.com/tutorials/terraform/aws-rds?in=terraform/aws&utm_source=WEBSITE&utm_medium=WEB_IO&utm_offer=ARTICLE_PAGE&utm_content=DOCS) tutorial on HashiCorp Learn.

~> **NOTE:** If you encounter a Terraform plan showing parameter changes after an apply (_i.e._, _perpetual diffs_), see the [Problematic Plan Changes](#problematic-plan-changes) example below for additional guidance.

## Example Usage

### Basic Usage

```terraform
resource "aws_db_parameter_group" "default" {
  name   = "rds-pg"
  family = "mysql5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }
}
```

### `create_before_destroy` Lifecycle Configuration

The [`create_before_destroy`](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#create_before_destroy)
lifecycle configuration is necessary for modifications that force re-creation of an existing,
in-use parameter group. This includes common situations like changing the group `name` or
bumping the `family` version during a major version upgrade. This configuration will prevent destruction
of the deposed parameter group while still in use by the database during upgrade.

Note: Using `create_before_destroy` requires that the new parameter group is created with a different
name than the existing one. This can be achieved by setting `name_prefix` instead of `name`, for example.

```terraform
resource "aws_db_parameter_group" "example" {
  name_prefix = "my-pg"
  family      = "postgres13"

  parameter {
    name  = "log_connections"
    value = "1"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_db_instance" "example" {
  # other attributes
  parameter_group_name = aws_db_parameter_group.example.name
  apply_immediately    = true
}
```

### Problematic Plan Changes

If you are experiencing unexpected `update in-place` plan changes after running `terraform apply` (_i.e._, "perpetual diffs"), it is likely due to conflicts between the AWS Provider's default behavior and AWS's requirements for managing parameter groups. The following characteristics of parameter management are relevant:

1. The AWS Provider's default `apply_method` is `immediate`.
2. AWS automatically assigns default parameters with predefined values and `apply_method` settings when you create a parameter group.
3. AWS does not allow changing the `apply_method` of a default parameter (or an existing parameter) without also modifying its `value`. For example, you cannot change the `apply_method` from `pending-reboot` to `immediate` or vice versa without adjusting the parameter's value.

See an example of this type of problem and solutions below.

#### Example of Problematic Configuration

The following Terraform configuration includes a parameter that overlaps with an AWS default parameter, using the same `name` (`default_password_lifetime`) and `value` (`0`). However:

- AWS sets the default `apply_method` for this parameter to `pending-reboot`.
- The AWS Provider defaults all parameters' `apply_method` to `immediate`.

This configuration attempts to change _only_ the `apply_method` from `pending-reboot` to `immediate`, which is not allowed by AWS.

```terraform
resource "aws_db_parameter_group" "test" {
  name   = "random-test-parameter"
  family = "mysql5.7"

  parameter {
    # By default, the apply_method is being set to "immediate"
    name  = "default_password_lifetime" # same as AWS default
    value = "0"                         # same as AWS default
  }
}
```

#### Solution 1: Remove the Default Parameter

Exclude the default parameter, such as `default_password_lifetime` in this example, from your configuration entirely. This ensures Terraform does not attempt to modify the parameter, leaving it with AWS's default settings.

```terraform
resource "aws_db_parameter_group" "test" {
  name   = "random-test-parameter"
  family = "mysql5.7"
}
```

#### Solution 2: Modify the Parameter's Value Also

Change the `value` of the parameter along with its `apply_method`. Since the AWS default `value` is `0`, selecting any other valid value (_e.g._, `1`) will resolve the issue.

```terraform
resource "aws_db_parameter_group" "test" {
  name   = "random-test-parameter"
  family = "mysql5.7"

  parameter {
    # Because of the default, the apply_method will also be changed from `pending-reboot` to `immediate`
    name  = "default_password_lifetime" # same as AWS default
    value = "1"                         # different from AWS default, "0"
  }
}
```

#### Solution 3: Align `apply_method` with AWS Defaults

Explicitly set the `apply_method` to match AWS's default value for this parameter (`pending-reboot`). This prevents conflicts between Terraform's default (`immediate`) and AWS's default where the `value` is not changing.

```terraform
resource "aws_db_parameter_group" "test" {
  name   = "random-test-parameter"
  family = "mysql5.7"

  parameter {
    apply_method = "pending-reboot"            # same as AWS default
    name         = "default_password_lifetime" # same as AWS default
    value        = "0"                         # same as AWS default
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Optional, Forces new resource) The name of the DB parameter group. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `family` - (Required, Forces new resource) The family of the DB parameter group.
* `description` - (Optional, Forces new resource) The description of the DB parameter group. Defaults to "Managed by Terraform".
* `parameter` - (Optional) The DB parameters to apply. See [`parameter` Block](#parameter-block) below for more details. Note that parameters may differ from a family to an other. Full list of all parameters can be discovered via [`aws rds describe-db-parameters`](https://docs.aws.amazon.com/cli/latest/reference/rds/describe-db-parameters.html) after initial creation of the group.
* `skip_destroy` - (Optional) Set to true if you do not wish the parameter group to be deleted at destroy time, and instead just remove the parameter group from the Terraform state.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `parameter` Block

The `parameter` blocks support the following arguments:

* `name` - (Required) The name of the DB parameter.
* `value` - (Required) The value of the DB parameter.
* `apply_method` - (Optional) "immediate" (default), or "pending-reboot". Some
    engines can't apply some parameters without a reboot, and you will need to
    specify "pending-reboot" here.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The db parameter group name.
* `arn` - The ARN of the db parameter group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DB Parameter groups using the `name`. For example:

```terraform
import {
  to = aws_db_parameter_group.rds_pg
  id = "rds-pg"
}
```

Using `terraform import`, import DB Parameter groups using the `name`. For example:

```console
% terraform import aws_db_parameter_group.rds_pg rds-pg
```
