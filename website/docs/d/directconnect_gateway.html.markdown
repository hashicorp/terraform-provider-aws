---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_directconnect_gateway"
description: |-
  Retrieve information about a Direct Connect Gateway
---

# Data Source: aws_directconnect_gateway

Retrieve information about a Direct Connect Gateway.

## Example Usage

```terraform
data "aws_directconnect_gateway" "example" {
  name = "example"
}
```

## Argument Reference

* `name` - (Required) The name of the gateway to retrieve.

## Attributes Reference

* `amazon_side_asn` - The ASN on the Amazon side of the connection.
* `id` - The ID of the gateway.
* `owner_account_id` - AWS Account ID of the gateway.
