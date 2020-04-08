---
subcategory: "Resource Groups Tagging API"
layout: "aws"
page_title: "AWS: aws_resourcegroupstaggingapi_resources"
description: |-
  Provides details about resource tagging.
---

# Data Source: aws_resourcegroupstaggingapi_resources

Provides details about resource tagging.

## Example Usage

### Get All Resource Tag Mappings

```hcl
data "aws_resourcegroupstaggingapi_resources" "test" {}
```

### Filter By Tag Key and Value

```hcl
data "aws_resourcegroupstaggingapi_resources" "test" {
  tag_filters {
    key    = "tag-key"
    values = ["tag-value-1", "tag-value-2"]
  }
}
```


## Argument Reference

The following arguments are supported:

* `exclude_compliant_resources` - (Optional) Specifies whether to exclude resources that are compliant with the tag policy. You can use this parameter only if the `include_compliance_details` argument is also set to `true`.
* `include_compliance_details` - (Optional) Specifies whether to include details regarding the compliance with the effective tag policy.
* `tag_filters` - (Optional) A `tag_filters` block. documented below.
* `resource_type_filter` - (Optional) The constraints on the resources that you want returned. The format of each resource type is `service:resourceType`. For example, specifying a resource type of ec2 returns all Amazon EC2 resources (which includes EC2 instances). Specifying a resource type of `ec2:instance` returns only EC2 instances.

### Tag Filters

A `tag_filters` block supports the following arguments:

If you do specify `tag_filters`, the response returns only those resources that are currently associated with the specified tag.
If you don't specify a `tag_filters`, the response includes all resources that were ever associated with tags. Resources that currently don't have associated tags are shown with an empty tag set.

* `key` - (Required) One part of a key-value pair that makes up a tag.
* `values` - (Optional) The optional part of a key-value pair that make up a tag. 

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `resource_tag_mapping_list` -  An `resource_tag_mapping_list` block. documented below.

### Resource Tag Mapping List

A `resource_tag_mapping_list` block supports the following attributes:

* `resource_arn` - The ARN of the resource.
* `compliance_details` - Information that shows whether a resource is compliant with the effective tag policy, including details on any noncompliant tag keys. Documented below.
* `tags` - tags assigned to the resource.


