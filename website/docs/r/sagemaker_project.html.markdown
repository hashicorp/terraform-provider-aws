---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_project"
description: |-
  Provides a SageMaker AI Project resource.
---

# Resource: aws_sagemaker_project

Provides a SageMaker AI Project resource.

 -> Note: If you are trying to use SageMaker AI projects with SageMaker AI studio you will need to add a tag with the key `sagemaker:studio-visibility` with value `true`. For more on requirements to use projects and permission needed see [AWS Docs](https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-projects-templates-custom.html).

## Example Usage

```terraform
resource "aws_sagemaker_project" "example" {
  project_name = "example"

  service_catalog_provisioning_details {
    product_id = aws_servicecatalog_product.example.id
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `project_name` - (Required) The name of the Project.
* `project_description` - (Optional) A description for the project.
* `service_catalog_provisioning_details` - (Required) The product ID and provisioning artifact ID to provision a service catalog. See [Service Catalog Provisioning Details](#service-catalog-provisioning-details) below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Service Catalog Provisioning Details

* `path_id` - (Optional) The path identifier of the product. This value is optional if the product has a default path, and required if the product has more than one path.
* `product_id` - (Required) The ID of the product to provision.
* `provisioning_artifact_id` - (Optional) The ID of the provisioning artifact.
* `provisioning_parameter` - (Optional) A list of key value pairs that you specify when you provision a product. See [Provisioning Parameter](#provisioning-parameter) below.

#### Provisioning Parameter

* `key` - (Required) The key that identifies a provisioning parameter.
* `value` - (Optional) The value of the provisioning parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Project.
* `id` - The name of the Project.
* `project_id` - The ID of the project.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Projects using the `project_name`. For example:

```terraform
import {
  to = aws_sagemaker_project.example
  id = "example"
}
```

Using `terraform import`, import SageMaker AI Projects using the `project_name`. For example:

```console
% terraform import aws_sagemaker_project.example example
```
