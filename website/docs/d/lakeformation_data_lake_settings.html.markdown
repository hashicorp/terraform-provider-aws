---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_data_lake_settings"
description: |-
    Get data lake administrators and default database and table permissions
---

# Data Source: aws_lakeformation_data_lake_settings

Get Lake Formation principals designated as data lake administrators and lists of principal permission entries for default create database and default create table permissions.

## Example Usage

```terraform
data "aws_lakeformation_data_lake_settings" "example" {
  catalog_id = "14916253649"
}
```

## Argument Reference

The following arguments are optional:

* `catalog_id` – (Optional) Identifier for the Data Catalog. By default, the account ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `admins` – List of ARNs of AWS Lake Formation principals (IAM users or roles).
* `read_only_admins` – List of ARNs of AWS Lake Formation principals (IAM users or roles) with only view access to the resources.
* `create_database_default_permissions` - Up to three configuration blocks of principal permissions for default create database permissions. Detailed below.
* `create_table_default_permissions` - Up to three configuration blocks of principal permissions for default create table permissions. Detailed below.
* `trusted_resource_owners` – List of the resource-owning account IDs that the caller's account can use to share their user access details (user ARNs).
* `allow_external_data_filtering` - Whether to allow Amazon EMR clusters to access data managed by Lake Formation.
* `external_data_filtering_allow_list` - A list of the account IDs of Amazon Web Services accounts with Amazon EMR clusters that are to perform data filtering.
* `authorized_session_tag_value_list` - Lake Formation relies on a privileged process secured by Amazon EMR or the third party integrator to tag the user's role while assuming it.
* `allow_full_table_external_data_access` - Whether to allow a third-party query engine to get data access credentials without session tags when a caller has full data access permissions.

### create_database_default_permissions

* `permissions` - List of permissions granted to the principal.
* `principal` - Principal who is granted permissions.

### create_table_default_permissions

* `permissions` - List of permissions granted to the principal.
* `principal` - Principal who is granted permissions.
