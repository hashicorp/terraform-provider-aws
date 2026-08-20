---
subcategory: "Lambda Web"
layout: "aws"
page_title: "AWS: aws_lambdaweb_resource_policy"
description: |-
  Manages a resource policy for an AWS Lambda Web function.
---

# Resource: aws_lambdaweb_resource_policy

Manages a resource-based policy for an AWS Lambda Web function. A resource policy attaches an IAM policy document directly to the function ARN, letting you grant or restrict who may reach the function's endpoints. A common use is locking a function down so that only a specific CloudFront distribution (or another trusted principal) can invoke it, keeping the raw endpoint domain from being reachable directly.

~> **Note:** Lambda Web is available in select regions. Regions active as of July 2026: `us-east-1`, `eu-west-1`. The `us-west-2` rollout is pending.

~> **Note:** A resource policy attaches to the parent web function only. Endpoint and revision ARNs (`.../endpoint/<name>`, `.../revision/<id>`) are not accepted.

## Example Usage

### Basic Usage

```terraform
resource "aws_lambdaweb_resource_policy" "example" {
  resource_arn = aws_lambdaweb_function.example.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowCloudFront"
        Effect    = "Allow"
        Principal = { Service = "cloudfront.amazonaws.com" }
        Action    = "lambda:InvokeWebFunction"
        Resource  = aws_lambdaweb_function.example.arn
        Condition = {
          StringEquals = {
            "aws:SourceArn" = aws_cloudfront_distribution.example.arn
          }
        }
      }
    ]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required) ARN of the Lambda Web function the policy is attached to. Must be a web function ARN (`arn:aws:lambda:<region>:<account>:web-function/<name>`). Changing this forces a new resource.
* `policy` - (Required) Resource-based IAM policy document as JSON. Terraform normalizes the document, so equivalent policies that differ only in key order or whitespace do not produce spurious diffs.
* `region` - (Optional) Region where this resource is managed. Defaults to the provider's configured region.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `revision_id` - Opaque revision identifier for the policy. The service returns a new value on every write; it is used for optimistic concurrency on update and delete.

## Import

In Terraform v1.12.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Lambda Web resource policy using the function `resource_arn`. For example:

```terraform
import {
  to = aws_lambdaweb_resource_policy.example
  id = "arn:aws:lambda:us-east-1:123456789012:web-function/example"
}
```

Using `terraform import`, import a Lambda Web resource policy using the function `resource_arn`. For example:

```console
% terraform import aws_lambdaweb_resource_policy.example arn:aws:lambda:us-east-1:123456789012:web-function/example
```
