---
subcategory: "Data Pipeline"
layout: "aws"
page_title: "AWS: aws_datapipeline_pipeline"
description: |-
  Provides a AWS DataPipeline Pipeline.
---

# Resource: aws_datapipeline_pipeline

Provides a DataPipeline Pipeline resource.

## Example Usage

```terraform
resource "aws_datapipeline_pipeline" "default" {
  name = "tf-pipeline-default"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of Pipeline.
* `description` - (Optional) The description of Pipeline.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The identifier of the client certificate.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_datapipeline_pipeline` using the id (Pipeline ID). For example:

```terraform
import {
  to = aws_datapipeline_pipeline.default
  id = "df-1234567890"
}
```

Using `terraform import`, import `aws_datapipeline_pipeline` using the id (Pipeline ID). For example:

```console
% terraform import aws_datapipeline_pipeline.default df-1234567890
```
