---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_gateway_cluster_association"
description: |-
  Associates an Amazon ECS Gateway with an ECS cluster.
---

# Resource: aws_ecs_gateway_cluster_association

Associates an [`aws_ecs_gateway`](ecs_gateway.html) with an ECS cluster. Services in the cluster can only use a gateway (through a `gateway_connector` block on [`aws_ecs_service`](ecs_service.html)) after the gateway is associated with the cluster.

## Example Usage

```terraform
resource "aws_ecs_gateway_cluster_association" "example" {
  cluster = aws_ecs_cluster.example.name
  gateway = aws_ecs_gateway.example.arn
}
```

## Argument Reference

The following arguments are required:

* `cluster` - (Required, Forces new resource) Name or ARN of the ECS cluster.
* `gateway` - (Required, Forces new resource) Name or ARN of the ECS gateway.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `cluster_arn` - ARN of the ECS cluster.
* `gateway_arn` - ARN of the ECS gateway.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ecs_gateway_cluster_association.example
  identity = {
    cluster = "example-cluster"
    gateway = "arn:aws:ecs:us-east-1:123456789012:gateway/example/0123456789abcdef"
  }
}

resource "aws_ecs_gateway_cluster_association" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `cluster` (String) Name or ARN of the ECS cluster.
* `gateway` (String) Name or ARN of the ECS gateway.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS Gateway Cluster Associations using `cluster` and `gateway` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ecs_gateway_cluster_association.example
  id = "example-cluster,arn:aws:ecs:us-east-1:123456789012:gateway/example/0123456789abcdef"
}
```

Using `terraform import`, import ECS Gateway Cluster Associations using `cluster` and `gateway` separated by a comma (`,`). For example:

```console
% terraform import aws_ecs_gateway_cluster_association.example example-cluster,arn:aws:ecs:us-east-1:123456789012:gateway/example/0123456789abcdef
```
