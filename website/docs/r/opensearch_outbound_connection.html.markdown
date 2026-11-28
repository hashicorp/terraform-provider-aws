---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_outbound_connection"
description: |-
  Terraform resource for managing an AWS OpenSearch Outbound Connection.
---

# Resource: aws_opensearch_outbound_connection

Manages an AWS Opensearch Outbound Connection.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_opensearch_outbound_connection" "foo" {
  connection_alias = "outbound_connection"
  connection_mode  = "DIRECT"
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
```

## Argument Reference

This resource supports the following arguments:

* `accept_connection` - (Optional, Forces new resource) Accepts the connection.
* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `connection_alias` - (Required, Forces new resource) Connection alias that will be used by the customer for this connection.
* `connection_mode` - (Optional, Forces new resource) Connection mode. Accepted values are `DIRECT` or `VPC_ENDPOINT`.
* `connection_properties` - (Optional, Forces new resource) Configuration block for the outbound connection.
* `local_domain_info` - (Required, Forces new resource) Configuration block for the local Opensearch domain.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `remote_domain_info` - (Required, Forces new resource) Configuration block for the remote Opensearch domain.

### `connection_properties` Block

* `cross_cluster_search` - (Optional, Forces new resource) Configuration block for cross cluster search.

### `cross_cluster_search` Block

* `skip_unavailable` - (Optional, Forces new resource) Skips unavailable clusters and can only be used for cross-cluster searches. Accepted values are `ENABLED` or `DISABLED`.

### `local_domain_info` Block

* `domain_name` - (Required, Forces new resource) Name of the local domain.
* `owner_id` - (Required, Forces new resource) Account ID of the owner of the local domain.
* `region` - (Required, Forces new resource) Region of the local domain.

### `remote_domain_info` Block

* `domain_name` - (Required, Forces new resource) Name of the remote domain.
* `owner_id` - (Required, Forces new resource) Account ID of the owner of the remote domain.
* `region` - (Required, Forces new resource) Region of the remote domain.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `connection_status` - Status of the connection request.
* `id` - ID of the connection.

### `connection_properties` Block

* `endpoint` - Endpoint of the remote domain, is only set when `connection_mode` is `VPC_ENDPOINT` and `accept_connection` is `TRUE`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS Opensearch Outbound Connections using the Outbound Connection ID. For example:

```terraform
import {
  to = aws_opensearch_outbound_connection.foo
  id = "connection-id"
}
```

Using `terraform import`, import AWS Opensearch Outbound Connections using the Outbound Connection ID. For example:

```console
% terraform import aws_opensearch_outbound_connection.foo connection-id
```
