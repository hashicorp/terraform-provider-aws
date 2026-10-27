---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_hub_content"
description: |-
  Provides details about an AWS SageMaker AI Hub Content.
---

# Data Source: aws_sagemaker_hub_content

Provides details about an AWS SageMaker AI Hub Content, such as a model in the SageMaker JumpStart public hub or a model reference in a private hub.

## Example Usage

### Public Hub Model

```terraform
data "aws_sagemaker_hub_content" "example" {
  hub_name         = "SageMakerPublicHub"
  hub_content_name = "meta-textgeneration-llama-2-13b-f"
  hub_content_type = "Model"
}

locals {
  hosting_artifact_uri = jsondecode(data.aws_sagemaker_hub_content.example.hub_content_document).HostingArtifactUri
}
```

### Private Hub Model Reference

```terraform
data "aws_sagemaker_hub_content" "example" {
  hub_name         = aws_sagemaker_hub_content_reference.example.hub_name
  hub_content_name = aws_sagemaker_hub_content_reference.example.hub_content_name
  hub_content_type = "ModelReference"
}
```

## Argument Reference

The following arguments are required:

* `hub_content_name` - (Required) Name of the hub content.
* `hub_content_type` - (Required) Type of the hub content. Valid values: `Model`, `Notebook`, `ModelReference`, `DataSet`, `JsonDoc`.
* `hub_name` - (Required) Name of the hub that contains the content.

The following arguments are optional:

* `hub_content_version` - (Optional) Version of the hub content. Defaults to the latest version.
* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `creation_time` - Date and time the hub content was created.
* `document_schema_version` - Version of the hub content document schema.
* `failure_reason` - Failure reason if importing the hub content failed.
* `hub_arn` - ARN of the hub that contains the content.
* `hub_content_arn` - ARN of the hub content, without the version suffix.
* `hub_content_dependency` - List of hub content dependencies. See [`hub_content_dependency`](#hub_content_dependency).
* `hub_content_description` - Description of the hub content.
* `hub_content_display_name` - Display name of the hub content.
* `hub_content_document` - JSON hub content document, including details such as the hosting image and artifact URIs.
* `hub_content_markdown` - Markdown content of the hub content.
* `hub_content_search_keywords` - List of searchable keywords for the hub content.
* `hub_content_status` - Status of the hub content.
* `last_modified_time` - Date and time the hub content was last modified.
* `reference_min_version` - Minimum version of the referenced public hub content, for `ModelReference` hub content.
* `sagemaker_public_hub_content_arn` - ARN of the referenced public hub content, without the version suffix, for `ModelReference` hub content.
* `support_status` - Support status of the hub content.
* `versioned_hub_content_arn` - ARN of the hub content, including the version suffix.

### `hub_content_dependency`

* `dependency_copy_path` - Hub content dependency copy path.
* `dependency_origin_path` - Hub content dependency origin path.
