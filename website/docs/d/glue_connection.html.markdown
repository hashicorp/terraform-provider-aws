---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_connection"
description: |-
  Get information on an AWS Glue Connection
---

# Data Source: aws_glue_connection

This data source can be used to fetch information about a specific Glue Connection.

## Example Usage

```terraform
data "aws_glue_connection" "example" {
  id = "123456789123:connection"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `id` - (Required) Concatenation of the catalog ID and connection name. For example, if your account ID is
`123456789123` and the connection name is `conn` then the ID is `123456789123:conn`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Glue Connection.
* `catalog_id` - Catalog ID of the Glue Connection.
* `athena_properties` - A map of connection properties specific to the Athena compute environment.
* `connection_properties` - A map of connection properties.
* `connection_type` - Type of Glue Connection.
* `description` - Description of the connection.
* `match_criteria` - A list of criteria that can be used in selecting this connection.
* `name` - Name of the Glue Connection.
* `physical_connection_requirements` - A map of physical connection requirements, such as VPC and SecurityGroup.
* `tags` - Tags assigned to the resource
