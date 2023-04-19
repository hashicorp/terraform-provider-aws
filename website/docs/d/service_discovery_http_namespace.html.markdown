---
subcategory: "Cloud Map"
layout: "aws"
page_title: "AWS: aws_service_discovery_http_namespace"
description: |-
  Retrieves information about a Service Discovery HTTP Namespace.
---

# Data Source: aws_service_discovery_http_namespace

## Example Usage

```terraform
data "aws_service_discovery_http_namespace" "example" {
  name = "development"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the http namespace.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of a namespace.
* `arn` - ARN that Amazon Route 53 assigns to the namespace when you create it.
* `description` - Description that you specify for the namespace when you create it.
* `http_name` - Name of an HTTP namespace.
* `tags` - Map of tags for the resource.
