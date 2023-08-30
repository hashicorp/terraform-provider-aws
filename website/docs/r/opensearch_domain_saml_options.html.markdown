---
subcategory: "OpenSearch"
layout: "aws"
page_title: "AWS: aws_opensearch_domain_saml_options"
description: |-
  Terraform resource for managing SAML authentication options for an AWS OpenSearch Domain.
---

# Resource: aws_opensearch_domain_saml_options

Manages SAML authentication options for an AWS OpenSearch Domain.

## Example Usage

### Basic Usage

```terraform
resource "aws_opensearch_domain" "example" {
  domain_name    = "example"
  engine_version = "OpenSearch_1.1"

  cluster_config {
    instance_type = "r4.large.search"
  }

  snapshot_options {
    automated_snapshot_start_hour = 23
  }

  tags = {
    Domain = "TestDomain"
  }
}

resource "aws_opensearch_domain_saml_options" "example" {
  domain_name = aws_opensearch_domain.example.domain_name
  saml_options {
    enabled = true
    idp {
      entity_id        = "https://example.com"
      metadata_content = file("./saml-metadata.xml")
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `domain_name` - (Required) Name of the domain.

The following arguments are optional:

* `saml_options` - (Optional) SAML authentication options for an AWS OpenSearch Domain.

### saml_options

* `enabled` - (Required) Whether SAML authentication is enabled.
* `idp` - (Optional) Information from your identity provider.
* `master_backend_role` - (Optional) This backend role from the SAML IdP receives full permissions to the cluster, equivalent to a new master user.
* `master_user_name` - (Optional) This username from the SAML IdP receives full permissions to the cluster, equivalent to a new master user.
* `roles_key` - (Optional) Element of the SAML assertion to use for backend roles. Default is roles.
* `session_timeout_minutes` - (Optional) Duration of a session in minutes after a user logs in. Default is 60. Maximum value is 1,440.
* `subject_key` - (Optional) Element of the SAML assertion to use for username. Default is NameID.

#### idp

* `entity_id` - (Required) Unique Entity ID of the application in SAML Identity Provider.
* `metadata_content` - (Required) Metadata of the SAML application in xml format.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Name of the domain the SAML options are associated with.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch domains using the `domain_name`. For example:

```terraform
import {
  to = aws_opensearch_domain_saml_options.example
  id = "domain_name"
}
```

Using `terraform import`, import OpenSearch domains using the `domain_name`. For example:

```console
% terraform import aws_opensearch_domain_saml_options.example domain_name
```
