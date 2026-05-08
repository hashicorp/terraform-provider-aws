---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_namespace_registration"
description: |-
  Registers a Redshift namespace to the AWS Glue Data Catalog
---

# Resource: aws_redshift_namespace_registration

Manages an Amazon Redshift namespace registration to the AWS Glue Data Catalog. Use this resource to enable access to a Redshift data warehouse using the Apache Iceberg REST API.

~> **NOTE:** This resource has limited drift detection capabilities. AWS does not provide a reliable API to verify registration status after creation. The resource verifies that the underlying cluster or namespace exists and that an internal data share was created, but cannot detect if the registration was removed outside of Terraform.

## Example Usage

### Serverless Namespace

```terraform
data "aws_caller_identity" "current" {}

resource "aws_redshiftserverless_namespace" "example" {
  namespace_name = "example"
  db_name        = "example"
}

resource "aws_redshiftserverless_workgroup" "example" {
  namespace_name = aws_redshiftserverless_namespace.example.namespace_name
  workgroup_name = "example"
}

resource "aws_redshift_namespace_registration" "example" {
  consumer_identifier             = format("DataCatalog/%s", data.aws_caller_identity.current.account_id)
  namespace_type                  = "serverless"
  serverless_namespace_identifier = aws_redshiftserverless_namespace.example.namespace_name
  serverless_workgroup_identifier = aws_redshiftserverless_workgroup.example.workgroup_name
}
```

### Provisioned Cluster

```terraform
data "aws_caller_identity" "current" {}

resource "aws_redshift_cluster" "example" {
  cluster_identifier = "example"
  database_name      = "example"
  master_username    = "exampleuser"
  master_password    = "ExamplePassword123!"
  node_type          = "dc2.large"
  cluster_type       = "single-node"
}

resource "aws_redshift_namespace_registration" "example" {
  consumer_identifier            = format("DataCatalog/%s", data.aws_caller_identity.current.account_id)
  namespace_type                 = "provisioned"
  provisioned_cluster_identifier = aws_redshift_cluster.example.cluster_identifier
}
```

## Argument Reference

The following arguments are required:

* `consumer_identifier` - (Required, Forces new resource) Consumer identifier for the registration. Typically in the format `DataCatalog/<account-id>`.
* `namespace_type` - (Required, Forces new resource) Type of namespace being registered. Valid values: `serverless`, `provisioned`.

The following arguments are optional:

* `provisioned_cluster_identifier` - (Optional, Forces new resource) Identifier of the provisioned cluster. Required when `namespace_type` is `provisioned`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `serverless_namespace_identifier` - (Optional, Forces new resource) Identifier of the serverless namespace. Required when `namespace_type` is `serverless`. Can be either the namespace name or namespace ID (UUID).
* `serverless_workgroup_identifier` - (Optional, Forces new resource) Identifier of the serverless workgroup. Required when `namespace_type` is `serverless`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

For serverless namespaces:

```terraform
import {
  to = aws_redshift_namespace_registration.example
  identity = {
    "consumer_identifier"             = "DataCatalog/123456789012"
    "namespace_type"                  = "serverless"
    "serverless_namespace_identifier" = "example-namespace"
    "serverless_workgroup_identifier" = "example-workgroup"
  }
}

resource "aws_redshift_namespace_registration" "example" {
  ### Configuration omitted for brevity ###
}
```

For provisioned clusters:

```terraform
import {
  to = aws_redshift_namespace_registration.example
  identity = {
    "consumer_identifier"            = "DataCatalog/123456789012"
    "namespace_type"                 = "provisioned"
    "provisioned_cluster_identifier" = "example-cluster"
  }
}

resource "aws_redshift_namespace_registration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `consumer_identifier` (String) Consumer identifier for the registration.
- `namespace_type` (String) Type of namespace being registered. Valid values: `serverless`, `provisioned`.

#### Optional

- `provisioned_cluster_identifier` (String) Identifier of the provisioned cluster. Required when `namespace_type` is `provisioned`.
- `serverless_namespace_identifier` (String) Identifier of the serverless namespace. Required when `namespace_type` is `serverless`.
- `serverless_workgroup_identifier` (String) Identifier of the serverless workgroup. Required when `namespace_type` is `serverless`.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Namespace Registration using the composite ID. For example:

For serverless namespaces:

```terraform
import {
  to = aws_redshift_namespace_registration.example
  id = "DataCatalog/123456789012,example-namespace,example-workgroup"
}
```

For provisioned clusters:

```terraform
import {
  to = aws_redshift_namespace_registration.example
  id = "DataCatalog/123456789012,example-cluster"
}
```

Using `terraform import`, import Redshift Namespace Registration using the composite ID. For example:

For serverless namespaces:

```console
% terraform import aws_redshift_namespace_registration.example DataCatalog/123456789012,example-namespace,example-workgroup
```

For provisioned clusters:

```console
% terraform import aws_redshift_namespace_registration.example DataCatalog/123456789012,example-cluster
```
