---
subcategory: "Elasticsearch"
layout: "aws"
page_title: "AWS: aws_elasticsearch_domain_saml_options"
description: |-
  Terraform resource for managing SAML authentication options for an AWS Elasticsearch Domain.
---

# Resource: aws_elasticsearch_domain_saml_options

Manages SAML authentication options for an AWS Elasticsearch Domain.

## Example Usage

### Basic Usage

```terraform
resource "aws_elasticsearch_domain" "example" {
  domain_name           = "example"
  elasticsearch_version = "1.5"

  cluster_config {
    instance_type = "r4.large.elasticsearch"
  }

  snapshot_options {
    automated_snapshot_start_hour = 23
  }

  tags = {
    Domain = "TestDomain"
  }
}

resource "aws_elasticsearch_domain_saml_options" "example" {
  domain_name = aws_elasticsearch_domain.example.domain_name
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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `saml_options` - (Optional) The SAML authentication options for an AWS Elasticsearch Domain.

### saml_options

* `enabled` - (Required) Whether SAML authentication is enabled.
* `idp` - (Optional) Information from your identity provider.
* `master_backend_role` - (Optional) This backend role from the SAML IdP receives full permissions to the cluster, equivalent to a new master user.
* `master_user_name` - (Optional) This username from the SAML IdP receives full permissions to the cluster, equivalent to a new master user.
* `roles_key` - (Optional) Element of the SAML assertion to use for backend roles. Default is roles.
* `session_timeout_minutes` - (Optional) Duration of a session in minutes after a user logs in. Default is 60. Maximum value is 1,440.
* `subject_key` - (Optional) Custom SAML attribute to use for user names. Default is an empty string - `""`. This will cause Elasticsearch to use the `NameID` element of the `Subject`, which is the default location for name identifiers in the SAML specification.

#### idp

* `entity_id` - (Required) The unique Entity ID of the application in SAML Identity Provider.
* `metadata_content` - (Required) The Metadata of the SAML application in xml format.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the domain the SAML options are associated with.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Elasticsearch domains using the `domain_name`. For example:

```terraform
import {
  to = aws_elasticsearch_domain_saml_options.example
  id = "domain_name"
}
```

Using `terraform import`, import Elasticsearch domains using the `domain_name`. For example:

```console
% terraform import aws_elasticsearch_domain_saml_options.example domain_name
```
