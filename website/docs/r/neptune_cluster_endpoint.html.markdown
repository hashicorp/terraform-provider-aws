---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_cluster_endpoint"
description: |-
  Provides an Neptune Cluster Endpoint Resource
---

# Resource: aws_neptune_cluster_endpoint

Provides an Neptune Cluster Endpoint Resource.

## Example Usage

```terraform
resource "aws_neptune_cluster_endpoint" "example" {
  cluster_identifier          = aws_neptune_cluster.test.cluster_identifier
  cluster_endpoint_identifier = "example"
  endpoint_type               = "READER"
}
```

## Argument Reference

This resource supports the following arguments:

* `cluster_identifier` - (Required, Forces new resources) The DB cluster identifier of the DB cluster associated with the endpoint.
* `cluster_endpoint_identifier` - (Required, Forces new resources) The identifier of the endpoint.
* `endpoint_type` - (Required) The type of the endpoint. One of: `READER`, `WRITER`, `ANY`.
* `excluded_members` - (Optional) List of DB instance identifiers that aren't part of the custom endpoint group. All other eligible instances are reachable through the custom endpoint. Only relevant if the list of static members is empty.
* `static_members` - (Optional) List of DB instance identifiers that are part of the custom endpoint group.
* `tags` - (Optional) A map of tags to assign to the Neptune cluster. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Neptune Cluster Endpoint Amazon Resource Name (ARN).
* `endpoint` - The DNS address of the endpoint.
* `id` - The Neptune Cluster Endpoint Identifier.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_neptune_cluster_endpoint` using the `cluster-identifier:endpoint-identfier`. For example:

```terraform
import {
  to = aws_neptune_cluster_endpoint.example
  id = "my-cluster:my-endpoint"
}
```

Using `terraform import`, import `aws_neptune_cluster_endpoint` using the `cluster-identifier:endpoint-identfier`. For example:

```console
% terraform import aws_neptune_cluster_endpoint.example my-cluster:my-endpoint
```
