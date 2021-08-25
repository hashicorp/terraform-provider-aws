---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_connection"
description: |-
  Retrieve information about a Direct Connect Connection.
---

# Data Source: aws_dx_connection

Retrieve information about a Direct Connect Connection.

## Example Usage

```hcl
data "aws_dx_connection" "example" {
  name = "tf-dx-connection"
}
```

## Argument Reference

* `name` - (Required) The name of the connection to retrieve.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the connection.
* `owner_account_id` - The ID of the AWS account that owns the connection.
