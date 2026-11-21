---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_application_capability"
description: |-
  Manages a capability registered on an Amazon OpenSearch UI application.
---

# Resource: aws_opensearch_application_capability

Manages a capability registered on an Amazon OpenSearch UI application. Capabilities enable optional features on an application, such as the AI Assistant.

For more information, see [Agentic AI for OpenSearch UI](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/application-ai-assistant.html) in the *Amazon OpenSearch Service Developer Guide*.

## Example Usage

### AI capability

```terraform
resource "aws_opensearch_application" "example" {
  name = "example"
}

resource "aws_opensearch_application_capability" "ai" {
  application_id  = aws_opensearch_application.example.id
  capability_name = "ai-capability"

  ai_config {}
}
```

## Argument Reference

The following arguments are required:

* `application_id` - (Required) Identifier of the OpenSearch UI application to register the capability on. Changing this value forces a new resource.
* `capability_name` - (Required) Name of the capability. Currently, `ai-capability` is the only supported value. Changing this value forces a new resource.

The following arguments are optional:

* `ai_config` - (Optional) Enables AI-powered features (chatbot, agentic investigation) on the application. Exactly one capability configuration block must be provided. The block accepts no arguments today. Adding or removing this block forces a new resource.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `status` - Current status of the capability. One of `creating`, `create_failed`, `active`, `updating`, `update_failed`, `deleting`, `delete_failed`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an OpenSearch application capability using the application ID and capability name separated by a comma (`,`). For example:

```terraform
import {
  to = aws_opensearch_application_capability.ai
  id = "abc123,ai-capability"
}
```

Using `terraform import`, import an OpenSearch application capability using the application ID and capability name separated by a comma (`,`). For example:

```console
% terraform import aws_opensearch_application_capability.ai abc123,ai-capability
```
