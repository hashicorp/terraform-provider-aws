---
subcategory: "Service Catalog AppRegistry"
layout: "aws"
page_title: "AWS: aws_servicecatalogappregistry_application"
description: |-
  Terraform resource for managing an AWS Service Catalog AppRegistry Application.
---
# Resource: aws_servicecatalogappregistry_application

Terraform resource for managing an AWS Service Catalog AppRegistry Application.

~> An AWS Service Catalog AppRegistry Application is displayed in the AWS Console under "MyApplications".

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalogappregistry_application" "example" {
  name = "example-app"
}
```

### Connecting Resources

```terraform
resource "aws_servicecatalogappregistry_application" "example" {
  name = "example-app"
}

resource "aws_s3_bucket" "bucket" {
  bucket = "example-bucket"

  tags = aws_servicecatalogappregistry_application.example.application_tag
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the application. The name must be unique within an AWS region.

The following arguments are optional:

* `description` - (Optional) Description of the application.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `application_tag` - A map with a single tag key-value pair used to associate resources with the application. This attribute can be passed directly into the `tags` argument of another resource, or merged into a map of existing tags.
* `arn` - ARN (Amazon Resource Name) of the application.
* `id` - Identifier of the application.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Service Catalog AppRegistry Application using the `id`. For example:

```terraform
import {
  to = aws_servicecatalogappregistry_application.example
  id = "application-id-12345678"
}
```

Using `terraform import`, import AWS Service Catalog AppRegistry Application using the `id`. For example:

```console
% terraform import aws_servicecatalogappregistry_application.example application-id-12345678
```
