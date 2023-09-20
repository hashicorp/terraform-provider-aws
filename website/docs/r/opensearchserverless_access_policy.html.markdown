---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_access_policy"
description: |-
  Terraform resource for managing an AWS OpenSearch Serverless Access Policy.
---

# Resource: aws_opensearchserverless_access_policy

Terraform resource for managing an AWS OpenSearch Serverless Access Policy. See AWS documentation for [data access policies](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/serverless-data-access.html) and [supported data access policy permissions](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/serverless-data-access.html#serverless-data-supported-permissions).

## Example Usage

### Grant all collection and index permissions

```terraform
data "aws_caller_identity" "current" {}

resource "aws_opensearchserverless_access_policy" "example" {
  name        = "example"
  type        = "data"
  description = "read and write permissions"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource = [
            "index/example-collection/*"
          ],
          Permission = [
            "aoss:*"
          ]
        },
        {
          ResourceType = "collection",
          Resource = [
            "collection/example-collection"
          ],
          Permission = [
            "aoss:*"
          ]
        }
      ],
      Principal = [
        data.aws_caller_identity.current.arn
      ]
    }
  ])
}
```

### Grant read-only collection and index permissions

```
data "aws_caller_identity" "current" {}

resource "aws_opensearchserverless_access_policy" "example" {
  name        = "example"
  type        = "data"
  description = "read-only permissions"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource = [
            "index/example-collection/*"
          ],
          Permission = [
            "aoss:DescribeIndex",
            "aoss:ReadDocument",
          ]
        },
        {
          ResourceType = "collection",
          Resource = [
            "collection/example-collection"
          ],
          Permission = [
            "aoss:DescribeCollectionItems"
          ]
        }
      ],
      Principal = [
        data.aws_caller_identity.current.arn
      ]
    }
  ])
}
```

### Grant SAML identity permissions

```
resource "aws_opensearchserverless_access_policy" "example" {
  name = "example"
  type = "data"
  description = "saml permissions"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource = [
            "index/example-collection/*"
          ],
          Permission = [
            "aoss:*"
          ]
        },
        {
          ResourceType = "collection",
          Resource = [
            "collection/example-collection"
          ],
          Permission = [
            "aoss:*"
          ]
        }
      ],
      Principal = [
        "saml/123456789012/myprovider/user/Annie",
        "saml/123456789012/anotherprovider/group/Accounting"
      ]
    }
  ])
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the policy.
* `policy` - (Required) JSON policy document to use as the content for the new policy
* `type` - (Required) Type of access policy. Must be `data`.

The following arguments are optional:

* `description` - (Optional) Description of the policy. Typically used to store information about the permissions defined in the policy.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `policy_version` - Version of the policy.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearchServerless Access Policy using the `name` and `type` arguments separated by a slash (`/`). For example:

```terraform
import {
  to = aws_opensearchserverless_access_policy.example
  id = "example/data"
}
```

Using `terraform import`, import OpenSearchServerless Access Policy using the `name` and `type` arguments separated by a slash (`/`). For example:

```console
% terraform import aws_opensearchserverless_access_policy.example example/data
```
