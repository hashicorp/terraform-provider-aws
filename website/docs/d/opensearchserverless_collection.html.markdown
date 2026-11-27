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

The following arguments are optional:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `id` - (Optional) ID of the collection.
* `name` - (Optional) Name of the collection.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

~> Exactly one of `id` or `name` is required.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the collection.
* `collection_endpoint` - Collection-specific endpoint used to submit index, search, and data upload requests to an OpenSearch Serverless collection.
* `created_date` - Date the Collection was created.
* `dashboard_endpoint` - Collection-specific endpoint used to access OpenSearch Dashboards.
* `description` - Description of the collection.
* `failure_code` - Failure code associated with the collection.
* `failure_message` - Failure message associated with the collection.
* `kms_key_arn` - ARN of the Amazon Web Services KMS key used to encrypt the collection.
* `last_modified_date` - Date the Collection was last modified.
* `standby_replicas` - Whether standby replicas should be used for a collection.
* `tags` - Map of tags assigned to the collection.
* `type` - Type of collection.
