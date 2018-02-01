---
layout: "aws"
page_title: "AWS: aws_dx_connection"
sidebar_current: "docs-aws-resource-dx-connection"
description: |-
  Provides a Connection of Direct Connect.
---

# aws_dx_connection

Provides a Connection of Direct Connect.

## Example Usage

```hcl
resource "aws_dx_connection" "hoge" {
  name = "tf-dx-connection"
  bandwidth = "1Gbps"
  location = "EqDC2"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the connection.
* `bandwidth` - (Required) The bandwidth of the connection. Available values: 1Gbps, 10Gbps. Case sensitive.
* `location` - (Required) The AWS Direct Connect location where the connection is located. See [DescribeLocations](https://docs.aws.amazon.com/directconnect/latest/APIReference/API_DescribeLocations.html) for the list of AWS Direct Connect locations. Use `locationCode`.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the connection.
