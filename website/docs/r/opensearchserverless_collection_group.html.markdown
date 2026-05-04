---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_collection_group"
description: |-
  Terraform resource for managing an AWS OpenSearch Serverless Collection Group.
---

# Resource: aws_opensearchserverless_collection_group

Terraform resource for managing an AWS OpenSearch Serverless Collection Group.

Collection groups let multiple OpenSearch Serverless collections share compute resources while keeping encryption and access controls independent.

## Example Usage

### Basic Usage

```terraform
resource "aws_opensearchserverless_collection_group" "example" {
  name             = "example-group"
  description      = "Shared compute for production collections"
  standby_replicas = "ENABLED"

  capacity_limits {
    min_indexing_capacity_in_ocu = 2
    max_indexing_capacity_in_ocu = 16
    min_search_capacity_in_ocu   = 2
    max_search_capacity_in_ocu   = 16
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) Name of the collection group.
* `standby_replicas` - (Required, Forces new resource) Indicates whether standby replicas should be used for collections in this group. Valid values are `ENABLED` and `DISABLED`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `capacity_limits` - (Optional) Configuration block for the collection group's indexing and search capacity limits. See [`capacity_limits`](#capacity_limits) below for details.
* `description` - (Optional) Description of the collection group.
* `tags` - (Optional) A map of tags to assign to the collection group. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `capacity_limits`

* `max_indexing_capacity_in_ocu` - (Optional) Maximum indexing capacity, in OpenSearch Compute Units (OCUs), for the collection group.
* `max_search_capacity_in_ocu` - (Optional) Maximum search capacity, in OpenSearch Compute Units (OCUs), for the collection group.
* `min_indexing_capacity_in_ocu` - (Optional) Minimum indexing capacity, in OpenSearch Compute Units (OCUs), for the collection group.
* `min_search_capacity_in_ocu` - (Optional) Minimum search capacity, in OpenSearch Compute Units (OCUs), for the collection group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the collection group.
* `created_date` - Date the collection group was created.
* `id` - Unique identifier for the collection group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_opensearchserverless_collection_group.example
  identity = {
    id = "example-group-id"
  }
}

resource "aws_opensearchserverless_collection_group" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) Unique identifier for the collection group.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch Serverless Collection Group using the `id`. For example:

```terraform
import {
  to = aws_opensearchserverless_collection_group.example
  id = "example-group-id"
}
```

Using `terraform import`, import OpenSearch Serverless Collection Group using the `id`. For example:

```console
% terraform import aws_opensearchserverless_collection_group.example example-group-id
```
