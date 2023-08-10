---
subcategory: "CloudSearch"
layout: "aws"
page_title: "AWS: aws_cloudsearch_domain_service_access_policy"
description: |-
  Provides an CloudSearch domain service access policy resource. 
---

# Resource: aws_cloudsearch_domain_service_access_policy

Provides an CloudSearch domain service access policy resource.

Terraform waits for the domain service access policy to become `Active` when applying a configuration.

## Example Usage

```terraform
resource "aws_cloudsearch_domain" "example" {
  name = "example-domain"
}

data "aws_iam_policy_document" "example" {
  statement {
    sid    = "search_only"
    effect = "Allow"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = [
      "cloudsearch:search",
      "cloudsearch:document",
    ]

    condition {
      test     = "IpAddress"
      variable = "aws:SourceIp"
      values   = ["192.0.2.0/32"]
    }
  }
}



resource "aws_cloudsearch_domain_service_access_policy" "example" {
  domain_name   = aws_cloudsearch_domain.example.id
  access_policy = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `access_policy` - (Required) The access rules you want to configure. These rules replace any existing rules. See the [AWS documentation](https://docs.aws.amazon.com/cloudsearch/latest/developerguide/configuring-access.html) for details.
* `domain_name` - (Required) The CloudSearch domain name the policy applies to.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `update` - (Default `20m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudSearch domain service access policies using the domain name. For example:

```terraform
import {
  to = aws_cloudsearch_domain_service_access_policy.example
  id = "example-domain"
}
```

Using `terraform import`, import CloudSearch domain service access policies using the domain name. For example:

```console
% terraform import aws_cloudsearch_domain_service_access_policy.example example-domain
```
