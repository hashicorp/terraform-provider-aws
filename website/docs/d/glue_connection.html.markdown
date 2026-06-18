---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_connection"
description: |-
  Provides details about an AWS Glue Connection.
---

# Data Source: aws_glue_connection

Provides details about an AWS Glue Connection.

## Example Usage

```terraform
data "aws_glue_connection" "example" {
  id = "123456789123:connection"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Concatenation of the catalog ID and connection name. For example, if your account ID is `123456789123` and the connection name is `conn` then the ID is `123456789123:conn`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Glue Connection.
* `athena_properties` - Map of connection properties specific to the Athena compute environment.
* `authentication_configuration` - Configuration block for authentication options.
* `catalog_id` - Catalog ID of the Glue Connection.
* `connection_properties` - Map of connection properties.
* `connection_type` - Type of Glue Connection.
* `description` - Description of the connection.
* `match_criteria` - List of criteria that can be used in selecting this connection.
* `name` - Name of the Glue Connection.
* `physical_connection_requirements` - Map of physical connection requirements, such as VPC and SecurityGroup.
* `tags` - Tags assigned to the resource.
