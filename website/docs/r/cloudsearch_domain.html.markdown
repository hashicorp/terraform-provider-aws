---
subcategory: "CloudSearch"
layout: "aws"
page_title: "AWS: aws_cloudsearch_domain"
description: |-
  Provides an CloudSearch domain resource. 
---

# Resource: aws_cloudsearch_domain

Provides an CloudSearch domain resource.

## Example Usage

```terraform
resource "aws_cloudsearch_domain" "example" {
  name          = "test-domain"
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
  access_policies = data.aws_iam_policy_document.cloudsearch_access_policy.json
}

data "aws_iam_policy_document" "cloudsearch_access_policy" {
  statement {
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions = [
      "cloudsearch:search",
      "cloudsearch:suggest"
    ]
  }
}
```

## Argument Reference

The following arguments are supported:

* `access_policies` - (Required) The AWS IAM access policy for the domain. See the [AWS documentation](https://docs.aws.amazon.com/cloudsearch/latest/developerguide/configuring-access.html#cloudsearch-access-policies) for more details.
* `endpoint_options` - (Optional) Domain endpoint options. Documented below.
* `multi_az` - (Optional) Whether or not to maintain extra instances for the domain in a second Availability Zone to ensure high availability.
* `name` - (Required) The name of the CloudSearch domain.
* `scaling_parameters` - (Optional) Domain scaling parameters. Documented below.

* `instance_type` - (Optional) The type of instance to start.
* `replication_count` - (Optional) The amount of replicas.
* `partition_count` - (Optional) The amount of partitions on each instance. Currently only supported by `search.2xlarge`.
* `index` - (Required) See [Indices](#indices) below for details.
* `wait_for_endpoints` - (Optional) - Default true, wait for the search service end point.  If you set this to false, the search and document endpoints won't be available to use as an attribute during the first run.

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

* `document_endpoint` - The doc service end point - see wait_for_endpoints parameter
* `search_endpoint` - The search service end point - see wait_for_endpoints parameter
* `domain_id` - The domain id

## Import

CloudSearch Domains can be imported using the `name`, e.g.,

```
$ terraform import aws_cloudsearch_domain.example test-domain
```