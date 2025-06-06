---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_application_providers"
description: |-
  Terraform data source for managing AWS SSO Admin Application Providers.
---


<!-- Please do not edit this file, it is generated. -->
# Data Source: aws_ssoadmin_application_providers

Terraform data source for managing AWS SSO Admin Application Providers.

## Example Usage

### Basic Usage

```typescript
// DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
import { Construct } from "constructs";
import { TerraformStack } from "cdktf";
/*
 * Provider bindings are generated by running `cdktf get`.
 * See https://cdk.tf/provider-generation for more details.
 */
import { DataAwsSsoadminApplicationProviders } from "./.gen/providers/aws/data-aws-ssoadmin-application-providers";
class MyConvertedCode extends TerraformStack {
  constructor(scope: Construct, name: string) {
    super(scope, name);
    new DataAwsSsoadminApplicationProviders(this, "example", {});
  }
}

```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS region.
* `applicationProviders` - A list of application providers available in the current region. See [`applicationProviders`](#application_providers-attribute-reference) below.

### `applicationProviders` Attribute Reference

* `applicationProviderArn` - ARN of the application provider.
* `displayData` - An object describing how IAM Identity Center represents the application provider in the portal. See [`displayData`](#display_data-attribute-reference) below.
* `federationProtocol` - Protocol that the application provider uses to perform federation. Valid values are `SAML` and `OAUTH`.

### `displayData` Attribute Reference

* `description` - Description of the application provider.
* `displayName` - Name of the application provider.
* `iconUrl` - URL that points to an icon that represents the application provider.

<!-- cache-key: cdktf-0.20.8 input-5f39dd203a382ab694a8464d271f34ab0c12fdd5b7e8f85a29d2707959cd666a -->