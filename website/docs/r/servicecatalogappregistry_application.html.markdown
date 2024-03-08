---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalogappregistry_application"
description: |-
  Terraform resource for managing an AWS Service Catalog AppRegistry Application.
---
# Resource: aws_servicecatalogappregistry_application

Terraform resource for managing an AWS Service Catalog AppRegistry Application. 

~> **NOTE: ** An AWS Service Catalog AppRegistry Application is also available in the AWS Console as "MyApplications"

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalogappregistry_application" "example" {
  name = "myApplication"
}
```

### Connecting resources

```terraform
resource "aws_servicecatalogappregistry_application" "example" {
  name = "myApplication"
}

resource "aws_s3_bucket" "bucket" {
  bucket = "application_bucket"

  tags = {
    awsApplication = aws_servicecatalogappregistry_application.example.arn
  }
}

```

## Argument Reference

The following arguments are required:

* `Name` - (Required) Name of the application. The name must be unique in the region in which you are creating the application.

The following arguments are optional:

* `description` - (Optional) Description of the application.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Application.
* `id` - Identifier of the Application

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
