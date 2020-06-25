---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_product"
description: |-
  Provides a resource for a Service Catalog Product definition from which instances can be provisioned.
---

# aws_servicecatalog_product

Provides a resource to create a Service Catalog Product.

Note that the way AWS Service Catalog works,
this Product definition is not associated to any portfolio, 
it has no launch constraints,
and it only creates the very first Provisioning Artifact that comes along with the product being created.
It is necessary to create an `aws_servicecatalog_portfolio_product_association` 
to link this to a portfolio before it can be provisioned,
and for that portfolio to have an `aws_servicecatalog_portfolio_principal_association` explicitly 
opting-in those principals (users, groups, roles) which can access it.
It is also common but not necessary to create launch constraints and additional provisioning artifacts
to govern the usage of this Product. 

## Example Usage

A rich example assuming the JSON is in a bucket already:

```hcl
resource "aws_servicecatalog_product" "test_existing_template" {
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = "My Existing Product"
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

A minimal product example which also uploads a template to a bucket:

```hcl
resource "aws_s3_bucket" "my_product_template_bucket" {
  bucket        = "%s"
  region        = "us-east-1"
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "my_product_template_object" {
  bucket  = aws_s3_bucket.my_product_template_bucket.id
  key     = "my_product_template.json"
  content = <<EOF
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Test CF teamplate for Service Catalog terraform dev",
  "Resources": {
    "Empty": {
      "Type": "AWS::CloudFormation::WaitConditionHandle"
    }
  }
}
EOF
}

resource "aws_servicecatalog_product" "minimal_test_new_template" {
  name                = "My New Minimal Product"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"

  provisioning_artifact {
    name        = "v1.0.0"
    info {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.my_product_template_bucket.id}/${aws_s3_bucket_object.my_product_template_object.key}"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of the product.
* `distributor` - (Optional) The distributor of the product.
* `name` - (Required) The name of the product.
* `owner` - (Required) The owner of the product.
* `product_type` - (Required) The type of product. Valid values: `CLOUD_FORMATION_TEMPLATE` or `MARKETPLACE`
* `provisioning_artifact` - (Required) The configuration of the provisioning artifact. This object supports the following:
  * `description` - (Optional) The description of the provisioning artifact.
  * `name` - (Required) The name of the provisioning artifact (for example, v1 v2beta). No spaces are allowed.
  * `type` - (Optional) The type of provisioning artifact. Valid Values: `CLOUD_FORMATION_TEMPLATE` or `MARKETPLACE_AMI` or `MARKETPLACE_CAR`. Default is `CLOUD_FORMATION_TEMPLATE`
  * `info` - (Required) The URL of the CloudFormation template in Amazon S3. Specify the URL as a map with one entry of the form `LoadTemplateFromURL = "https://s3.amazonaws.com/bucket/file.json"`. 
* `support_description` - (Optional) The support information about the product.
* `support_email` - (Optional) The contact email for product support.
* `support_url` - (Optional) The contact URL for product support.
* `tags` - (Optional) Tags to apply to the product.

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
$ terraform import aws_servicecatalog_product.test prod-a1b2c3d4e5f6g
```
