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

* `id` - (Required) A concatenation of the catalog ID and connection name. For example, if your account ID is 
`123456789123` and the connection name is `conn` then the ID is `123456789123:conn`.

## Attributes Reference

* `catalog_id` - The catalog ID of the Glue Connection.

* `name` - The name of the Glue Connection.

* `creation_time` - The time of creation of the Glue Connection.

* `connection_type` - The type of Glue Connection.
