---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_domain_policy"
description: |-
  Provides an OpenSearch Domain Policy.
---

# Resource: aws_opensearch_domain_policy

Allows setting policy to an OpenSearch domain while referencing domain attributes (e.g., ARN).

## Example Usage

```terraform
resource "aws_opensearch_domain" "example" {
  domain_name    = "tf-test"
  engine_version = "OpenSearch_1.1"
}

data "aws_iam_policy_document" "main" {
  effect = "Allow"

  principals {
    type        = "*"
    identifiers = ["*"]
  }

  actions   = ["es:*"]
  resources = ["${aws_opensearch_domain.example.arn}/*"]

  condition {
    test     = "IpAddress"
    variable = "aws:SourceIp"
    values   = "127.0.0.1/32"
  }
}

resource "aws_opensearch_domain_policy" "main" {
  domain_name     = aws_opensearch_domain.example.domain_name
  access_policies = data.aws_iam_policy_document.main.json
}
```

## Argument Reference

This resource supports the following arguments:

* `access_policies` - (Optional) IAM policy document specifying the access policies for the domain
* `domain_name` - (Required) Name of the domain.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `update` - (Default `180m`)
* `delete` - (Default `90m`)
