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

* `name` - (Required) The name of the http namespace.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of a namespace.
* `arn` - The ARN that Amazon Route 53 assigns to the namespace when you create it.
* `description` - The description that you specify for the namespace when you create it.
* `http_name` - The name of an HTTP namespace.
* `tags` - A map of tags to assign to the namespace. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

