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
  entity_identifier = aws_datazone_domain.example.root_domain_unit_id
  entity_type       = "DOMAIN_UNIT"
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
  entity_identifier = aws_datazone_domain.example.root_domain_unit_id
  entity_type       = "DOMAIN_UNIT"
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
  entity_identifier = aws_datazone_domain.example.root_domain_unit_id
  entity_type       = "DOMAIN_UNIT"
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

* `detail` - (Required) Policy grant detail. Exactly one sub-block must be specified. See [`detail` Block](#detail-block) below.
* `domain_identifier` - (Required) Identifier of the domain where the policy grant is created.
* `entity_identifier` - (Required) Identifier of the entity to which the policy grant applies.
* `entity_type` - (Required) Type of entity to which the policy grant applies. Valid values: `ASSET_TYPE`, `DOMAIN_UNIT`, `ENVIRONMENT_BLUEPRINT_CONFIGURATION`, `ENVIRONMENT_PROFILE`.
* `policy_type` - (Required) Type of the managed policy. Valid values: `ADD_TO_PROJECT_MEMBER_POOL`, `CREATE_ASSET_TYPE`, `CREATE_DOMAIN_UNIT`, `CREATE_ENVIRONMENT`, `CREATE_ENVIRONMENT_FROM_BLUEPRINT`, `CREATE_ENVIRONMENT_PROFILE`, `CREATE_FORM_TYPE`, `CREATE_GLOSSARY`, `CREATE_PROJECT`, `CREATE_PROJECT_FROM_PROJECT_PROFILE`, `DELEGATE_CREATE_ENVIRONMENT_PROFILE`, `OVERRIDE_DOMAIN_UNIT_OWNERS`, `OVERRIDE_PROJECT_OWNERS`, `USE_ASSET_TYPE`.
* `principal` - (Required) Principal to which the policy grant applies. Exactly one sub-block must be specified. See [`principal` Block](#principal-block) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `detail` Block

Exactly one of the following sub-blocks must be specified, corresponding to the `policy_type`:

* `add_to_project_member_pool` - (Optional) Configuration for the `ADD_TO_PROJECT_MEMBER_POOL` policy type. See [`add_to_project_member_pool` Block](#add_to_project_member_pool-block) below.
* `create_asset_type` - (Optional) Configuration for the `CREATE_ASSET_TYPE` policy type. See [`create_asset_type` Block](#create_asset_type-block) below.
* `create_domain_unit` - (Optional) Configuration for the `CREATE_DOMAIN_UNIT` policy type. See [`create_domain_unit` Block](#create_domain_unit-block) below.
* `create_environment` - (Optional) Configuration for the `CREATE_ENVIRONMENT` policy type. Empty block.
* `create_environment_from_blueprint` - (Optional) Configuration for the `CREATE_ENVIRONMENT_FROM_BLUEPRINT` policy type. Empty block.
* `create_environment_profile` - (Optional) Configuration for the `CREATE_ENVIRONMENT_PROFILE` policy type. See [`create_environment_profile` Block](#create_environment_profile-block) below.
* `create_form_type` - (Optional) Configuration for the `CREATE_FORM_TYPE` policy type. See [`create_form_type` Block](#create_form_type-block) below.
* `create_glossary` - (Optional) Configuration for the `CREATE_GLOSSARY` policy type. See [`create_glossary` Block](#create_glossary-block) below.
* `create_project` - (Optional) Configuration for the `CREATE_PROJECT` policy type. See [`create_project` Block](#create_project-block) below.
* `create_project_from_project_profile` - (Optional) Configuration for the `CREATE_PROJECT_FROM_PROJECT_PROFILE` policy type. See [`create_project_from_project_profile` Block](#create_project_from_project_profile-block) below.
* `delegate_create_environment_profile` - (Optional) Configuration for the `DELEGATE_CREATE_ENVIRONMENT_PROFILE` policy type. Empty block.
* `override_domain_unit_owners` - (Optional) Configuration for the `OVERRIDE_DOMAIN_UNIT_OWNERS` policy type. See [`override_domain_unit_owners` Block](#override_domain_unit_owners-block) below.
* `override_project_owners` - (Optional) Configuration for the `OVERRIDE_PROJECT_OWNERS` policy type. See [`override_project_owners` Block](#override_project_owners-block) below.
* `use_asset_type` - (Optional) Configuration for the `USE_ASSET_TYPE` policy type. See [`use_asset_type` Block](#use_asset_type-block) below.

### `add_to_project_member_pool` Block

* `include_child_domain_units` - (Optional) Whether to include child domain units.

### `create_asset_type` Block

* `include_child_domain_units` - (Optional) Whether to include child domain units.

### `create_domain_unit` Block

* `include_child_domain_units` - (Optional) Whether to include child domain units.

### `create_environment_profile` Block

* `domain_unit_id` - (Optional) Identifier of the domain unit.

### `create_form_type` Block

* `include_child_domain_units` - (Optional) Whether to include child domain units.

### `create_glossary` Block

* `include_child_domain_units` - (Optional) Whether to include child domain units.

### `create_project` Block

* `include_child_domain_units` - (Optional) Whether to include child domain units.

### `create_project_from_project_profile` Block

* `include_child_domain_units` - (Optional) Whether to include child domain units.
* `project_profiles` - (Optional) List of project profile identifiers.

### `override_domain_unit_owners` Block

* `include_child_domain_units` - (Optional) Whether to include child domain units.

### `override_project_owners` Block

* `include_child_domain_units` - (Optional) Whether to include child domain units.

### `use_asset_type` Block

* `domain_unit_id` - (Optional) Identifier of the domain unit.

### `principal` Block

Exactly one of the following sub-blocks must be specified:

* `domain_unit` - (Optional) Domain unit principal. See [`domain_unit` Block](#domain_unit-block) below.
* `group` - (Optional) Group principal. See [`group` Block](#group-block) below.
* `project` - (Optional) Project principal. See [`project` Block](#project-block) below.
* `user` - (Optional) User principal. See [`user` Block](#user-block) below.

### `domain_unit` Block

* `all_domain_units_grant_filter` - (Optional) Filter to grant access to all domain units. Empty block.
* `domain_unit_designation` - (Required) Designation of the domain unit principal. Valid values: `OWNER`.
* `domain_unit_identifier` - (Optional) Identifier of the domain unit.

### `group` Block

* `group_identifier` - (Required) Identifier of the group principal.

### `project` Block

* `domain_unit_filter` - (Optional) Filter for domain unit scoping. See [`domain_unit_filter` Block](#domain_unit_filter-block) below.
* `project_designation` - (Required) Designation of the project principal. Valid values: `CONTRIBUTOR`, `OWNER`, `PROJECT_CATALOG_STEWARD`.
* `project_identifier` - (Optional) Identifier of the project.

### `domain_unit_filter` Block

* `domain_unit` - (Required) Identifier of the domain unit for filtering.
* `include_child_domain_units` - (Optional) Whether to include child domain units in the filter.

### `user` Block

* `all_users_grant_filter` - (Optional) Filter to grant access to all users. Empty block.
* `user_identifier` - (Optional) Identifier of the user principal.

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
