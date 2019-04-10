---
layout: "aws"
page_title: "AWS: aws_dx_connection"
sidebar_current: "docs-aws-datasource-dx-connection"
description: |-
  Retrieve information about a Direct Connect Connection
---

# Data Source: aws_dx_connection

Retrieve information about a Direct Connect Connection.

## Example Usage

```hcl
data "aws_dx_connection" "example" {
  name = "example"
}
```

## Argument Reference

* `name` - (Required) The name of the connection to retrieve.

## Attributes Reference

* `id` - The ID of the connection.
* `state` - The state of the connection.
* `location` - The AWS Direct Connect location where the connection is location.
* `bandwidth` - The Bandwidth of the connection.
* `jumbo_frame_capable` - Boolean value representing if jumbo frames have been enabled for this connection
