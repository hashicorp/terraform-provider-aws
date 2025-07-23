---
subcategory: "AppIntegrations"
layout: "aws"
page_title: "AWS: aws_appintegrations_data_integration"
description: |-
  Provides details about a specific Amazon AppIntegrations Data Integration
---

# Resource: aws_appintegrations_data_integration

Provides an Amazon AppIntegrations Data Integration resource.

## Example Usage

```terraform
resource "aws_appintegrations_data_integration" "example" {
  name        = "example"
  description = "example"
  kms_key     = aws_kms_key.test.arn
  source_uri  = "Salesforce://AppFlow/example"

  schedule_config {
    first_execution_from = "1439788442681"
    object               = "Account"
    schedule_expression  = "rate(1 hour)"
  }

  tags = {
    "Key1" = "Value1"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Specifies the description of the Data Integration.
* `kms_key` - (Required) Specifies the KMS key Amazon Resource Name (ARN) for the Data Integration.
* `name` - (Required) Specifies the name of the Data Integration.
* `schedule_config` - (Required) A block that defines the name of the data and how often it should be pulled from the source. The Schedule Config block is documented below.
* `source_uri` - (Required) Specifies the URI of the data source. Create an [AppFlow Connector Profile](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/appflow_connector_profile) and reference the name of the profile in the URL. An example of this value for Salesforce is `Salesforce://AppFlow/example` where `example` is the name of the AppFlow Connector Profile.
* `tags` - (Optional) Tags to apply to the Data Integration. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

A `schedule_config` block supports the following arguments:

* `first_execution_from` - (Required) The start date for objects to import in the first flow run as an Unix/epoch timestamp in milliseconds or in ISO-8601 format. This needs to be a time in the past, meaning that the data created or updated before this given date will not be downloaded.
* `object` - (Required) The name of the object to pull from the data source. Examples of objects in Salesforce include `Case`, `Account`, or `Lead`.
* `schedule_expression` - (Required) How often the data should be pulled from data source. Examples include `rate(1 hour)`, `rate(3 hours)`, `rate(1 day)`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the Data Integration.
* `id` - The identifier of the Data Integration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon AppIntegrations Data Integrations using the `id`. For example:

```terraform
import {
  to = aws_appintegrations_data_integration.example
  id = "12345678-1234-1234-1234-123456789123"
}
```

Using `terraform import`, import Amazon AppIntegrations Data Integrations using the `id`. For example:

```console
% terraform import aws_appintegrations_data_integration.example 12345678-1234-1234-1234-123456789123
```
