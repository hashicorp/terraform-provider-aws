---
layout: "aws"
page_title: "AWS: aws_pricing_product"
sidebar_current: "docs-aws-datasource-pricing-product"
description: |-
  Get information regarding the pricing of an Amazon product
---

# Data Source: aws_pricing_product

Use this data source to get the pricing information of all products in AWS.
This data source is only available in a us-east-1 or ap-south-1 provider.

## Example Usage

```hcl
data "aws_pricing_product" "test1" {
    service_code = "AmazonEC2"
    
    filters = [
        {
        field = "instanceType"
        value = "c5.xlarge"
        },
        {
        field = "operatingSystem"
        value = "Linux"
        },
        {
        field = "location"
        value = "US East (N. Virginia)"
        },
        {
        field = "preInstalledSw"
        value = "NA"
        },
        {
        field = "licenseModel"
        value = "No License required"
        },
        {
        field = "tenancy"
        value = "Shared"
        },
    ]
    
    json_query = "terms.OnDemand.*.priceDimensions.*.pricePerUnit.USD"
}

data "aws_pricing_product" "test2" {
    service_code = "AmazonRedshift"
    
    filters = [
        {
        field = "instanceType"
        value = "ds1.xlarge"
        },
        {
        field = "location"
        value = "US East (N. Virginia)"
        },
    ]
    
    json_query = "terms.OnDemand.*.priceDimensions.*.pricePerUnit.USD"
}
```

## Argument Reference

 * `service_code` - (Required) The code of the service. Available service codes can be fetched using the DescribeServices pricing API call.
 * `filters` - (Required) A list of filters. Passed directly to the API (see GetProducts API reference). These filters must describe a single product, this resource will fail if more than one product is returned by the API.
 * `json_query` - (Required) The JSON query used to fetch the wanted value. In a [GJSON format](https://github.com/tidwall/gjson).

## Attributes Reference

 * `query_result` - Set to the result of the JSON query applied on the product returned from the API.
