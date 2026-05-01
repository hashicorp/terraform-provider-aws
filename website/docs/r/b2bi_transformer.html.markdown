---
subcategory: "B2B Data Interchange"
layout: "aws"
page_title: "AWS: aws_b2bi_transformer"
description: |-
  Manages a B2BI Transformer.
---

# Resource: aws_b2bi_transformer

Manages an AWS B2B Data Interchange Transformer. A transformer describes how to process incoming EDI documents and extract the relevant information, or how to generate outbound EDI documents from JSON/XML data.

See the [B2B Data Interchange User Guide](https://docs.aws.amazon.com/b2bi/latest/userguide/what-is-b2bi.html) for more information.

## Example Usage

### Inbound X12 Transformer

```terraform
resource "aws_b2bi_transformer" "example" {
  name = "example-transformer"

  input_conversion {
    from_format = "X12"

    format_options {
      x12 {
        transaction_set = "X12_110"
        version         = "VERSION_4010"
      }
    }
  }

  mapping {
    template_language = "JSONATA"
    template          = "{}"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the transformer.
* `input_conversion` - (Optional) The input conversion configuration. See [`input_conversion` Block](#input_conversion-block) for details.
* `mapping` - (Optional) The mapping template configuration. See [`mapping` Block](#mapping-block) for details.
* `output_conversion` - (Optional) The output conversion configuration. See [`output_conversion` Block](#output_conversion-block) for details.
* `sample_documents` - (Optional) Sample documents configuration. See [`sample_documents` Block](#sample_documents-block) for details.
* `status` - (Optional) The status of the transformer. Valid values are `active` and `inactive`.
* `tags` - (Optional) A map of tags to assign to the resource.

### `input_conversion` Block

* `from_format` - (Required) The format of the incoming document. Valid values are `X12`.
* `format_options` - (Optional) Format-specific options. See [`format_options` Block](#format_options-block) for details.

### `output_conversion` Block

* `to_format` - (Required) The format of the outgoing document. Valid values are `X12`.
* `format_options` - (Optional) Format-specific options. See [`format_options` Block](#format_options-block) for details.

### `format_options` Block

* `x12` - (Optional) X12 format options.
    * `transaction_set` - (Optional) The X12 transaction set (e.g., `X12_110`, `X12_210`, `X12_214`, `X12_810`, `X12_850`, `X12_855`, `X12_856`, `X12_860`, `X12_997`).
    * `version` - (Optional) The X12 version (e.g., `VERSION_4010`, `VERSION_4030`, `VERSION_5010`).

### `mapping` Block

* `template_language` - (Required) The language of the mapping template. Valid values are `JSONATA` and `XSLT`.
* `template` - (Optional) The mapping template content.

### `sample_documents` Block

* `bucket_name` - (Required) The S3 bucket name containing sample documents.
* `keys` - (Required) List of sample document keys.
    * `input` - (Optional) The S3 key for the input sample document.
    * `output` - (Optional) The S3 key for the output sample document.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `transformer_arn` - The Amazon Resource Name (ARN) of the transformer.
* `transformer_id` - The unique identifier of the transformer.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import B2BI Transformers using the `transformer_id`. For example:

```terraform
import {
  to = aws_b2bi_transformer.example
  id = "tr-1234abcd5678efghj"
}
```

Using `terraform import`, import B2BI Transformers using the `transformer_id`. For example:

```console
% terraform import aws_b2bi_transformer.example tr-1234abcd5678efghj
```
