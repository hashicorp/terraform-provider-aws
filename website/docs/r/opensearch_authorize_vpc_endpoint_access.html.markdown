---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_authorize_vpc_endpoint_access"
description: |-
  Terraform resource for managing an AWS OpenSearch Authorize Vpc Endpoint Access.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->`
# Resource: aws_opensearch_authorize_vpc_endpoint_access

Terraform resource for managing an AWS OpenSearch Authorize Vpc Endpoint Access.

## Example Usage

### Basic Usage

```terraform
resource "aws_opensearch_authorize_vpc_endpoint_access" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Authorize Vpc Endpoint Access. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch Authorize Vpc Endpoint Access using the `example_id_arg`. For example:

```terraform
import {
  to = aws_opensearch_authorize_vpc_endpoint_access.example
  id = "authorize_vpc_endpoint_access-id-12345678"
}
```

Using `terraform import`, import OpenSearch Authorize Vpc Endpoint Access using the `example_id_arg`. For example:

```console
% terraform import aws_opensearch_authorize_vpc_endpoint_access.example authorize_vpc_endpoint_access-id-12345678
```
