---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_product"
description: |-
  This data source provides information about a Service Catalog product.
---

# Data Source: aws_servicecatalog_product

Use this data source to retrieve information about a Service Catalog product.

~> **NOTE:** A "provisioning artifact" is also known as a "version," and a "distributor" is also known as a "vendor."

## Example Usage

### Basic Usage

```terraform
data "aws_servicecatalog_product" "example" {
  id = "prod-dnigbtea24ste"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) ID of the product.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values are `en` (English), `jp` (Japanese), `zh` (Chinese). The default value is `en`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the product.
* `created_time` - Time when the product was created.
* `description` - Description of the product.
* `distributor` - Vendor of the product.
* `has_default_path` - Whether the product has a default path.
* `name` - Name of the product.
* `owner` - Owner of the product.
* `status` - Status of the product.
* `support_description` - Field that provides support information about the product.
* `support_email` - Contact email for product support.
* `support_url` - Contact URL for product support.
* `tags` - Tags applied to the product.
* `type` - Type of product.
