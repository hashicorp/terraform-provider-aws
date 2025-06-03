---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_router_configuration"
description: |-
  Terraform data source for managing an AWS Direct Connect Router Configuration.
---

# Data Source: aws_dx_router_configuration

Terraform data source for retrieving Router Configuration instructions for a given AWS Direct Connect Virtual Interface and Router Type.

## Example Usage

### Basic Usage

```terraform
data "aws_dx_router_configuration" "example" {
  virtual_interface_id   = "dxvif-abcde123"
  router_type_identifier = "CiscoSystemsInc-2900SeriesRouters-IOS124"
}
```

## Argument Reference

This resource supports the following arguments:

* `virtual_interface_id` - (Required) ID of the Direct Connect Virtual Interface
* `router_type_identifier` - (Required) ID of the Router Type. For example: `CiscoSystemsInc-2900SeriesRouters-IOS124`

There is currently no AWS API to retrieve the full list of `router_type_identifier` values. Here is a list of known `RouterType` objects that can be used:

```json
{
  "routerTypes": [
    {"platform":"2900 Series Routers","routerTypeIdentifier":"CiscoSystemsInc-2900SeriesRouters-IOS124","software":"IOS 12.4+","vendor":"Cisco Systems, Inc.","xsltTemplateName":"customer-router-cisco-generic.xslt","xsltTemplateNameForMacSec":""},
    {"platform":"3700 Series Routers","routerTypeIdentifier":"CiscoSystemsInc-3700SeriesRouters-IOS124","software":"IOS 12.4+","vendor":"Cisco Systems, Inc.","xsltTemplateName":"customer-router-cisco-generic.xslt","xsltTemplateNameForMacSec":""},
    {"platform":"7200 Series Routers","routerTypeIdentifier":"CiscoSystemsInc-7200SeriesRouters-IOS124","software":"IOS 12.4+","vendor":"Cisco Systems, Inc.","xsltTemplateName":"customer-router-cisco-generic.xslt","xsltTemplateNameForMacSec":""},
    {"platform":"Nexus 7000 Series Switches","routerTypeIdentifier":"CiscoSystemsInc-Nexus7000SeriesSwitches-NXOS51","software":"NX-OS 5.1+","vendor":"Cisco Systems, Inc.","xsltTemplateName":"customer-switch-cisco-nexus-generic.xslt","xsltTemplateNameForMacSec":""},
    {"platform":"Nexus 9K+ Series Switches","routerTypeIdentifier":"CiscoSystemsInc-Nexus9KSeriesSwitches-NXOS93","software":"NX-OS 9.3+","vendor":"Cisco Systems, Inc.","xsltTemplateName":"customer-switch-cisco-nexus-generic.xslt","xsltTemplateNameForMacSec":"customer-switch-cisco-nexus-generic-macsec.xslt"},
    {"platform":"M/MX Series Routers","routerTypeIdentifier":"JuniperNetworksInc-MMXSeriesRouters-JunOS95","software":"JunOS 9.5+","vendor":"Juniper Networks, Inc.","xsltTemplateName":"customer-router-juniper-generic.xslt","xsltTemplateNameForMacSec":"customer-router-juniper-generic-macsec.xslt"},
    {"platform":"SRX Series Routers","routerTypeIdentifier":"JuniperNetworksInc-SRXSeriesRouters-JunOS95","software":"JunOS 9.5+","vendor":"Juniper Networks, Inc.","xsltTemplateName":"customer-router-juniper-generic.xslt","xsltTemplateNameForMacSec":""},
    {"platform":"T Series Routers","routerTypeIdentifier":"JuniperNetworksInc-TSeriesRouters-JunOS95","software":"JunOS 9.5+","vendor":"Juniper Networks, Inc.","xsltTemplateName":"customer-router-juniper-generic.xslt","xsltTemplateNameForMacSec":""},
    {"platform":"PA-3000+ and 5000+ series","routerTypeIdentifier":"PaloAltoNetworks-PA3000and5000series-PANOS803","software":"PAN-OS 8.0.3+","vendor":"Palo Alto Networks","xsltTemplateName":"customer-router-palo-alto-generic.xslt","xsltTemplateNameForMacSec":""}]
}
```

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `customer_router_config` - Instructions for configuring your router
* `router` - Block of the router type details

A `router` block supports the following attributes:

* `platform` - Router platform
* `router_type_identifier` - Router type identifier
* `software` - Router operating system
* `vendor` - Router vendor
* `xslt_template_name` - Router XSLT Template Name
* `xslt_template_name_for_mac` - Router XSLT Template Name for MacSec
