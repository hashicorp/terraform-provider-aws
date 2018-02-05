---
layout: "aws"
page_title: "AWS: aws_lb_ssl_policy"
sidebar_current: "docs-aws-datasource-lb-ssl-policy"
description: |-
  Provides a Load Balancer SSL Policy data source.
---

# Data Source: aws_lb_ssl_policy

Provides information about a Load Balancer SSL Policy.

~> **Note:** When you don't specify `name`, the default policy will be set as `name`.

## Example Usage

```hcl
data "aws_lb_ssl_policy" "default" {}

data "aws_lb_ssl_policy" "tls" {
  name = "ELBSecurityPolicy-TLS-1-2-2017-01"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the policy.

## Attributes Reference

The following attributes are supported:

* `id` - The name of the policy.
* `name` - The name of the policy.
