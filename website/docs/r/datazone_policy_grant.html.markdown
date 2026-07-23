---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_policy_grant"
description: |-
  Manages an AWS DataZone Policy Grant.
---

# Resource: aws_datazone_policy_grant

Manages an AWS DataZone Policy Grant.

## Example Usage

### Basic Usage

```terraform
resource "aws_datazone_policy_grant" "example" {
  domain_identifier = aws_datazone_domain.example.id
  entity_type       = "DOMAIN_UNIT"
  entity_identifier = aws_datazone_domain.example.root_domain_unit_id
  policy_type       = "CREATE_DOMAIN_UNIT"

  detail {
    create_domain_unit {}
  }

  principal {
    user {
      all_users_grant_filter {}
    }
  }
}
```

### With Include Child Domain Units

```terraform
resource "aws_datazone_policy_grant" "example" {
  domain_identifier = aws_datazone_domain.example.id
  entity_type       = "DOMAIN_UNIT"
  entity_identifier = aws_datazone_domain.example.root_domain_unit_id
  policy_type       = "CREATE_DOMAIN_UNIT"

  detail {
    create_domain_unit {
      include_child_domain_units = true
    }
  }

  principal {
    user {
      all_users_grant_filter {}
    }
  }
}
```

### With Project Principal

```terraform
resource "aws_datazone_policy_grant" "example" {
  domain_identifier = aws_datazone_domain.example.id
  entity_type       = "DOMAIN_UNIT"
  entity_identifier = aws_datazone_domain.example.root_domain_unit_id
  policy_type       = "CREATE_GLOSSARY"

  detail {
    create_glossary {}
  }

  principal {
    project {
      project_designation = "OWNER"
      project_identifier  = aws_datazone_project.example.id
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `domain_identifier` - (Required) Identifier of the domain where the policy grant is created.
* `entity_type` - (Required) Type of entity to which the policy grant applies. Valid values: `ASSET_TYPE`, `DOMAIN_UNIT`, `ENVIRONMENT_BLUEPRINT_CONFIGURATION`, `ENVIRONMENT_PROFILE`.
* `entity_identifier` - (Required) Identifier of the entity to which the policy grant applies.
* `policy_type` - (Required) Type of the managed policy. Valid values: `ADD_TO_PROJECT_MEMBER_POOL`, `CREATE_ASSET_TYPE`, `CREATE_DOMAIN_UNIT`, `CREATE_ENVIRONMENT`, `CREATE_ENVIRONMENT_FROM_BLUEPRINT`, `CREATE_ENVIRONMENT_PROFILE`, `CREATE_FORM_TYPE`, `CREATE_GLOSSARY`, `CREATE_PROJECT`, `CREATE_PROJECT_FROM_PROJECT_PROFILE`, `DELEGATE_CREATE_ENVIRONMENT_PROFILE`, `OVERRIDE_DOMAIN_UNIT_OWNERS`, `OVERRIDE_PROJECT_OWNERS`, `USE_ASSET_TYPE`.
* `detail` - (Required) Policy grant detail. See [`detail`](#detail) below.
* `principal` - (Required) Principal to which the policy grant applies. See [`principal`](#principal) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `detail`

Exactly one of the following sub-blocks must be specified, corresponding to the `policy_type`:

* `add_to_project_member_pool` - (Optional) Configuration for the `ADD_TO_PROJECT_MEMBER_POOL` policy type. See [`include_child_domain_units` detail](#include_child_domain_units-detail) below.
* `create_asset_type` - (Optional) Configuration for the `CREATE_ASSET_TYPE` policy type. See [`include_child_domain_units` detail](#include_child_domain_units-detail) below.
* `create_domain_unit` - (Optional) Configuration for the `CREATE_DOMAIN_UNIT` policy type. See [`include_child_domain_units` detail](#include_child_domain_units-detail) below.
* `create_environment` - (Optional) Configuration for the `CREATE_ENVIRONMENT` policy type. Empty block.
* `create_environment_from_blueprint` - (Optional) Configuration for the `CREATE_ENVIRONMENT_FROM_BLUEPRINT` policy type. Empty block.
* `create_environment_profile` - (Optional) Configuration for the `CREATE_ENVIRONMENT_PROFILE` policy type. See [`domain_unit_id` detail](#domain_unit_id-detail) below.
* `create_form_type` - (Optional) Configuration for the `CREATE_FORM_TYPE` policy type. See [`include_child_domain_units` detail](#include_child_domain_units-detail) below.
* `create_glossary` - (Optional) Configuration for the `CREATE_GLOSSARY` policy type. See [`include_child_domain_units` detail](#include_child_domain_units-detail) below.
* `create_project` - (Optional) Configuration for the `CREATE_PROJECT` policy type. See [`include_child_domain_units` detail](#include_child_domain_units-detail) below.
* `create_project_from_project_profile` - (Optional) Configuration for the `CREATE_PROJECT_FROM_PROJECT_PROFILE` policy type. See [`create_project_from_project_profile` detail](#create_project_from_project_profile-detail) below.
* `delegate_create_environment_profile` - (Optional) Configuration for the `DELEGATE_CREATE_ENVIRONMENT_PROFILE` policy type. Empty block.
* `override_domain_unit_owners` - (Optional) Configuration for the `OVERRIDE_DOMAIN_UNIT_OWNERS` policy type. See [`include_child_domain_units` detail](#include_child_domain_units-detail) below.
* `override_project_owners` - (Optional) Configuration for the `OVERRIDE_PROJECT_OWNERS` policy type. See [`include_child_domain_units` detail](#include_child_domain_units-detail) below.
* `use_asset_type` - (Optional) Configuration for the `USE_ASSET_TYPE` policy type. See [`domain_unit_id` detail](#domain_unit_id-detail) below.

### `include_child_domain_units` detail

Used by `add_to_project_member_pool`, `create_asset_type`, `create_domain_unit`, `create_form_type`, `create_glossary`, `create_project`, `override_domain_unit_owners`, and `override_project_owners`.

* `include_child_domain_units` - (Optional) Whether to include child domain units.

### `domain_unit_id` detail

Used by `create_environment_profile` and `use_asset_type`.

* `domain_unit_id` - (Optional) Identifier of the domain unit.

### `create_project_from_project_profile` detail

* `include_child_domain_units` - (Optional) Whether to include child domain units.
* `project_profiles` - (Optional) List of project profile identifiers.

### `principal`

Exactly one of the following sub-blocks must be specified:

* `domain_unit` - (Optional) Domain unit principal. See [`domain_unit`](#domain_unit) below.
* `group` - (Optional) Group principal. See [`group`](#group) below.
* `project` - (Optional) Project principal. See [`project`](#project) below.
* `user` - (Optional) User principal. See [`user`](#user) below.

### `domain_unit`

* `domain_unit_designation` - (Required) Designation of the domain unit principal. Valid values: `OWNER`.
* `domain_unit_identifier` - (Optional) Identifier of the domain unit.
* `all_domain_units_grant_filter` - (Optional) Filter to grant access to all domain units. Empty block.

### `group`

* `group_identifier` - (Required) Identifier of the group principal.

### `project`

* `project_designation` - (Required) Designation of the project principal. Valid values: `CONTRIBUTOR`, `OWNER`, `PROJECT_CATALOG_STEWARD`.
* `project_identifier` - (Optional) Identifier of the project.
* `domain_unit_filter` - (Optional) Filter for domain unit scoping. See [`domain_unit_filter`](#domain_unit_filter) below.

### `domain_unit_filter`

* `domain_unit` - (Required) Identifier of the domain unit for filtering.
* `include_child_domain_units` - (Optional) Whether to include child domain units in the filter.

### `user`

* `user_identifier` - (Optional) Identifier of the user principal.
* `all_users_grant_filter` - (Optional) Filter to grant access to all users. Empty block.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Timestamp when the policy grant was created (RFC3339 format).
* `created_by` - User who created the policy grant.
* `grant_id` - Identifier of the policy grant.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_datazone_policy_grant.example
  identity = {
    domain_identifier = "dzd_54nakfrg9k6sri"
    entity_type       = "DOMAIN_UNIT"
    entity_identifier = "9v3oj4n26k4yrq"
    policy_type       = "CREATE_DOMAIN_UNIT"
    grant_id          = "3v8lox42tj5zic"
  }
}

resource "aws_datazone_policy_grant" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `domain_identifier` (String) Identifier of the domain.
* `entity_type` (String) Type of entity.
* `entity_identifier` (String) Identifier of the entity.
* `policy_type` (String) Type of the managed policy.
* `grant_id` (String) Identifier of the policy grant.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Policy Grant using the `domain_identifier,entity_type,entity_identifier,policy_type,grant_id`. For example:

```terraform
import {
  to = aws_datazone_policy_grant.example
  id = "dzd_54nakfrg9k6sri,DOMAIN_UNIT,9v3oj4n26k4yrq,CREATE_DOMAIN_UNIT,3v8lox42tj5zic"
}
```

Using `terraform import`, import DataZone Policy Grant using the `domain_identifier,entity_type,entity_identifier,policy_type,grant_id`. For example:

```console
% terraform import aws_datazone_policy_grant.example dzd_54nakfrg9k6sri,DOMAIN_UNIT,9v3oj4n26k4yrq,CREATE_DOMAIN_UNIT,3v8lox42tj5zic
```
