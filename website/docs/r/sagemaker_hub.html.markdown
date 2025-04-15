---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_hub"
description: |-
  Provides a SageMaker AI Hub resource.
---

# Resource: aws_sagemaker_hub

Provides a SageMaker AI Hub resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_hub" "example" {
  hub_name        = "example"
  hub_description = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `hub_name` - (Required) The name of the hub.
* `hub_description` - (Required) A description of the hub.
* `hub_display_name` - (Optional) The display name of the hub.
* `hub_search_keywords` - (Optional) The searchable keywords for the hub.
* `s3_storage_config` - (Optional) The Amazon S3 storage configuration for the hub. See [S3 Storage Config](#s3-storage-config) details below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### S3 Storage Config

* `s3_output_path` - (Optional) The Amazon S3 bucket prefix for hosting hub content.interface.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the Hub.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Hub.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Hubs using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_hub.test_hub
  id = "my-code-repo"
}
```

Using `terraform import`, import SageMaker AI Hubs using the `name`. For example:

```console
% terraform import aws_sagemaker_hub.test_hub my-code-repo
```
