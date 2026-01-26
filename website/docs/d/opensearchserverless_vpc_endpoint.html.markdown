---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_vpc_endpoint"
description: |-
  Terraform data source for managing an AWS OpenSearch Serverless VPC Endpoint.
---

# Data Source: aws_opensearchserverless_vpc_endpoint

Terraform data source for managing an AWS OpenSearch Serverless VPC Endpoint.

## Example Usage

```terraform
data "aws_opensearchserverless_vpc_endpoint" "example" {
  vpc_endpoint_id = "vpce-829a4487959e2a839"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_endpoint_id` - (Required) The unique identifier of the endpoint.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `created_date` - The date the endpoint was created.
* `name` - The name of the endpoint.
* `security_group_ids` - The IDs of the security groups that define the ports, protocols, and sources for inbound traffic that you are authorizing into your endpoint.
* `subnet_ids` - The IDs of the subnets from which you access OpenSearch Serverless.
* `vpc_id` - The ID of the VPC from which you access OpenSearch Serverless.
