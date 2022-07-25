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

`aws_opensearch_domain_policy` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `update` - (Optional, Default: `180m`) How long to wait for updates.
* `delete` - (Optional, Default: `90m`) How long to wait for deletion.
