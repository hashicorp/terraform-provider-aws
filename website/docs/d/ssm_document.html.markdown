---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_document"
description: |-
  Provides a SSM Document data source
---

# Data Source: aws_ssm_document

Gets the contents of the specified Systems Manager document.

## Example Usage

To get the contents of the document owned by AWS.

```terraform
data "aws_ssm_document" "foo" {
  name            = "AWS-GatherSoftwareInventory"
  document_format = "YAML"
}

output "content" {
  value = data.aws_ssm_document.foo.content
}
```

To get the contents of the custom document.

```terraform
data "aws_ssm_document" "test" {
  name            = aws_ssm_document.test.name
  document_format = "JSON"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) The name of the document.
* `document_format` - The format of the document. Valid values: `JSON`, `TEXT`, `YAML`.
* `document_version` - The document version.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the document. If the document is an AWS managed document, this value will be set to the name of the document instead.
* `content` - The content for the SSM document in JSON or YAML format.
* `document_type` - The type of the document.
