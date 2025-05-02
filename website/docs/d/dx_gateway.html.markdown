---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_gateway"
description: |-
  Retrieve information about a Direct Connect Gateway
---

# Data Source: aws_dx_gateway

Retrieve information about a Direct Connect Gateway.

## Example Usage

```terraform
data "aws_dx_gateway" "example" {
  name = "example"
}
```

## Argument Reference

* `name` - (Required) Name of the gateway to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `amazon_side_asn` - ASN on the Amazon side of the connection.
* `arn` - ARN of the gateway.
* `id` - ID of the gateway.
* `owner_account_id` - AWS Account ID of the gateway.
