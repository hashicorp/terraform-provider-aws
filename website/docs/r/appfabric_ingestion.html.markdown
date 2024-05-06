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
	app = "OKTA"
	app_bundle_identifier = "[App Bundle ARN]"
	tenant_id = "example.okta.com"
	ingestion_type = "auditLog"
	tags = {
    Environment = "test"
  }
}
```

## Argument Reference

The following arguments are required:

* `app` - (Required) Name of the application.
Valid values are:

  - SLACK
  - ASANA
  - JIRA
  - M365
  - M365AUDITLOGS
  - ZOOM
  - ZENDESK
  - OKTA
  - GOOGLE
  - DROPBOX
  - SMARTSHEET
  - CISCO

* `app_bundle_identifier` - (Required) Amazon Resource Name (ARN) or Universal Unique Identifier (UUID) of the app bundle to use for the request. 

* `ingestion_type` - (Required) Ingestion type. Valid values are `auditLog`.

* `tenant_id` - (Required) ID of the application tenant.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `app_bundle_arn` - Amazon Resource Name (ARN) of the app bundle to use for the request.
* `arn` - ARN of the Ingestion. 
* `state` - Status of the Ingestion. Valid values are: `enabled` or `disabled`

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

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
