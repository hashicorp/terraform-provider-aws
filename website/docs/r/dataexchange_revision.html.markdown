---
subcategory: "Data Exchange"
layout: "aws"
page_title: "AWS: aws_dataexchange_revision"
description: |-
  Provides a DataExchange Revision
---

# Resource: aws_dataexchange_revision

Provides a resource to manage AWS Data Exchange Revisions.

## Example Usage

```terraform
resource "aws_dataexchange_revision" "example" {
  data_set_id = aws_dataexchange_data_set.example.id
}
```

## Argument Reference

* `data_set_id` - (Required) The dataset id.
* `comment` - (Required) An optional comment about the revision.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Id of the data set.
* `revision_id` - The Id of the revision.
* `arn` - The Amazon Resource Name of this data set.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataExchange Revisions using their `data-set-id:revision-id`. For example:

```terraform
import {
  to = aws_dataexchange_revision.example
  id = "4fa784c7-ccb4-4dbf-ba4f-02198320daa1:4fa784c7-ccb4-4dbf-ba4f-02198320daa1"
}
```

Using `terraform import`, import DataExchange Revisions using their `data-set-id:revision-id`. For example:

```console
% terraform import aws_dataexchange_revision.example 4fa784c7-ccb4-4dbf-ba4f-02198320daa1:4fa784c7-ccb4-4dbf-ba4f-02198320daa1
```
