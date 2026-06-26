---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_hub_content_reference"
description: |-
  Manages a SageMaker AI Hub Content Reference resource.
---

# Resource: aws_sagemaker_hub_content_reference

Manages a SageMaker AI Hub Content Reference resource. A hub content reference copies a model from the SageMaker JumpStart public hub into a private hub so that it is accessible to users in that hub.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_hub" "example" {
  hub_name        = "example"
  hub_description = "example"
}

resource "aws_sagemaker_hub_content_reference" "example" {
  hub_name                         = aws_sagemaker_hub.example.hub_name
  hub_content_name                 = "example-llama"
  sagemaker_public_hub_content_arn = "arn:aws:sagemaker:us-east-1:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct"
}
```

### With minimum version

```terraform
resource "aws_sagemaker_hub_content_reference" "example" {
  hub_name                         = aws_sagemaker_hub.example.hub_name
  hub_content_name                 = "example-llama"
  sagemaker_public_hub_content_arn = "arn:aws:sagemaker:us-east-1:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct"
  min_version                      = "1.0.0"
}
```

## Argument Reference

This resource supports the following arguments:

* `hub_content_name` - (Required) Name of the hub content reference.
* `hub_name` - (Required) Name of the private SageMaker Hub to add the content reference to.
* `min_version` - (Optional) Minimum version of the hub content to reference. Use `"1.0.0"` to support all versions. Changing this value to an empty string forces replacement of the resource.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sagemaker_public_hub_content_arn` - (Required) ARN of the public SageMaker JumpStart hub content to reference. The ARN must not include a version suffix.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `hub_arn` - ARN of the private SageMaker Hub that contains the content reference.
* `hub_content_arn` - ARN of the hub content reference (without version suffix). The min_version is stripped off from the end of this ARN to make it usable to list tags.
* `hub_content_status` - Status of the hub content reference. Valid values include `Available`, `Importing`, `Deleting`, `ImportFailed`, `DeleteFailed`.
* `hub_content_version` - Version of the hub content reference.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `delete` - (Default `15m`)
* `update` - (Default `15m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_sagemaker_hub_content_reference.example
  identity = {
    hub_name         = "my-hub"
    hub_content_name = "my-content-reference"
  }
}

resource "aws_sagemaker_hub_content_reference" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `hub_content_name` - (String) Name of the hub content reference.
* `hub_name` - (String) Name of the private SageMaker Hub.

#### Optional

* `account_id` - (String) AWS account where this resource is managed.
* `region` - (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Hub Content References using `hub_name,hub_content_name`. For example:

```terraform
import {
  to = aws_sagemaker_hub_content_reference.example
  id = "my-hub,my-content-reference"
}
```

Using `terraform import`, import SageMaker AI Hub Content References using `hub_name,hub_content_name`. For example:

```console
% terraform import aws_sagemaker_hub_content_reference.example my-hub,my-content-reference
```
