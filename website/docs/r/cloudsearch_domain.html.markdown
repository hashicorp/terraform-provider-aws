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

  index_field {
    name            = "headline"
    type            = "text"
    search          = true
    return          = true
    sort            = true
    highlight       = false
    analysis_scheme = "_en_default_"
  }

  index_field {
    name   = "price"
    type   = "double"
    search = true
    facet  = true
    return = true
    sort   = true

    source_fields = "headline"
  }
}
```

## Argument Reference

The following arguments are supported:

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

* `name` - (Required) A unique name for the field. Field names must begin with a letter and be at least 3 and no more than 64 characters long. The allowed characters are: `a`-`z` (lower-case letters), `0`-`9`, and `_` (underscore). The name `score` is reserved and cannot be used as a field name.
* `type` - (Required) The field type. Valid values: `date`, `date-array`, `double`, `double-array`, `int`, `int-array`, `literal`, `literal-array`, `text`, `text-array`.
* `analysis_scheme` - (Optional) The analysis scheme you want to use for a `text` field. The analysis scheme specifies the language-specific text processing options that are used during indexing.
* `default_value` - (Optional) The default value for the field. This value is used when no value is specified for the field in the document data.
* `facet` - (Optional) You can get facet information by enabling this.
* `highlight` - (Optional) You can highlight information.
* `return` - (Optional) You can enable returning the value of all searchable fields.
* `search` - (Optional) You can set whether this index should be searchable or not.
* `sort` - (Optional) You can enable the property to be sortable.
* `source_fields` - (Optional) A comma-separated list of source fields to map to the field. Specifying a source field copies data from one field to another, enabling you to use the same source data in different ways by configuring different options for the fields.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The domain's ARN.
* `document_service_endpoint` - The service endpoint for updating documents in a search domain.
* `domain_id` - An internally generated unique identifier for the domain.
* `search_service_endpoint` - The service endpoint for requesting search results from a search domain.

## Timeouts

`aws_cloudsearch_domain` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `30 minutes`) How long to wait for the CloudSearch domain to be created.
* `update` - (Default `30 minutes`) How long to wait for the CloudSearch domain to be updated.
* `delete` - (Default `20 minutes`) How long to wait for the CloudSearch domain to be deleted.

## Import

CloudSearch Domains can be imported using the `name`, e.g.,

```
$ terraform import aws_cloudsearch_domain.example example-domain
```
