---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_inbound_connection_accepter"
description: |-
  Terraform resource for managing an AWS OpenSearch Inbound Connection Accepter.
---

# Resource: aws_opensearch_inbound_connection_accepter

Manages an [AWS Opensearch Inbound Connection Accepter](https://docs.aws.amazon.com/opensearch-service/latest/APIReference/API_AcceptInboundConnection.html). If connecting domains from different AWS accounts, ensure that the accepter is configured to use the AWS account where the _remote_ opensearch domain exists.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_opensearch_outbound_connection" "foo" {
  connection_alias = "outbound_connection"
  local_domain_info {
    owner_id    = data.aws_caller_identity.current.account_id
    region      = data.aws_region.current.region
    domain_name = aws_opensearch_domain.local_domain.domain_name
  }

  remote_domain_info {
    owner_id    = data.aws_caller_identity.current.account_id
    region      = data.aws_region.current.region
    domain_name = aws_opensearch_domain.remote_domain.domain_name
  }
}

resource "aws_opensearch_inbound_connection_accepter" "foo" {
  connection_id = aws_opensearch_outbound_connection.foo.id
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `connection_id` - (Required, Forces new resource) ID of the connection to accept.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `connection_status` - Status of the connection request.
* `id` - ID of the connection to accept.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS Opensearch Inbound Connection Accepters using the Inbound Connection ID. For example:

```terraform
import {
  to = aws_opensearch_inbound_connection_accepter.foo
  id = "connection-id"
}
```

Using `terraform import`, import AWS Opensearch Inbound Connection Accepters using the Inbound Connection ID. For example:

```console
% terraform import aws_opensearch_inbound_connection_accepter.foo connection-id
```
