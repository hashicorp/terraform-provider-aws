---
layout: "aws"
page_title: "AWS: aws_service_discovery_dns_namespace"
sidebar_current: "docs-aws-service-discovery-dns-namespace"
description: |-
  Retrieve information about service discovery private and public dns namespace
---

# Data Source: aws_service_discovery_dns_namespace

## Example Usage


```hcl
data "aws_service_discovery_dns_namespace" "test" {
  name = "example.terraform.local"
  dns_type = "DNS_PRIVATE"
}
```

## Argument Reference

* `name` - (Required) The name of the service discovery dns namespace.
* `dns_type` - (Required) The type of the namespace. Allowed values are DNS_PUBLIC or DNS_PRIVATE.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) of the namespace.
* `description` - A description of the namespace.
* `id` - The namespace ID.
* `hosted_zone` - The ID for the Route 53 hosted zone that AWS Cloud Map creates when you create a namespace.