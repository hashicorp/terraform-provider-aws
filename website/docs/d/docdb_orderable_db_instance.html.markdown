---
subcategory: "DocumentDB"
layout: "aws"
page_title: "AWS: aws_docdb_orderable_db_instance"
description: |-
  Information about DocumentDB orderable DB instances.
---

# Data Source: aws_docdb_orderable_db_instance

Information about DocumentDB orderable DB instances.

## Example Usage

```terraform
data "aws_docdb_orderable_db_instance" "test" {
  engine         = "docdb"
  engine_version = "3.6.0"
  license_model  = "na"

  preferred_instance_classes = ["db.r5.large", "db.r4.large", "db.t3.medium"]
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `engine` - (Optional) DB engine. Default: `docdb`
* `engine_version` - (Optional) Version of the DB engine.
* `instance_class` - (Optional) DB instance class. Examples of classes are `db.r5.12xlarge`, `db.r5.24xlarge`, `db.r5.2xlarge`, `db.r5.4xlarge`, `db.r5.large`, `db.r5.xlarge`, and `db.t3.medium`. (Conflicts with `preferred_instance_classes`.)
* `license_model` - (Optional) License model. Default: `na`
* `preferred_instance_classes` - (Optional) Ordered list of preferred DocumentDB DB instance classes. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned. (Conflicts with `instance_class`.)
* `vpc` - (Optional) Enable to show only VPC.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `availability_zones` - Availability zones where the instance is available.
