---
subcategory: "Service Catalog AppRegistry"
layout: "aws"
page_title: "AWS: aws_servicecatalogappregistry_application"
description: |-
  Terraform data source for managing an AWS Service Catalog AppRegistry Application.
---

# Data Source: aws_servicecatalogappregistry_application

Terraform data source for managing an AWS Service Catalog AppRegistry Application.

## Example Usage

### Basic Usage

```terraform
data "aws_servicecatalogappregistry_application" "example" {
  id = "application-1234"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Application identifier.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `application_tag` - A map with a single tag key-value pair used to associate resources with the application.
* `arn` - ARN (Amazon Resource Name) of the application.
* `description` - Description of the application.
* `name` - Name of the application.
