---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_database"
description: |-
  Manages a Lightsail managed database instance.
---

# Resource: aws_lightsail_database

Manages a Lightsail database. Use this resource to create and manage fully managed database instances with automated backups, monitoring, and maintenance in Lightsail.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones"](https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/) for more details

## Example Usage

### Basic MySQL Blueprint

```terraform
resource "aws_lightsail_database" "example" {
  relational_database_name = "example-database"
  availability_zone        = "us-east-1a"
  master_database_name     = "exampledb"
  master_password          = "examplepassword123"
  master_username          = "exampleuser"
  blueprint_id             = "mysql_8_0"
  bundle_id                = "micro_1_0"
}
```

### Basic PostgreSQL Blueprint

```terraform
resource "aws_lightsail_database" "example" {
  relational_database_name = "example-database"
  availability_zone        = "us-east-1a"
  master_database_name     = "exampledb"
  master_password          = "examplepassword123"
  master_username          = "exampleuser"
  blueprint_id             = "postgres_12"
  bundle_id                = "micro_1_0"
}
```

### Custom Backup and Maintenance Windows

Below is an example that sets a custom backup and maintenance window. Times are specified in UTC. This example will allow daily backups to take place between 16:00 and 16:30 each day. This example also requires any maintenance tasks (anything that would cause an outage, including changing some attributes) to take place on Tuesdays between 17:00 and 17:30. An action taken against this database that would cause an outage will wait until this time window to make the requested changes.

```terraform
resource "aws_lightsail_database" "example" {
  relational_database_name     = "example-database"
  availability_zone            = "us-east-1a"
  master_database_name         = "exampledb"
  master_password              = "examplepassword123"
  master_username              = "exampleuser"
  blueprint_id                 = "postgres_12"
  bundle_id                    = "micro_1_0"
  preferred_backup_window      = "16:00-16:30"
  preferred_maintenance_window = "Tue:17:00-Tue:17:30"
}
```

### Final Snapshots

To enable creating a final snapshot of your database on deletion, use the `final_snapshot_name` argument to provide a name to be used for the snapshot.

```terraform
resource "aws_lightsail_database" "example" {
  relational_database_name     = "example-database"
  availability_zone            = "us-east-1a"
  master_database_name         = "exampledb"
  master_password              = "examplepassword123"
  master_username              = "exampleuser"
  blueprint_id                 = "postgres_12"
  bundle_id                    = "micro_1_0"
  preferred_backup_window      = "16:00-16:30"
  preferred_maintenance_window = "Tue:17:00-Tue:17:30"
  final_snapshot_name          = "example-final-snapshot"
}
```

### Apply Immediately

To enable applying changes immediately instead of waiting for a maintenance window, use the `apply_immediately` argument.

```terraform
resource "aws_lightsail_database" "example" {
  relational_database_name = "example-database"
  availability_zone        = "us-east-1a"
  master_database_name     = "exampledb"
  master_password          = "examplepassword123"
  master_username          = "exampleuser"
  blueprint_id             = "postgres_12"
  bundle_id                = "micro_1_0"
  apply_immediately        = true
}
```

## Argument Reference

The following arguments are required:

* `blueprint_id` - (Required) Blueprint ID for your database. A blueprint describes the major engine version of a database. You can get a list of database blueprints IDs by using the AWS CLI command: `aws lightsail get-relational-database-blueprints`
* `bundle_id` - (Required) Bundle ID for your database. A bundle describes the performance specifications for your database (see list below). You can get a list of database bundle IDs by using the AWS CLI command: `aws lightsail get-relational-database-bundles`.
* `master_database_name` - (Required) Name of the master database created when the Lightsail database resource is created.
* `master_password` - (Required, Sensitive) Password for the master user of your database. The password can include any printable ASCII character except "/", """, or "@".
* `master_username` - (Required) Master user name for your database.
* `relational_database_name` - (Required) Name to use for your Lightsail database resource. Names be unique within each AWS Region in your Lightsail account.

The following arguments are optional:

* `apply_immediately` - (Optional) Whether to apply changes immediately. When false, applies changes during the preferred maintenance window. Some changes may cause an outage.
* `availability_zone` - (Optional) Availability Zone in which to create your database. Use the us-east-2a case-sensitive format.
* `backup_retention_enabled` - (Optional) Whether to enable automated backup retention for your database. When false, disables automated backup retention for your database. Disabling backup retention deletes all automated database backups. Before disabling this, you may want to create a snapshot of your database.
* `final_snapshot_name` - (Required unless `skip_final_snapshot = true`) Name of the database snapshot created if skip final snapshot is false, which is the default value for that parameter.
* `preferred_backup_window` - (Optional) Daily time range during which automated backups are created for your database if automated backups are enabled. Must be in the hh24:mi-hh24:mi format. Example: `16:00-16:30`. Specified in Coordinated Universal Time (UTC).
* `preferred_maintenance_window` - (Optional) Weekly time range during which system maintenance can occur on your database. Must be in the ddd:hh24:mi-ddd:hh24:mi format. Specified in Coordinated Universal Time (UTC). Example: `Tue:17:00-Tue:17:30`
* `publicly_accessible` - (Optional) Whether the database is accessible to resources outside of your Lightsail account. A value of true specifies a database that is available to resources outside of your Lightsail account. A value of false specifies a database that is available only to your Lightsail resources in the same region as your database.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `skip_final_snapshot` - (Optional) Whether a final database snapshot is created before your database is deleted. If true is specified, no database snapshot is created. If false is specified, a database snapshot is created before your database is deleted. You must specify the final relational database snapshot name parameter if the skip final snapshot parameter is false.
* `tags` - (Optional) Map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Blueprint IDs

A list of all available Lightsail Blueprints for Relational Databases the [aws lightsail get-relational-database-blueprints](https://docs.aws.amazon.com/cli/latest/reference/lightsail/get-relational-database-blueprints.html) aws cli command.

### Examples

- `mysql_8_0`
- `postgres_12`

### Prefix

A Blueprint ID starts with a prefix of the engine type.

### Suffix

A Blueprint ID has a suffix of the engine version.

## Bundles

A list of all available Lightsail Bundles for Relational Databases the [aws lightsail get-relational-database-bundles](https://docs.aws.amazon.com/cli/latest/reference/lightsail/get-relational-database-bundles.html) aws cli command.

### Examples

- `small_1_0`
- `small_ha_1_0`
- `large_1_0`
- `large_ha_1_0`

### Prefix

A Bundle ID starts with one of the below size prefixes:

- `micro_`
- `small_`
- `medium_`
- `large_`

### Infixes (Optional for HA Database)

A Bundle ID can have the following infix added in order to use the HA option of the selected bundle.

- `ha_`

### Suffix

A Bundle ID ends with one of the following suffix: `1_0`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the database (matches `id`).
* `ca_certificate_identifier` - Certificate associated with the database.
* `cpu_count` - Number of vCPUs for the database.
* `created_at` - Date and time when the database was created.
* `disk_size` - Size of the disk for the database.
* `engine` - Database software (for example, MySQL).
* `engine_version` - Database engine version (for example, 5.7.23).
* `id` - ARN of the database (matches `arn`).
* `master_endpoint_address` - Master endpoint FQDN for the database.
* `master_endpoint_port` - Master endpoint network port for the database.
* `ram_size` - Amount of RAM in GB for the database.
* `secondary_availability_zone` - Secondary Availability Zone of a high availability database. The secondary database is used for failover support of a high availability database.
* `support_code` - Support code for the database. Include this code in your email to support when you have questions about a database in Lightsail. This code enables our support team to look up your Lightsail information more easily.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider `default_tags` configuration block.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lightsail Databases using their name. For example:

```terraform
import {
  to = aws_lightsail_database.example
  id = "example-database"
}
```

Using `terraform import`, import Lightsail Databases using their name. For example:

```console
% terraform import aws_lightsail_database.example example-database
```
