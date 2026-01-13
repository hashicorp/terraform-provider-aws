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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the http namespace.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of a namespace.
* `arn` - ARN that Amazon Route 53 assigns to the namespace when you create it.
* `description` - Description that you specify for the namespace when you create it.
* `http_name` - Name of an HTTP namespace.
* `tags` - Map of tags for the resource.
