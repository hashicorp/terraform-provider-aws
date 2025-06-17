---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acls"
description: |-
  Retrieves a list of Web ACL names in a region.
---

# Data Source: aws_wafv2_web_acls

This resource can be useful for getting back a list of Web ACL names for a region.

## Example Usage

The following example retrieves a list of Web ACL names with `REGIONAL` scope.

```terraform
data "aws_wafv2_web_acls" "example" {
  scope = "REGIONAL"
}

output "example" {
  value = data.aws_wafv2_web_acls.example.names
}

# Retrieve Web ACL names that start with "FMManagedWebACLv2"
output "example_managed_by_fm" {
  value = [for acl_name in data.aws_wafv2_web_acls.example.names : acl_name if startswith(acl_name, "FMManagedWebACLv2")]
}
```

## Argument Reference

The following arguments are required:

* `scope` - (Required) The scope of the Web ACLs to be listed. Valid values are `CLOUDFRONT` or `REGIONAL`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `names` - A list of Web ACL names with the specified scope.
