---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_collection"
description: |-
  Terraform data source for managing an AWS OpenSearch Serverless Collection.
---

# Data Source: aws_opensearchserverless_collection

Terraform data source for managing an AWS OpenSearch Serverless Collection.

## Example Usage

### Basic Usage

```terraform
data "aws_opensearchserverless_collection" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) ID of the collection. Either `id` or `name` must be provided.
* `name` - (Required) Name of the collection. Either `name` or `id` must be provided.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the collection.
* `collection_endpoint` - Collection-specific endpoint used to submit index, search, and data upload requests to an OpenSearch Serverless collection.
* `created_date` - Date the Collection was created.
* `dashboard_endpont` - Collection-specific endpoint used to access OpenSearch Dashboards.
* `description` - Description of the collection.
* `kms_key_arn` - The ARN of the Amazon Web Services KMS key used to encrypt the collection.
* `last_modified_date` - Date the Collection was last modified.
* `tags` - A map of tags to assign to the collection.
* `type` - Type of collection.
