---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_vpc_endpoint"
description: |-
  Terraform resource for managing an AWS OpenSearch VPC Endpoint.
---

# Resource: aws_opensearch_vpc_endpoint

Manages an [AWS Opensearch VPC Endpoint](https://docs.aws.amazon.com/opensearch-service/latest/APIReference/API_CreateVpcEndpoint.html). Creates an Amazon OpenSearch Service-managed VPC endpoint.

## Example Usage

### Basic Usage

```terraform

resource "aws_opensearch_vpc_endpoint" "foo" {
  domain_arn = aws_opensearch_domain.domain_1.arn
  vpc_options {
    security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
    subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id]
  }
}

```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `domain_arn` - (Required, Forces new resource) ARN of the domain to create the endpoint for.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_options` - (Required) Options to specify the subnets and security groups for the endpoint.

### `vpc_options` Block

* `security_group_ids` - (Optional) List of security group IDs associated with the VPC endpoints for the domain. If you do not provide a security group ID, OpenSearch Service uses the default security group for the VPC.
* `subnet_ids` - (Required) List of subnet IDs associated with the VPC endpoints for the domain. If your domain uses multiple Availability Zones, you need to provide two subnet IDs, one per zone. Otherwise, provide only one.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `endpoint` - Connection endpoint ID for connecting to the domain.
* `id` - Unique identifier of the endpoint.

### `vpc_options` Block

* `availability_zones` - Availability Zones associated with the VPC endpoint.
* `vpc_id` - ID of the VPC associated with the endpoint.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `60m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch VPC endpoints using the `id`. For example:

```terraform
import {
  to = aws_opensearch_vpc_endpoint.example
  id = "endpoint-id"
}
```

Using `terraform import`, import OpenSearch VPC endpoints using the `id`. For example:

```console
% terraform import aws_opensearch_vpc_endpoint.example endpoint-id
```
