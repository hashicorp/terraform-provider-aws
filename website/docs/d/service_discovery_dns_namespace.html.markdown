---
subcategory: "Cloud Map"
layout: "aws"
page_title: "AWS: aws_service_discovery_dns_namespace"
description: |-
  Retrieves information about a Service Discovery private or public DNS namespace.
---

# Data Source: aws_service_discovery_dns_namespace

Retrieves information about a Service Discovery private or public DNS namespace.

## Example Usage

```hcl
data "aws_service_discovery_dns_namespace" "test" {
  name = "example.terraform.local"
  type = "DNS_PRIVATE"
}
```

## Argument Reference

* `name` - (Required) Name of the namespace.
* `type` - (Required) Type of the namespace. Allowed values are `DNS_PUBLIC` or `DNS_PRIVATE`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the namespace.
* `description` - Description of the namespace.
* `id` - Namespace ID.
* `hosted_zone` - ID for the hosted zone that Amazon Route 53 creates when you create a namespace.
* `tags` - Map of tags for the resource.
