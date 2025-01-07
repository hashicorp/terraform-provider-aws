---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_authorize_vpc_endpoint_access"
description: |-
  Terraform resource for managing an AWS OpenSearch Authorize Vpc Endpoint Access.
---

# Resource: aws_opensearch_authorize_vpc_endpoint_access

Terraform resource for managing an AWS OpenSearch Authorize Vpc Endpoint Access.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}

resource "aws_opensearch_authorize_vpc_endpoint_access" "test" {
  domain_name = aws_opensearch_domain.test.domain_name
  account     = data.aws_caller_identity.current.account_id
}
```

## Argument Reference

The following arguments are required:

* `account` - (Required) AWS account ID to grant access to.
* `domain_name` - (Required) Name of OpenSearch Service domain to provide access to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `authorized_principal` - Information about the Amazon Web Services account or service that was provided access to the domain. See [authorized principal](#authorized_principal) attribute for further details.

### authorized_principal

* `principal` - IAM principal that is allowed to access to the domain.
* `principal_type` - Type of principal.

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
