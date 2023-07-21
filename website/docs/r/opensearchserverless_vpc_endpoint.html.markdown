---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_vpc_endpoint"
description: |-
  Terraform resource for managing an AWS OpenSearch Serverless VPC Endpoint.
---

# Resource: aws_opensearchserverless_vpc_endpoint

Terraform resource for managing an AWS OpenSearchServerless VPC Endpoint.

## Example Usage

### Basic Usage

```terraform
resource "aws_opensearchserverless_vpc_endpoint" "example" {
  name       = "myendpoint"
  subnet_ids = [aws_subnet.example.id]
  vpc_id     = aws_vpc.example.id
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the interface endpoint.
* `subnet_ids` - (Required) One or more subnet IDs from which you'll access OpenSearch Serverless. Up to 6 subnets can be provided.
* `vpc_id` - (Required) ID of the VPC from which you'll access OpenSearch Serverless.

The following arguments are optional:

* `security_group_ids` - (Optional) One or more security groups that define the ports, protocols, and sources for inbound traffic that you are authorizing into your endpoint. Up to 5 security groups can be provided.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identified of the Vpc Endpoint.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

OpenSearchServerless Vpc Endpointa can be imported using the `id`, e.g.,

```
$ terraform import aws_opensearchserverless_vpc_endpoint.example vpce-8012925589
```
