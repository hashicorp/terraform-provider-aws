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
  local_domain_info {
    owner_id    = data.aws_caller_identity.current.account_id
    region      = data.aws_region.current.name
    domain_name = aws_opensearch_domain.local_domain.domain_name
  }

  remote_domain_info {
    owner_id    = data.aws_caller_identity.current.account_id
    region      = data.aws_region.current.name
    domain_name = aws_opensearch_domain.remote_domain.domain_name
  }
}
```

## Argument Reference

The following arguments are supported:

* `connection_alias` - (Required, Forces new resource) Specifies the connection alias that will be used by the customer for this connection.
* `local_domain_info` - (Required, Forces new resource) Configuration block for the local Opensearch domain.
* `remote_domain_info` - (Required, Forces new resource) Configuration block for the remote Opensearch domain.

### local_domain_info

* `owner_id` - (Required, Forces new resource) The Account ID of the owner of the local domain.
* `domain_name` - (Required, Forces new resource) The name of the local domain.
* `region` - (Required, Forces new resource) The region of the local domain.

### remote_domain_info

* `owner_id` - (Required, Forces new resource) The Account ID of the owner of the remote domain.
* `domain_name` - (Required, Forces new resource) The name of the remote domain.
* `region` - (Required, Forces new resource) The region of the remote domain.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Id of the connection.
* `connection_status` - Status of the connection request.

## Import

AWS Opensearch Outbound Connections can be imported by using the Outbound Connection ID, e.g.,

```
$ terraform import aws_opensearch_outbound_connection.foo connection-id
```
