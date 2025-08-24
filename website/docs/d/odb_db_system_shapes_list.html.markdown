---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_db_system_shapes_list"
page_title: "AWS: aws_odb_db_system_shapes_list"
description: |-
  Terraform data source to retrieve available system shapes Oracle Database@AWS.
---

# Data Source: aws_odb_db_system_shapes_list

Terraform data source to retrieve available system shapes Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_db_system_shapes_list" "example" {

}
```

## Argument Reference

The following arguments are optional:

* `availability_zone_id` - (Optional) The physical ID of the AZ, for example, use1-az4. This ID persists across accounts.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `db_system_shapes` - IThe list of shapes and their properties. Information about a hardware system model (shape) that's available for an Exadata infrastructure. The shape determines resources, such as CPU cores, memory, and storage, to allocate to the Exadata infrastructure.
