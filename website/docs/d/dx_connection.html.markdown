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

* `name` - (Required) Name of the connection to retrieve.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the connection.
* `aws_device` - Direct Connect endpoint on which the physical connection terminates.
* `bandwidth` - Bandwidth of the connection.
* `id` - ID of the connection.
* `location` - AWS Direct Connect location where the connection is located.
* `owner_account_id` - ID of the AWS account that owns the connection.
* `provider_name` - Name of the service provider associated with the connection.
* `tags` - Map of tags for the resource.
* `vlan_id` - The VLAN ID.
