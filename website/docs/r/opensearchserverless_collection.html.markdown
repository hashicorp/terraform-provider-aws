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
resource "aws_opensearchserverless_security_policy" "example" {
  name = "example"
  type = "encryption"
  policy = jsonencode({
    "Rules" = [
      {
        "Resource" = [
          "collection/example"
        ],
        "ResourceType" = "collection"
      }
    ],
    "AWSOwnedKey" = true
  })
}

resource "aws_opensearchserverless_collection" "example" {
  name = "example"

  depends_on = [aws_opensearchserverless_security_policy.example]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the collection.

The following arguments are optional:

* `description` - (Optional) Description of the collection.
* `tags` - (Optional) A map of tags to assign to the collection. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) Type of collection. One of `SEARCH` or `TIMESERIES`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the collection.
* `collection_endpoint` - Collection-specific endpoint used to submit index, search, and data upload requests to an OpenSearch Serverless collection.
* `dashboard_endpont` - Collection-specific endpoint used to access OpenSearch Dashboards.
* `kms_key_arn` - The ARN of the Amazon Web Services KMS key used to encrypt the collection.
* `id` - Unique identifier for the collection.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `20m`)
- `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearchServerless Collection using the `id`. For example:

```terraform
import {
  to = aws_opensearchserverless_collection.example
  id = "example"
}
```

Using `terraform import`, import OpenSearchServerless Collection using the `id`. For example:

```console
% terraform import aws_opensearchserverless_collection.example example
```
