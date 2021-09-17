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

* `arn` - The ARN of the connection.
* `aws_device` - The Direct Connect endpoint on which the physical connection terminates.
* `bandwidth` - The bandwidth of the connection.
* `id` - The ID of the connection.
* `location` - The AWS Direct Connect location where the connection is located.
* `owner_account_id` - The ID of the AWS account that owns the connection.
* `provider_name` - The name of the service provider associated with the connection.
* `tags` - A map of tags for the resource.
