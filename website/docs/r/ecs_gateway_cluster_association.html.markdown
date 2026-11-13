---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_gateway_cluster_association"
description: |-
  Manages an association between an ECS Gateway and an ECS cluster.
---

# Resource: aws_ecs_gateway_cluster_association

Manages an association between an ECS Gateway and an ECS cluster. Services in an associated cluster can route traffic through the gateway. A cluster can be associated with more than one gateway.

~> **Note:** While services in the cluster route traffic through the gateway, AWS keeps a disassociated association in the `DISASSOCIATING` state. Destroying this resource waits until those services stop using the gateway, up to the `delete` timeout.

## Example Usage

### Basic Usage

```terraform
resource "aws_ecs_gateway_cluster_association" "example" {
  gateway_arn = aws_ecs_gateway.example.arn
  cluster_arn = aws_ecs_cluster.example.arn
}
```

## Argument Reference

The following arguments are required:

* `cluster_arn` - (Required, Forces new resource) ARN of the ECS cluster.
* `gateway_arn` - (Required, Forces new resource) ARN of the ECS Gateway.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `delete` - (Default `20m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ecs_gateway_cluster_association.example
  identity = {
    gateway_arn = "arn:aws:ecs:us-west-2:123456789012:gateway/example/a1b2c3d4-5678-90ab-cdef-11111example"
    cluster_arn = "arn:aws:ecs:us-west-2:123456789012:cluster/example"
  }
}

resource "aws_ecs_gateway_cluster_association" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `cluster_arn` (String) ARN of the ECS cluster.
* `gateway_arn` (String) ARN of the ECS Gateway.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS Gateway Cluster Associations using `gateway_arn` and `cluster_arn` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ecs_gateway_cluster_association.example
  id = "arn:aws:ecs:us-west-2:123456789012:gateway/example/a1b2c3d4-5678-90ab-cdef-11111example,arn:aws:ecs:us-west-2:123456789012:cluster/example"
}
```

Using `terraform import`, import ECS Gateway Cluster Associations using `gateway_arn` and `cluster_arn` separated by a comma (`,`). For example:

```console
% terraform import aws_ecs_gateway_cluster_association.example arn:aws:ecs:us-west-2:123456789012:gateway/example/a1b2c3d4-5678-90ab-cdef-11111example,arn:aws:ecs:us-west-2:123456789012:cluster/example
```
