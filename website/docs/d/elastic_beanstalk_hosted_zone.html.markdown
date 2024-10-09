---
subcategory: "Elastic Beanstalk"
layout: "aws"
page_title: "AWS: aws_elastic_beanstalk_hosted_zone"
description: |-
  Get an elastic beanstalk hosted zone.
---

# Data Source: aws_elastic_beanstalk_hosted_zone

Use this data source to get the ID of an [elastic beanstalk hosted zone](http://docs.aws.amazon.com/general/latest/gr/rande.html#elasticbeanstalk_region).

## Example Usage

```terraform
data "aws_elastic_beanstalk_hosted_zone" "current" {}
```

## Argument Reference

* `region` - (Optional) Region you'd like the zone for. By default, fetches the current region.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the hosted zone.

* `region` - Region of the hosted zone.
