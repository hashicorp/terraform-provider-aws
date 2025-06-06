---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_saml_provider"
description: |-
  Get information on an IAM SAML provider.
---


<!-- Please do not edit this file, it is generated. -->
# Data Source: aws_iam_saml_provider

This data source can be used to fetch information about a specific
IAM SAML provider. This will allow you to easily retrieve the metadata
document of an existing SAML provider.

## Example Usage

```typescript
// DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
import { Construct } from "constructs";
import { TerraformStack } from "cdktf";
/*
 * Provider bindings are generated by running `cdktf get`.
 * See https://cdk.tf/provider-generation for more details.
 */
import { DataAwsIamSamlProvider } from "./.gen/providers/aws/data-aws-iam-saml-provider";
class MyConvertedCode extends TerraformStack {
  constructor(scope: Construct, name: string) {
    super(scope, name);
    new DataAwsIamSamlProvider(this, "example", {
      arn: "arn:aws:iam::123456789:saml-provider/myprovider",
    });
  }
}

```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Required) ARN assigned by AWS for the provider.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `createDate` - Creation date of the SAML provider in RFC1123 format, e.g. `Mon, 02 Jan 2006 15:04:05 MST`.
* `name` - Name of the provider.
* `samlMetadataDocument` - The XML document generated by an identity provider that supports SAML 2.0.
* `tags` - Tags attached to the SAML provider.
* `validUntil` - Expiration date and time for the SAML provider in RFC1123 format, e.g. `Mon, 02 Jan 2007 15:04:05 MST`.

<!-- cache-key: cdktf-0.20.8 input-6e43cd667085c5de501c5b9360c134f469f76ebffca96dc5bbea2c9305485db0 -->