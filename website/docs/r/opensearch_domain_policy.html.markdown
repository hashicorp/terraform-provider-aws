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

resource "aws_opensearch_domain_policy" "main" {
  domain_name = aws_opensearch_domain.example.domain_name

  access_policies = <<POLICIES
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "es:*",
            "Principal": "*",
            "Effect": "Allow",
            "Condition": {
                "IpAddress": {"aws:SourceIp": "127.0.0.1/32"}
            },
            "Resource": "${aws_opensearch_domain.example.arn}/*"
        }
    ]
}
POLICIES
}
```

## Argument Reference

The following arguments are supported:

* `access_policies` - (Optional) IAM policy document specifying the access policies for the domain
* `domain_name` - (Required) Name of the domain.

## Attributes Reference

No additional attributes are exported.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `update` - (Default `180m`)
* `delete` - (Default `90m`)
