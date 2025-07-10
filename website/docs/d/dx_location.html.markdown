---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_location"
description: |-
  Retrieve information about a specific AWS Direct Connect location in the current AWS Region.
---

# Data Source: aws_dx_location

Retrieve information about a specific AWS Direct Connect location in the current AWS Region.
These are the locations that can be specified when configuring [`aws_dx_connection`](/docs/providers/aws/r/dx_connection.html) or [`aws_dx_lag`](/docs/providers/aws/r/dx_lag.html) resources.

~> **Note:** This data source is different from the [`aws_dx_locations`](/docs/providers/aws/d/dx_locations.html) data source which retrieves information about all the AWS Direct Connect locations in the current AWS Region.

## Example Usage

```terraform
data "aws_dx_location" "example" {
  location_code = "CS32A-24FL"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `location_code` - (Required) Code for the location to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `available_macsec_port_speeds` - The available MAC Security (MACsec) port speeds for the location.
* `available_port_speeds` - The available port speeds for the location.
* `available_providers` - Names of the service providers for the location.
* `location_name` - Name of the location. This includes the name of the colocation partner and the physical site of the building.
