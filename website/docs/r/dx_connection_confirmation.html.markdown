---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_connection_confirmation"
description: |-
  Provides a confirmation of the creation of the specified hosted connection on an interconnect.
---

# Resource: aws_dx_connection_confirmation

Provides a confirmation of the creation of the specified hosted connection on an interconnect.

## Example Usage

```terraform
resource "aws_dx_connection_confirmation" "confirmation" {
  connection_id = "dxcon-ffabc123"
}
```

## Argument Reference

This resource supports the following arguments:

* `connection_id` - (Required) The ID of the hosted connection.

### Removing `aws_dx_connection_confirmation` from your configuration

Removing an `aws_dx_connection_confirmation` resource from your configuration will remove it
from your statefile and management, **but will not destroy the Hosted Connection.**

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the connection.
