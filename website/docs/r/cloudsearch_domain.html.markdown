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
  name          = "example-domain"
  instance_type = "search.medium"

  index {
    name            = "headline"
    type            = "text"
    search          = true
    return          = true
    sort            = true
    highlight       = false
    analysis_scheme = "_en_default_"
  }

  index {
    name   = "price"
    type   = "double"
    search = true
    facet  = true
    return = true
    sort   = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `endpoint_options` - (Optional) Domain endpoint options. Documented below.
* `multi_az` - (Optional) Whether or not to maintain extra instances for the domain in a second Availability Zone to ensure high availability.
* `name` - (Required) The name of the CloudSearch domain.
* `scaling_parameters` - (Optional) Domain scaling parameters. Documented below.

* `instance_type` - (Optional) The type of instance to start.
* `replication_count` - (Optional) The amount of replicas.
* `partition_count` - (Optional) The amount of partitions on each instance. Currently only supported by `search.2xlarge`.
* `index` - (Required) See [Indices](#indices) below for details.

### endpoint_options

This configuration block supports the following attributes:

* `enforce_https` - (Optional) Enables or disables the requirement that all requests to the domain arrive over HTTPS.
* `tls_security_policy` - (Optional) The minimum required TLS version. See the [AWS documentation](https://docs.aws.amazon.com/cloudsearch/latest/developerguide/API_DomainEndpointOptions.html) for valid values.

### scaling_parameters

This configuration block supports the following attributes:

* `desired_instance_type` - (Optional) The instance type that you want to preconfigure for your domain. See the [AWS documentation](https://docs.aws.amazon.com/cloudsearch/latest/developerguide/API_ScalingParameters.html) for valid values.
* `desired_partition_count` - (Optional) The number of partitions you want to preconfigure for your domain. Only valid when you select `search.2xlarge` as the instance type.
* `desired_replication_count` - (Optional) The number of replicas you want to preconfigure for each index partition.

### Indices

Each of the `index` entities represents an index field of the domain.

* `name` - (Required) Represents the property field name.
* `type` - (Required) Represents the property type. It can be one of `int,int-array,double,double-array,literal,literal-array,text,text-array,date,date-array,lation` [AWS's Docs](http://docs.aws.amazon.com/cloudsearch/latest/developerguide/configuring-index-fields.html)
* `search` - (Required) You can set whether this index should be searchable or not. (index of type text ist always searchable)
* `facet` - (Optional) You can get facet information by enabling this.
* `return` - (Required) You can enable returning the value of all searchable fields.
* `sort` - (Optional) You can enable the property to be sortable.
* `highlight` - (Optional) You can highlight information.
* `analysis_scheme` - (Optional) Only needed with type `text`. [AWS's Docs - supported languages](http://docs.aws.amazon.com/cloudsearch/latest/developerguide/text-processing.html)
* `default_value` - (Optional) The default value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The domain's ARN.
* `document_service_endpoint` - The service endpoint for updating documents in a search domain.
* `domain_id` - An internally generated unique identifier for the domain.
* `search_service_endpoint` - The service endpoint for requesting search results from a search domain.

## Timeouts

`aws_cloudsearch_domain` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `20 minutes`) How long to wait for the CloudSearch domain to be created.
* `update` - (Default `20 minutes`) How long to wait for the CloudSearch domain to be updated.
* `delete` - (Default `20 minutes`) How long to wait for the CloudSearch domain to be deleted.

## Import

CloudSearch Domains can be imported using the `name`, e.g.,

```
$ terraform import aws_cloudsearch_domain.example example-domain
```