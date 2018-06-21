---
layout: "aws"
page_title: "AWS: aws_servicecatalog_product"
sidebar_current: "docs-aws-resource-servicecatalog-product"
description: |-
  Provides a resource to create a Service Catalog Product
---

# aws_servicecatalog_product

Provides a resource to create a Service Catalog Product.
Noted that this only creates the very first Provisioning Artifact that comes along with the product being created.

## Example Usage

```hcl
resource "aws_servicecatalog_product" "test" {
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = "My First Product"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "arbitrary support description"
  support_email       = "arbitrary@email.com"
  support_url         = "http://arbitrary_url/foo.html"

  provisioning_artifact {
    description = "arbitrary description"
    name        = "v1.0.0"
    info {
      LoadTemplateFromURL = "https://s3.amazonaws.com/bucket-xyz/cloudformation.json"
    }
  }

  tags {
     Foo = "bar"
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of the product.
* `destributor` - (Optional) The distributor of the product.
* `name` - (Required) The name of the product.
* `owner` - (Required) The owner of the product.
* `product_type` - (Required) The type of product. Valid values: `CLOUD_FORMATION_TEMPLATE` or `MARKETPLACE`
* `provisioning_artifact` - (Required) The configuration of the provisioning artifact. This object supports the following:
  * `description` - (Optional) The description of the provisioning artifact.
  * `name` - (Required) The name of the provisioning artifact (for example, v1 v2beta). No spaces are allowed.
  * `type` - (Optional) The type of provisioning artifact. Valid Values: `CLOUD_FORMATION_TEMPLATE` or `MARKETPLACE_AMI` or `MARKETPLACE_CAR`. Default is `CLOUD_FORMATION_TEMPLATE`
  * `info` - (Required) The URL of the CloudFormation template in Amazon S3. Specify the URL as `LoadTemplateFromURL = https://s3.amazonaws.com/cf-templates-ozkq9d3hgiq2-us-east-1/...`
* `support_description` - (Optional) The support information about the product.
* `support_email` - (Optional) The contact email for product support.
* `support_url` - (Optional) The contact URL for product support.
* `tags` - (Optional) Tags to apply to the connection.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Service Catalog Product.
* `product_arn` - The ARN of the product.
* `has_default_path` - Indicates whether the product has a default path.
* `provisioning_artifact` - Object attributes that are exported are:
  * `id` - The ID of the Provisioning Artifact
  * `active` - Indicates whether the product version is active.
  * `created_time` - The UTC time stamp of the creation time.

## Import

Service Catalog Products can be imported using the service catalog product id, e.g.

```
$ terraform import aws_servicecatalog_product.test p-12344321
```
