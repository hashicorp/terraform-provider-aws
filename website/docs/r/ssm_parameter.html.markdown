---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_parameter"
description: |-
  Provides a SSM Parameter resource
---

# Resource: aws_ssm_parameter

Provides an SSM Parameter resource.

~> **Note:** The `overwrite` argument makes it possible to overwrite an existing SSM Parameter created outside of Terraform.

-> **Note:** Write-Only argument `value_wo` is available to use in place of `value`. Write-Only arguments are supported in HashiCorp Terraform 1.11.0 and later. [Learn more](https://developer.hashicorp.com/terraform/language/v1.11.x/resources/ephemeral#write-only-arguments).

## Example Usage

### Basic example

```terraform
resource "aws_ssm_parameter" "foo" {
  name  = "foo"
  type  = "String"
  value = "bar"
}
```

### Encrypted string using default SSM KMS key

```terraform
resource "aws_db_instance" "default" {
  allocated_storage    = 10
  storage_type         = "gp2"
  engine               = "mysql"
  engine_version       = "5.7.16"
  instance_class       = "db.t2.micro"
  db_name              = "mydb"
  username             = "foo"
  password             = var.database_master_password
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "default.mysql5.7"
}

resource "aws_ssm_parameter" "secret" {
  name        = "/production/database/password/master"
  description = "The parameter description"
  type        = "SecureString"
  value       = var.database_master_password

  tags = {
    environment = "production"
  }
}
```

~> **Note:** The unencrypted value of a SecureString will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the parameter. If the name contains a path (e.g., any forward slashes (`/`)), it must be fully qualified with a leading forward slash (`/`). For additional requirements and constraints, see the [AWS SSM User Guide](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-parameter-name-constraints.html).
* `type` - (Required) Type of the parameter. Valid types are `String`, `StringList` and `SecureString`.

The following arguments are optional:

* `allowed_pattern` - (Optional) Regular expression used to validate the parameter value.
* `data_type` - (Optional) Data type of the parameter. Valid values: `text`, `aws:ssm:integration` and `aws:ec2:image` for AMI format, see the [Native parameter support for Amazon Machine Image IDs](https://docs.aws.amazon.com/systems-manager/latest/userguide/parameter-store-ec2-aliases.html).
* `description` - (Optional) Description of the parameter.
* `insecure_value` - (Optional, exactly one of `value`, `value_wo`  or `insecure_value` is required) Value of the parameter. **Use caution:** This value is _never_ marked as sensitive in the Terraform plan output. This argument is not valid with a `type` of `SecureString`.
* `key_id` - (Optional) KMS key ID or ARN for encrypting a SecureString.
* `overwrite` - (Optional) Overwrite an existing parameter. If not specified, defaults to `false` during create operations to avoid overwriting existing resources and then `true` for all subsequent operations once the resource is managed by Terraform. [Lifecycle rules](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle) should be used to manage non-standard update behavior.
* `tags` - (Optional) Map of tags to assign to the object. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tier` - (Optional) Parameter tier to assign to the parameter. If not specified, will use the default parameter tier for the region. Valid tiers are `Standard`, `Advanced`, and `Intelligent-Tiering`. Downgrading an `Advanced` tier parameter to `Standard` will recreate the resource. For more information on parameter tiers, see the [AWS SSM Parameter tier comparison and guide](https://docs.aws.amazon.com/systems-manager/latest/userguide/parameter-store-advanced-parameters.html).
* `value` - (Optional, exactly one of `value`, `value_wo` or `insecure_value` is required) Value of the parameter. This value is always marked as sensitive in the Terraform plan output, regardless of `type`. In Terraform CLI version 0.15 and later, this may require additional configuration handling for certain scenarios. For more information, see the [Terraform v0.15 Upgrade Guide](https://www.terraform.io/upgrade-guides/0-15.html#sensitive-output-values).
* `value_wo` - (Optional, Write-Only, exactly one of `value`, `value_wo` or `insecure_value` is required) Value of the parameter. This value is always marked as sensitive in the Terraform plan output, regardless of `type`. Additionally, `write-only` values are never stored to state. `value_wo_version` can be used to trigger an update and is required with this argument. In Terraform CLI version 0.15 and later, this may require additional configuration handling for certain scenarios. For more information, see the [Terraform v0.15 Upgrade Guide](https://www.terraform.io/upgrade-guides/0-15.html#sensitive-output-values).
* `value_wo_version` - (Optional) Used together with `value_wo` to trigger an update. Increment this value when an update to the `value_wo` is required.

~> **NOTE:** `aws:ssm:integration` data_type parameters must be of the type `SecureString` and the name must start with the prefix `/d9d01087-4a3f-49e0-b0b4-d568d7826553/ssm/integrations/webhook/`. See [here](https://docs.aws.amazon.com/systems-manager/latest/userguide/creating-integrations.html) for information on the usage of `aws:ssm:integration` parameters.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the parameter.
* `has_value_wo` - Indicates whether the resource has a `value_wo` set.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version` - Version of the parameter.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSM Parameters using the parameter store `name`. For example:

```terraform
import {
  to = aws_ssm_parameter.my_param
  id = "/my_path/my_paramname"
}
```

Using `terraform import`, import SSM Parameters using the parameter store `name`. For example:

```console
% terraform import aws_ssm_parameter.my_param /my_path/my_paramname
```
