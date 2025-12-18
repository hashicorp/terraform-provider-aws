---
subcategory: "CloudSearch"
layout: "aws"
page_title: "AWS: aws_cloudsearch_domain"
description: |-
  Provides an CloudSearch domain resource. 
---

# Resource: aws_cloudsearch_domain

Provides an CloudSearch domain resource.

Terraform waits for the domain to become `Active` when applying a configuration.

## Example Usage

```terraform
resource "aws_cloudsearch_domain" "example" {
  name = "example-domain"

  scaling_parameters {
    desired_instance_type = "search.medium"
  }

  # Text field: supports highlight and analysis_scheme, but NOT facet or search
  # (text fields are always searchable)
  index_field {
    name            = "headline"
    type            = "text"
    return          = true
    highlight       = true
    analysis_scheme = "_en_default_"
  }

  # Double field: supports facet, search, return, sort
  index_field {
    name   = "price"
    type   = "double"
    search = true
    facet  = true
    return = true
    sort   = true
  }

  # Literal field: supports facet, search, return, sort
  index_field {
    name   = "category"
    type   = "literal"
    search = true
    facet  = true
    return = true
    sort   = true
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `endpoint_options` - (Optional) Domain endpoint options. Documented below.
* `index_field` - (Optional) The index fields for documents added to the domain. Documented below.
* `multi_az` - (Optional) Whether or not to maintain extra instances for the domain in a second Availability Zone to ensure high availability.
* `name` - (Required) The name of the CloudSearch domain.
* `scaling_parameters` - (Optional) Domain scaling parameters. Documented below.

### endpoint_options

This configuration block supports the following attributes:

* `enforce_https` - (Optional) Enables or disables the requirement that all requests to the domain arrive over HTTPS.
* `tls_security_policy` - (Optional) The minimum required TLS version. See the [AWS documentation](https://docs.aws.amazon.com/cloudsearch/latest/developerguide/API_DomainEndpointOptions.html) for valid values.

### scaling_parameters

This configuration block supports the following attributes:

* `desired_instance_type` - (Optional) The instance type that you want to preconfigure for your domain. See the [AWS documentation](https://docs.aws.amazon.com/cloudsearch/latest/developerguide/API_ScalingParameters.html) for valid values.
* `desired_partition_count` - (Optional) The number of partitions you want to preconfigure for your domain. Only valid when you select `search.2xlarge` as the instance type.
* `desired_replication_count` - (Optional) The number of replicas you want to preconfigure for each index partition.

### index_field

This configuration block supports the following attributes:

* `name` - (Required) A unique name for the field. Field names must begin with a letter and be at least 1 and no more than 64 characters long. The allowed characters are: `a`-`z` (lower-case letters), `0`-`9`, and `_` (underscore). The name `score` is reserved and cannot be used as a field name.
* `type` - (Required) The field type. Valid values: `date`, `date-array`, `double`, `double-array`, `int`, `int-array`, `literal`, `literal-array`, `text`, `text-array`.
* `analysis_scheme` - (Optional) The analysis scheme you want to use for a `text` field. The analysis scheme specifies the language-specific text processing options that are used during indexing. **Only valid for `text` and `text-array` fields.**
* `default_value` - (Optional) The default value for the field. This value is used when no value is specified for the field in the document data.
* `facet` - (Optional, Default: `false`) Enable faceting for this field. **Cannot be set for `text` or `text-array` fields** (text fields are always searchable and cannot be faceted).
* `highlight` - (Optional, Default: `false`) Enable highlighting for this field. **Only valid for `text` and `text-array` fields.**
* `return` - (Optional, Default: `false`) Enable returning this field's value in search results.
* `search` - (Optional, Default: `false`) Enable this field to be searchable. **Cannot be set for `text` or `text-array` fields** (text fields are always searchable).
* `sort` - (Optional, Default: `false`) Enable sorting by this field. **Cannot be set for array field types** (`int-array`, `double-array`, `literal-array`, `date-array`, `text-array`).
* `source_fields` - (Optional) A comma-separated list of source fields to map to the field. Specifying a source field copies data from one field to another, enabling you to use the same source data in different ways by configuring different options for the fields.

~> **NOTE:** Terraform will return a validation error if you attempt to set attributes that are invalid for your chosen field type, even if set to `false`. Remove invalid attributes from your configuration.

**Boolean Attribute Behavior:**

All boolean attributes (`facet`, `highlight`, `return`, `search`, `sort`) are Optional with `Default: false` in Terraform. When not specified in your configuration, Terraform explicitly sends `false` to AWS CloudSearch.

~> **Important:** There is ambiguity in AWS CloudSearch's behavior for omitted boolean options. AWS documentation states options are "enabled by default" when not specified, but the AWS Console displays them as disabled/unchecked, and the API omits the entire options block when nothing is configured. We have not validated which representation reflects actual runtime behavior. Terraform's approach of explicitly sending `false` ensures consistent, predictable behavior and has been the provider's design since creation.

To explicitly enable features, set them to `true`:

```terraform
index_field {
  name   = "myfield"
  type   = "literal"
  facet  = true
  search = true
}
```

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The domain's ARN.
* `document_service_endpoint` - The service endpoint for updating documents in a search domain.
* `domain_id` - An internally generated unique identifier for the domain.
* `search_service_endpoint` - The service endpoint for requesting search results from a search domain.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudSearch Domains using the `name`. For example:

```terraform
import {
  to = aws_cloudsearch_domain.example
  id = "example-domain"
}
```

Using `terraform import`, import CloudSearch Domains using the `name`. For example:

```console
% terraform import aws_cloudsearch_domain.example example-domain
```
