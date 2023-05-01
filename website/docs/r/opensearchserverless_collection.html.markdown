---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_collection"
description: |-
  Terraform resource for managing an AWS OpenSearch Collection.
---

# Resource: aws_opensearchserverless_collection

Terraform resource for managing an AWS OpenSearch Serverless Collection.

## Example Usage

### Basic Usage

```terraform
resource "aws_opensearchserverless_collection" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the collection.

The following arguments are optional:

* `description` - (Optional) Description of the collection.
* `tags` - (Optional) A map of tags to assign to the collection. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) Type of collection. One of `SEARCH` or `TIMESERIES`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the collection.
* `id` - Unique identifier for the collection.

## Import

OpenSearchServerless Collection can be imported using the `id`, e.g.,

```
$ terraform import aws_opensearchserverless_collection.example example
```
