---
subcategory: "B2B Data Interchange"
layout: "aws"
page_title: "AWS: aws_b2bi_capability"
description: |-
  Manages a B2BI Capability.
---

# Resource: aws_b2bi_capability

Manages an AWS B2B Data Interchange Capability. A capability contains the information required to transform incoming EDI documents into JSON or XML outputs.

## Example Usage

```terraform
resource "aws_b2bi_capability" "example" {
  name = "example-capability"
  type = "edi"

  configuration {
    edi {
      input_location {
        bucket_name = aws_s3_bucket.example.bucket
        key         = "input/"
      }

      output_location {
        bucket_name = aws_s3_bucket.example.bucket
        key         = "output/"
      }

      transformer_id = aws_b2bi_transformer.example.transformer_id

      type {
        x12_details {
          transaction_set = "X12_110"
          version         = "VERSION_4010"
        }
      }
    }
  }
}
```

## Argument Reference

* `name` - (Required) The name of the capability.
* `type` - (Required, Forces new resource) The type of capability. Currently only `edi` is supported.
* `configuration` - (Required) The capability configuration. See [`configuration` Block](#configuration-block).
* `instructions_documents` - (Optional) S3 locations for instruction documents. See [`instructions_documents` Block](#instructions_documents-block).
* `tags` - (Optional) A map of tags to assign to the resource.

### `configuration` Block

* `edi` - (Required) EDI configuration.
    * `input_location` - (Required) S3 location for input documents.
        * `bucket_name` - (Required) The S3 bucket name.
        * `key` - (Required) The S3 key prefix.
    * `output_location` - (Required) S3 location for output documents.
        * `bucket_name` - (Required) The S3 bucket name.
        * `key` - (Required) The S3 key prefix.
    * `transformer_id` - (Required) The ID of the transformer to use. Must be in `active` status.
    * `type` - (Required) The EDI type configuration.
        * `x12_details` - (Required) X12 format details.
            * `transaction_set` - (Optional) The X12 transaction set.
            * `version` - (Optional) The X12 version.

### `instructions_documents` Block

* `bucket_name` - (Required) The S3 bucket name.
* `key` - (Required) The S3 key.

## Attribute Reference

* `capability_arn` - The ARN of the capability.
* `capability_id` - The unique identifier of the capability.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

B2BI Capabilities can be imported using the `capability_id`:

```console
% terraform import aws_b2bi_capability.example ca-1111aaaa2222bbbb3
```

~> **Note:** The `configuration.edi.type` block is not returned by the API on read. After import, you may need to run `terraform apply` once to set this value.
