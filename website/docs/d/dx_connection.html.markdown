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
# Find a connection from name only
data "aws_dx_connection" "example" {
  name = "example"
}
# Find a connection from name and tags
data "aws_dx_connection" "example" {
  name = "example"
  tags = {
    Location = "PEH51"
    LagGroup = "LAG-001"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the connection to retrieve.
* `tags` - (Optional) A map of tags associated with the connection.

## Attributes Reference

* `id` - The ID of the connection.
* `arn` - The ARN of the connection.
* `state` - The state of the connection.
* `location` - The AWS Direct Connect location where the connection is location.
* `bandwidth` - The Bandwidth of the connection.
* `jumbo_frame_capable` - Boolean value representing if jumbo frames have been enabled for this connection
