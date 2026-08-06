---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_connections"
description: |-
  Retrieve information about all the AWS Direct Connect connections in the current AWS Region.
---

# Data Source: aws_dx_connections

Retrieve information about all the AWS Direct Connect connections in the current AWS Region.

~> **Note:** This data source is different from the [`aws_dx_connection`](/docs/providers/aws/d/dx_connection.html) data source which retrieves information about a specific AWS Direct Connect connection in the current AWS Region.

~> **Note:** Connections in the `deleted` and `rejected` states are included in `connections`. Filter on `state` before acting on a connection — see [Filtering by State](#filtering-by-state).

## Example Usage

```terraform
data "aws_dx_connections" "all" {}
```

### Filtering by State

Filtering is performed in your configuration, since the AWS API does not support filtering connections:

```terraform
data "aws_dx_connections" "all" {}

resource "aws_cloudwatch_metric_alarm" "connection_state" {
  for_each = toset([
    for connection in data.aws_dx_connections.all.connections :
    connection.id if connection.state == "available"
  ])

  alarm_name          = "dx-connection-state-${each.value}"
  namespace           = "AWS/DX"
  metric_name         = "ConnectionState"
  statistic           = "Minimum"
  comparison_operator = "LessThanThreshold"
  threshold           = 1
  period              = 300
  evaluation_periods  = 1

  dimensions = {
    ConnectionId = each.value
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `connections` - List of objects describing the Direct Connect connections. See [`connections`](#connections).

### connections

* `arn` - ARN of the connection.
* `aws_device` - Direct Connect endpoint on which the physical connection terminates.
* `bandwidth` - Bandwidth of the connection, such as `1Gbps`.
* `id` - ID of the connection, such as `dxcon-ffre0ec3`.
* `location` - AWS Direct Connect location where the connection is located.
* `name` - Name of the connection.
* `owner_account_id` - ID of the AWS account that owns the connection.
* `partner_name` - Name of the AWS Direct Connect service provider associated with the connection.
* `provider_name` - Name of the service provider associated with the connection.
* `state` - State of the connection.
* `tags` - Map of tags assigned to the connection.
* `vlan_id` - VLAN ID.
