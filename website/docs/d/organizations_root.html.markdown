---
layout: "aws"
page_title: "AWS: aws_organizations_root"
sidebar_current: "docs-aws-datasource-organizations-root"
description: |-
    Provides details about the root organizational unit
---

# Data Source: aws_organizations_root

`aws_organizations_root` provides details about the [root organizational unit](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_getting-started_concepts.html#root).

Will give an error if Organizations aren't enabled - see `aws_organizations_organization`.

## Example Usage

```hcl
data "aws_organizations_root" "root" {}
```

## Argument Reference

None.

## Attributes Reference

The following attributes are exported:

* `arn` - The ARN of the organizational unit
* `id` - The ID of the organizational unit (`r-...`)
