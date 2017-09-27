---
layout: "aws"
page_title: "AWS: aws_cloudsearch_domain"
sidebar_current: "docs-aws-resource-cloudsearch"
description: |-
  Provides an CloudSearch domain resource.
---

# aws_cloudsearch_domain

Provides an CloudSearch domain resource.

## Example Usage

```hcl
resource "aws_cloudsearch_domain" "my_domain" {
  domain_name   = "test-domain"
  instance_type = "search.m3.medium"

  indexes = [
    {
      name            = "headline"
      type            = "text"
      search          = true
      return          = true
      sort            = true
      highlight       = false
      analysis_scheme = "_de_default_"
    },
    {
      name   = "price"
      type   = "double"
      search = true
      facet  = true
      return = true
      sort   = true
    },
  ]

  access_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "*"
        ]
      },
      "Action": [
        "cloudsearch:*"
      ]
    }
  ]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) Domain name
* `instance_type` - (Optional) The type of instance to start
* `replication_count` - (Optional) The amount of replicas.
* `partition_count` - (Optional) The amount of partitions on each instance. Currently only supported by `search.m3.2xlarge`
* `indexes` - (Required) See [Indexes](#indexes) below for details.
* `access_policy` - (Required) The iam access policy.

### Indexes

Each of the `indexes` entities represents an index field of the domain.

* `name` - (Required) Represents the property field name.
* `type` - (Required) Represents the property type. It can be one of `int,int-array,double,double-array,literal,literal-array,text,text-array,date,date-array,lation` [AWS's Docs](http://docs.aws.amazon.com/cloudsearch/latest/developerguide/configuring-index-fields.html)
* `search` - (Required) You can set whether this index should be searchable or not. (index of type text ist always searchable)
* `facet` - (Optional) You can get facet information by enabling this.
* `return` - (Required) You can enable returning the value of all searchable fields.
* `sort` - (Optional) You can enable the property to be sortable.
* `highlight` - (Optional) You can highlight information.
* `analysis_scheme` - (Optional) Only needed with type `text`. [AWS's Docs - supported languages](http://docs.aws.amazon.com/cloudsearch/latest/developerguide/text-processing.html)
* `default_value` - (Optional) The default value.