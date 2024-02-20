---
subcategory: "Chime SDK Voice"
layout: "aws"
page_title: "AWS: aws_chimesdkvoice_sip_media_application"
description: |-
  A ChimeSDKVoice SIP Media Application is a managed object that passes values from a SIP rule to a target AWS Lambda function.
---

# Resource: aws_chimesdkvoice_sip_media_application

A ChimeSDKVoice SIP Media Application is a managed object that passes values from a SIP rule to a target AWS Lambda function.

## Example Usage

### Basic Usage

```terraform
resource "aws_chimesdkvoice_sip_media_application" "example" {
  aws_region = "us-east-1"
  name       = "example-sip-media-application"
  endpoints {
    lambda_arn = aws_lambda_function.test.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `aws_region` - (Required) The AWS Region in which the AWS Chime SDK Voice Sip Media Application is created.
* `endpoints` - (Required)  List of endpoints (Lambda Amazon Resource Names) specified for the SIP media application. Currently, only one endpoint is supported. See [`endpoints`](#endpoints).
* `name` - (Required) The name of the AWS Chime SDK Voice Sip Media Application.

The following arguments are optional:

* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `endpoints`

The endpoint assigned to the SIP media application.

* `lambda_arn` - (Required) Valid Amazon Resource Name (ARN) of the Lambda function, version, or alias. The function must be created in the same AWS Region as the SIP media application.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` -  ARN (Amazon Resource Name) of the AWS Chime SDK Voice Sip Media Application
* `id` - The SIP media application ID.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a ChimeSDKVoice SIP Media Application using the `id`. For example:

```terraform
import {
  to = aws_chimesdkvoice_sip_media_application.example
  id = "abcdef123456"
}
```

Using `terraform import`, import a ChimeSDKVoice SIP Media Application using the `id`. For example:

```console
% terraform import aws_chimesdkvoice_sip_media_application.example abcdef123456
```
