---
layout: "aws"
page_title: "AWS: aws_glue_catalog_database"
sidebar_current: "docs-aws-resource-glue-catalog-database"
description: |-
  Provides a Glue Catalog Database.
---

# aws_glue_catalog_database

Provides a Glue Catalog Database Resource. You can refer to the [Glue Developer Guide](http://docs.aws.amazon.com/glue/latest/dg/populate-data-catalog.html) for a full explanation of the Glue Data Catalog functionality

## Example Usage

```hcl-terraform
resource "aws_glue_catalog_database" "aws_glue_catalog_database" {
  name = "MyCatalogDatabase"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the database.
* `description` - (Optional) Description of the database.
* `location_uri` - (Optional) The location of the database (for example, an HDFS path).
* `parameters` - (Optional) A list of key-value pairs that define parameters and properties of the database.

## Attributes Reference

The following attributes are exported:

* `create_time` - The time at which the metadata database was created in the catalog.

## Import

Glue Catalog Databasess can be imported using the `name`, e.g.

```
$ terraform import aws_glue_catalog_database.database my_database
```
