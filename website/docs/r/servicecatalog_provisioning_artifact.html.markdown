---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_provisioning_artifact"
description: |-
  Manages a Service Catalog Provisioning Artifact
---

# Resource: aws_servicecatalog_provisioning_artifact

Manages a Service Catalog Provisioning Artifact for a specified product.

-> A "provisioning artifact" is also referred to as a "version."

~> **NOTE:** You cannot create a provisioning artifact for a product that was shared with you.

~> **NOTE:** The user or role that use this resource must have the `cloudformation:GetTemplate` IAM policy permission. This policy permission is required when using the `template_physical_id` argument.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_provisioning_artifact" "example" {
  name         = "example"
  product_id   = aws_servicecatalog_product.example.id
  type         = "CLOUD_FORMATION_TEMPLATE"
  template_url = "https://${aws_s3_bucket.example.bucket_regional_domain_name}/${aws_s3_object.example.key}"
}
```

## Argument Reference

The following arguments are required:

* `product_id` - (Required) Identifier of the product.
* `template_physical_id` - (Required if `template_url` is not provided) Template source as the physical ID of the resource that contains the template. Currently only supports CloudFormation stack ARN. Specify the physical ID as `arn:[partition]:cloudformation:[region]:[account ID]:stack/[stack name]/[resource ID]`.
* `template_url` - (Required if `template_physical_id` is not provided) Template source as URL of the CloudFormation template in Amazon S3.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). The default value is `en`.
* `active` - (Optional) Whether the product version is active. Inactive provisioning artifacts are invisible to end users. End users cannot launch or update a provisioned product from an inactive provisioning artifact. Default is `true`.
* `description` - (Optional) Description of the provisioning artifact (i.e., version), including how it differs from the previous provisioning artifact.
* `disable_template_validation` - (Optional) Whether AWS Service Catalog stops validating the specified provisioning artifact template even if it is invalid.
* `guidance` - (Optional) Information set by the administrator to provide guidance to end users about which provisioning artifacts to use. Valid values are `DEFAULT` and `DEPRECATED`. The default is `DEFAULT`. Users are able to make updates to a provisioned product of a deprecated version but cannot launch new provisioned products using a deprecated version.
* `name` - (Optional) Name of the provisioning artifact (for example, `v1`, `v2beta`). No spaces are allowed.
* `type` - (Optional) Type of provisioning artifact. See [AWS Docs](https://docs.aws.amazon.com/servicecatalog/latest/dg/API_ProvisioningArtifactProperties.html) for valid list of values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_time` - Time when the provisioning artifact was created.
* `id` - Provisioning artifact identifier and product identifier separated by a colon.
* `provisioning_artifact_id` - Provisioning artifact identifier.
* `status` - Status of the provisioning artifact.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `3m`)
- `read` - (Default `10m`)
- `update` - (Default `3m`)
- `delete` - (Default `3m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_servicecatalog_provisioning_artifact` using the provisioning artifact ID and product ID separated by a colon. For example:

```terraform
import {
  to = aws_servicecatalog_provisioning_artifact.example
  id = "pa-ij2b6lusy6dec:prod-el3an0rma3"
}
```

Using `terraform import`, import `aws_servicecatalog_provisioning_artifact` using the provisioning artifact ID and product ID separated by a colon. For example:

```console
% terraform import aws_servicecatalog_provisioning_artifact.example pa-ij2b6lusy6dec:prod-el3an0rma3
```
