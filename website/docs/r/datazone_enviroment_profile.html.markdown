---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_environment_profile"
description: |-
  Terraform resource for managing an AWS DataZone Environment Profile.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->`
# Resource: aws_datazone_environment_profile

Terraform resource for managing an AWS DataZone Environment Profile.

## Example Usage

### Basic Usage

```terraform
resource "aws_datazone_environment_profile" "example" {
	aws_account_id = data.aws_caller_identity.example.account_id
	aws_account_region = data.aws_region.example.name
	environment_blueprint_identifier = data.aws_datazone_environment_blueprint.example.id
	description = "desc"
	name = "name"
	project_identifier = aws_datazone_project.example.id
	domain_identifier = aws_datazone_domain.example.i

}
```

## Argument Reference

The following arguments are required:

* `aws_account_Id` - (Required) -  Id of the AWS account being used. Must follow regex of ^\d{12}$.
* `aws_account_region` - (Required) -  Desired region for environment profile. Must follow regex of ^[a-z]{2}-[a-z]{4,10}-\d$.
* `domain_identifier` - (Required) -  Domain Identifier for environment profile.
* `name` - (Required) -  Name of the environment profile. Must follow regex of ^[\w -]+$ and have the length between 1 and 64.
* `environment_blueprint_identifier` - (Required) -  ID of the blueprint which the environment will be created with. Must follow regex of ^[a-zA-Z0-9_-]{1,36}$.
* `project_identifier` - (Required) -  Project identifier for environment profile. Must follow regex of ^[a-zA-Z0-9_-]{1,36}$.

The following arguments are optional:

* `description` - (Optional) Description of environment profile. Must be between the length of 0 and 2048.
* `aws_account_Id` - (Optional) -  Id of the AWS account being used. Must follow regex of ^\d{12}$
* `user_parameters` - (Optional) -  Array of user parameters of the environment profile with the following attributes:
    * `name` - (Required) -  Name of the environment profile parameter.
    * `value` - (Required) -  Value of the environment profile parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Creation time of environment profile.
* `created_by` - Creator of environment profile.
* `id` - ID of environment profile. 
* `updated_at` - Time of last update to environment profile.
* `user_parameters` - Array of user parameters of the environment profile with the following attributes:
    * `field_type` - Filed type of the parameter.
    * `key_name` - Key name of the parameter.
    * `default_value` -  Default value of the parameter.
    * `description` - Description of the parameter.
    * `is_editable` -  Bool that specifies if the parameter is editable.
    * `is_optional` - Bool that specifies if the parameter is editable.

## Timeouts

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Environment Profile using the `example_id_arg`. For example:

```terraform
import {
  to = aws_datazone_environment_profile.example
  id = "environment_profile-id-12345678"
}
```

Using `terraform import`, import DataZone Environment Profile using the `example_id_arg`. For example:

```console
% terraform import aws_datazone_environment_profile.example environment_profile-id-12345678
```
