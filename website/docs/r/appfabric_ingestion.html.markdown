---
subcategory: "AppFabric"
layout: "aws"
page_title: "AWS: aws_appfabric_ingestion"
description: |-
  Terraform resource for managing an AWS AppFabric Ingestion.
---

# Resource: aws_appfabric_ingestion

Terraform resource for managing an AWS AppFabric Ingestion.

## Example Usage

### Basic Usage

```terraform
resource "aws_appfabric_ingestion" "example" {
  app            = "OKTA"
  app_bundle_arn = aws_appfabric_app_bundle.example.arn
  tenant_id      = "example.okta.com"
  ingestion_type = "auditLog"
  tags = {
    Environment = "test"
  }
}
```

## Argument Reference

The following arguments are required:

* `app` - (Required) Name of the application.
Refer to the AWS Documentation for the [list of valid values](https://docs.aws.amazon.com/appfabric/latest/api/API_CreateIngestion.html#appfabric-CreateIngestion-request-app)
* `app_bundle_arn` - (Required) Amazon Resource Name (ARN) of the app bundle to use for the request.
* `ingestion_type` - (Required) Ingestion type. Valid values are `auditLog`.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tenant_id` - (Required) ID of the application tenant.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Ingestion.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppFabric Ingestion using the `app_bundle_identifier` and `arn` separated by `,`. For example:

```terraform
import {
  to = aws_appfabric_ingestion.example
  id = "arn:aws:appfabric:[region]:[account]:appbundle/a9b91477-8831-43c0-970c-xxxxxxxxxx,arn:aws:appfabric:[region]:[account]:appbundle/a9b91477-8831-43c0-970c-xxxxxxxxxx/ingestion/32251416-710b-4425-96ca-xxxxxxxxxx"
}
```

Using `terraform import`, import AppFabric Ingestion using the `app_bundle_identifier` and `arn` separated by `,`. For example:

```console
% terraform import aws_appfabric_ingestion.example arn:aws:appfabric:[region]:[account]:appbundle/a9b91477-8831-43c0-970c-xxxxxxxxxx,arn:aws:appfabric:[region]:[account]:appbundle/a9b91477-8831-43c0-970c-xxxxxxxxxx/ingestion/32251416-710b-4425-96ca-xxxxxxxxxx
```
