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

### Authorize an AWS Account

```terraform
data "aws_caller_identity" "current" {}

resource "aws_opensearch_authorize_vpc_endpoint_access" "test" {
  domain_name = aws_opensearch_domain.test.domain_name
  account     = data.aws_caller_identity.current.account_id
}
```

### Authorize the OpenSearch Service Principal (Dashboard)

```terraform
resource "aws_opensearch_authorize_vpc_endpoint_access" "dashboard" {
  domain_name = aws_opensearch_domain.test.domain_name
  service     = "application.opensearchservice.amazonaws.com"
}
```

### Authorize the OpenSearch Service Principal in Specific Regions

```terraform
resource "aws_opensearch_authorize_vpc_endpoint_access" "dashboard" {
  domain_name = aws_opensearch_domain.test.domain_name
  service     = "application.opensearchservice.amazonaws.com"

  service_options {
    supported_regions = ["us-east-1", "us-west-2"]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `domain_name` - (Required) Name of OpenSearch Service domain to provide access to.
* `account` - (Optional) AWS account ID to grant access to. Exactly one of `account` or `service` must be specified.
* `service` - (Optional) AWS service principal to grant access to. Currently only `application.opensearchservice.amazonaws.com` (the OpenSearch Dashboard) is supported. Exactly one of `account` or `service` must be specified.
* `service_options` - (Optional) Options for the service principal. Only valid when `service` is set. See [service_options](#service_options) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### service_options

* `supported_regions` - (Optional) Set of Regions in which the service principal is allowed to use the endpoint access.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `authorized_principal` - Information about the Amazon Web Services account or service that was provided access to the domain. See [authorized principal](#authorized_principal) attribute for further details.

### authorized_principal

* `principal` - IAM principal that is allowed to access to the domain.
* `principal_type` - Type of principal.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch Authorize Vpc Endpoint Access using the `domain_name` and the principal (either an AWS account ID or a service principal) separated by a comma (,). For example:

```terraform
import {
  to = aws_opensearch_authorize_vpc_endpoint_access.example
  id = "authorize_vpc_endpoint_access-id-12345678,123456789012"
}
```

To import an authorization granted to a service principal:

```terraform
import {
  to = aws_opensearch_authorize_vpc_endpoint_access.dashboard
  id = "authorize_vpc_endpoint_access-id-12345678,application.opensearchservice.amazonaws.com"
}
```

Using `terraform import`, import OpenSearch Authorize Vpc Endpoint Access using the `domain_name` and the principal separated by a comma (,). For example:

```console
% terraform import aws_opensearch_authorize_vpc_endpoint_access.example authorize_vpc_endpoint_access-id-12345678,123456789012
```
