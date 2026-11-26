---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_security_policy"
description: |-
  Terraform resource for managing an AWS OpenSearch Serverless Security Policy.
---

# Resource: aws_opensearchserverless_security_policy

Terraform resource for managing an AWS OpenSearch Serverless Security Policy. See AWS documentation for [encryption policies](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/serverless-encryption.html#serverless-encryption-policies) and [network policies](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/serverless-network.html#serverless-network-policies).

## Example Usage

### Encryption Security Policy

#### Single Collection

```terraform
resource "aws_opensearchserverless_security_policy" "example" {
  name        = "example"
  type        = "encryption"
  description = "encryption security policy for example-collection"
  policy = jsonencode({
    Rules = [
      {
        Resource = [
          "collection/example-collection"
        ],
        ResourceType = "collection"
      }
    ],
    AWSOwnedKey = true
  })
}
```

#### Multiple Collections

```terraform
resource "aws_opensearchserverless_security_policy" "example" {
  name        = "example"
  type        = "encryption"
  description = "encryption security policy for collections that begin with \"example\""
  policy = jsonencode({
    Rules = [
      {
        Resource = [
          "collection/example*"
        ],
        ResourceType = "collection"
      }
    ],
    AWSOwnedKey = true
  })
}
```

#### Customer Managed Key

```terraform
resource "aws_opensearchserverless_security_policy" "example" {
  name        = "example"
  type        = "encryption"
  description = "encryption security policy using customer KMS key"
  policy = jsonencode({
    Rules = [
      {
        Resource = [
          "collection/customer-managed-key-collection"
        ],
        ResourceType = "collection"
      }
    ],
    AWSOwnedKey = false
    KmsARN      = "arn:aws:kms:us-east-1:123456789012:key/93fd6da4-a317-4c17-bfe9-382b5d988b36"
  })
}
```

### Network Security Policy

#### Public Access

```terraform
resource "aws_opensearchserverless_security_policy" "example" {
  name        = "example"
  type        = "network"
  description = "Public access"
  policy = jsonencode([
    {
      Description = "Public access to collection and Dashboards endpoint for example collection",
      Rules = [
        {
          ResourceType = "collection",
          Resource = [
            "collection/example-collection"
          ]
        },
        {
          ResourceType = "dashboard"
          Resource = [
            "collection/example-collection"
          ]
        }
      ],
      AllowFromPublic = true
    }
  ])
}
```

#### VPC Access

```terraform
resource "aws_opensearchserverless_security_policy" "example" {
  name        = "example"
  type        = "network"
  description = "VPC access"
  policy = jsonencode([
    {
      Description = "VPC access to collection and Dashboards endpoint for example collection",
      Rules = [
        {
          ResourceType = "collection",
          Resource = [
            "collection/example-collection"
          ]
        },
        {
          ResourceType = "dashboard"
          Resource = [
            "collection/example-collection"
          ]
        }
      ],
      AllowFromPublic = false,
      SourceVPCEs = [
        "vpce-050f79086ee71ac05"
      ]
    }
  ])
}
```

#### Mixed Access by Collection

```terraform
resource "aws_opensearchserverless_security_policy" "example" {
  name        = "example"
  type        = "network"
  description = "Mixed access for marketing and sales"
  policy = jsonencode([
    {
      "Description" : "Marketing access",
      "Rules" : [
        {
          "ResourceType" : "collection",
          "Resource" : [
            "collection/marketing*"
          ]
        },
        {
          "ResourceType" : "dashboard",
          "Resource" : [
            "collection/marketing*"
          ]
        }
      ],
      "AllowFromPublic" : false,
      "SourceVPCEs" : [
        "vpce-050f79086ee71ac05"
      ]
    },
    {
      "Description" : "Sales access",
      "Rules" : [
        {
          "ResourceType" : "collection",
          "Resource" : [
            "collection/finance"
          ]
        }
      ],
      "AllowFromPublic" : true
    }
  ])
}
```

#### VPC Collection with Public Dashboards

This example creates an OpenSearch Serverless VPC endpoint and references it in the policy, so the collection endpoint is reachable only from the VPC while the Dashboards endpoint remains public.

```terraform
resource "aws_opensearchserverless_vpc_endpoint" "example" {
  name               = "example"
  vpc_id             = aws_vpc.example.id
  subnet_ids         = [aws_subnet.example.id]
  security_group_ids = [aws_security_group.example.id]
}

resource "aws_opensearchserverless_security_policy" "example" {
  name        = "example"
  type        = "network"
  description = "VPC access for collection endpoint, public access for Dashboards"
  policy = jsonencode([
    {
      Description = "VPC access for collection endpoint",
      Rules = [
        {
          ResourceType = "collection",
          Resource = [
            "collection/example-collection"
          ]
        }
      ],
      AllowFromPublic = false,
      SourceVPCEs = [
        aws_opensearchserverless_vpc_endpoint.example.id
      ]
    },
    {
      Description = "Public access for Dashboards",
      Rules = [
        {
          ResourceType = "dashboard",
          Resource = [
            "collection/example-collection"
          ]
        }
      ],
      AllowFromPublic = true
    }
  ])
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the policy.
* `policy` - (Required) JSON policy document to use as the content for the new policy
* `type` - (Required) Type of security policy. One of `encryption` or `network`.

The following arguments are optional:

* `description` - (Optional) Description of the policy. Typically used to store information about the permissions defined in the policy.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `policy_version` - Version of the policy.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_opensearchserverless_security_policy.example
  identity = {
    name = "example"
    type = "encryption"
  }
}

resource "aws_opensearchserverless_security_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the policy.
* `type` (String) Type of security policy.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearchServerless Security Policy using the `name` and `type` arguments separated by a slash (`/`). For example:

```terraform
import {
  to = aws_opensearchserverless_security_policy.example
  id = "example/encryption"
}
```

Using `terraform import`, import OpenSearchServerless Security Policy using the `name` and `type` arguments separated by a slash (`/`). For example:

```console
% terraform import aws_opensearchserverless_security_policy.example example/encryption
```
