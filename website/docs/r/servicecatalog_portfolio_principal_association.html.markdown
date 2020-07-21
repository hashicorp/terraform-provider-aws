---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_portfolio_principal_association"
description: |-
  Provides a resource to control the association of an IAM Principal with a Service Catalog Portfolio
---

# aws_servicecatalog_portfolio_principal_association

Provides a resource to control the association of an IAM Principal (user, role, group) 
with a Service Catalog Portfolio.

This is necessary for any product in the portfolio to be provisioned.

In most cases this is simple and straightforward. 
However, there are some pathological edge cases that can arise 
as the association is not an identifiable resource in the usual sense. 
For instance if an association between a given portfolio and principal were created twice,
with two instances in Terraform, and then one of them is deleted, 
there would be one instance remaining in Terraform but
the association be absent in AWS until the remaining Terraform instance is re-applied.   


## Example Usage

```hcl
resource "aws_servicecatalog_portfolio_principal_association" "test" {
  portfolio_id  = "port-01234567890abc"
  principal_arn = "arn:aws:iam::0123456789ab:root"
}
```

## Argument Reference

The following arguments are supported:

* `portfolio_id` - (Required) The portfolio identifier
* `principal_arn` - (Required) The ARN of the principal


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A generated ID to represent this association, of the form `${portfolio_id}:${principal_arn}`.


## Import

Service Catalog Portfolio-Product Associations can be imported using the ID constructed 
from the portfolio ID and principal ARN, e.g.

```
$ terraform import aws_servicecatalog_portfolio_principal_association.test port-01234567890abc:arn:aws:iam::0123456789ab:root
```

