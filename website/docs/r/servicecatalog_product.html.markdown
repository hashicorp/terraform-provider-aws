---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_product"
description: |-
  Manages a Service Catalog Product
---

# Resource: aws_servicecatalog_product

Manages a Service Catalog Product.

~> **NOTE:** The user or role that uses this resources must have the `cloudformation:GetTemplate` IAM policy permission. This policy permission is required when using the `template_physical_id` argument.

-> A "provisioning artifact" is also referred to as a "version." A "distributor" is also referred to as a "vendor."

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_product" "example" {
  name  = "example"
  owner = "example-owner"
  type  = "CLOUD_FORMATION_TEMPLATE"

  provisioning_artifact_parameters {
    template_url = "https://s3.amazonaws.com/cf-templates-ozkq9d3hgiq2-us-east-1/temp1.json"
  }

  tags = {
    foo = "bar"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the product.
* `owner` - (Required) Owner of the product.
* `provisioning_artifact_parameters` - (Required) Configuration block for provisioning artifact (i.e., version) parameters. See [`provisioning_artifact_parameters` Block](#provisioning_artifact_parameters-block) for details.
* `type` - (Required) Type of product. See [AWS Docs](https://docs.aws.amazon.com/servicecatalog/latest/dg/API_CreateProduct.html#API_CreateProduct_RequestSyntax) for valid list of values.

The following arguments are optional:

* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.
* `description` - (Optional) Description of the product.
* `distributor` - (Optional) Distributor (i.e., vendor) of the product.
* `support_description` - (Optional) Support information about the product.
* `support_email` - (Optional) Contact email for product support.
* `support_url` - (Optional) Contact URL for product support.
* `tags` - (Optional) Tags to apply to the product. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `provisioning_artifact_parameters` Block

The `provisioning_artifact_parameters` configuration block supports the following arguments:

* `description` - (Optional) Description of the provisioning artifact (i.e., version), including how it differs from the previous provisioning artifact.
* `disable_template_validation` - (Optional) Whether AWS Service Catalog stops validating the specified provisioning artifact template even if it is invalid.
* `name` - (Optional) Name of the provisioning artifact (for example, `v1`, `v2beta`). No spaces are allowed.
* `template_physical_id` - (Required if `template_url` is not provided) Template source as the physical ID of the resource that contains the template. Currently only supports CloudFormation stack ARN. Specify the physical ID as `arn:[partition]:cloudformation:[region]:[account ID]:stack/[stack name]/[resource ID]`.
* `template_url` - (Required if `template_physical_id` is not provided) Template source as URL of the CloudFormation template in Amazon S3.
* `type` - (Optional) Type of provisioning artifact. See [AWS Docs](https://docs.aws.amazon.com/servicecatalog/latest/dg/API_ProvisioningArtifactProperties.html) for valid list of values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the product.
* `created_time` - Time when the product was created.
* `has_default_path` - Whether the product has a default path. If the product does not have a default path, call `ListLaunchPaths` to disambiguate between paths.  Otherwise, `ListLaunchPaths` is not required, and the output of ProductViewSummary can be used directly with `DescribeProvisioningParameters`.
* `id` - Product ID. For example, `prod-dnigbtea24ste`.
* `status` - Status of the product.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `read` - (Default `10m`)
- `update` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_servicecatalog_product` using the product ID. For example:

```terraform
import {
  to = aws_servicecatalog_product.example
  id = "prod-dnigbtea24ste"
}
```

Using `terraform import`, import `aws_servicecatalog_product` using the product ID. For example:

```console
% terraform import aws_servicecatalog_product.example prod-dnigbtea24ste
```
