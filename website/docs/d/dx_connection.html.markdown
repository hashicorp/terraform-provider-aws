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

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the connection.
* `aws_device` - Direct Connect endpoint on which the physical connection terminates.
* `bandwidth` - Bandwidth of the connection.
* `id` - ID of the connection.
* `location` - AWS Direct Connect location where the connection is located.
* `owner_account_id` - ID of the AWS account that owns the connection.
* `partner_name` - The name of the AWS Direct Connect service provider associated with the connection.
* `provider_name` - Name of the service provider associated with the connection.
* `tags` - Map of tags for the resource.
* `vlan_id` - The VLAN ID.
