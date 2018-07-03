package aws

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsPricingProduct_ec2(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsPricingProductConfigEc2("test1", "c5.large") + testAccDataSourceAwsPricingProductConfigEc2("test2", "c5.xlarge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_pricing_product.test1", "query_result"),
					resource.TestCheckResourceAttrSet("data.aws_pricing_product.test2", "query_result"),
					testAccPricingCheckValueIsFloat("data.aws_pricing_product.test1"),
					testAccPricingCheckValueIsFloat("data.aws_pricing_product.test2"),
					testAccPricingCheckGreaterValue("data.aws_pricing_product.test2", "data.aws_pricing_product.test1"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsPricingProduct_redshift(t *testing.T) {
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsPricingProductConfigRedshift(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_pricing_product.test", "query_result"),
					testAccPricingCheckValueIsFloat("data.aws_pricing_product.test"),
				),
			},
		},
	})
}

func testAccDataSourceAwsPricingProductConfigEc2(dataName string, instanceType string) string {
	return fmt.Sprintf(`data "aws_pricing_product" "%s" {
		service_code = "AmazonEC2"
	  
		filters = [
		  {
			field = "instanceType"
			value = "%s"
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
`, dataName, instanceType)
}

func testAccDataSourceAwsPricingProductConfigRedshift() string {
	return fmt.Sprintf(`data "aws_pricing_product" "test" {
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
`)
}

func testAccPricingCheckValueIsFloat(data string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[data]

		if !ok {
			return fmt.Errorf("Can't find resource: %s", data)
		}

		queryResult := rs.Primary.Attributes["query_result"]
		if _, err := strconv.ParseFloat(queryResult, 32); err != nil {
			return fmt.Errorf("%s query_result value (%s) is not a float: %s", data, queryResult, err)
		}

		return nil
	}
}

func testAccPricingCheckGreaterValue(dataWithGreaterValue string, otherData string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		greaterResource, ok := s.RootModule().Resources[dataWithGreaterValue]
		if !ok {
			return fmt.Errorf("Can't find resource: %s", dataWithGreaterValue)
		}

		lesserResource, ok := s.RootModule().Resources[otherData]
		if !ok {
			return fmt.Errorf("Can't find resource: %s", otherData)
		}

		greaterValue := greaterResource.Primary.Attributes["query_result"]
		lesserValue := lesserResource.Primary.Attributes["query_result"]

		if greaterValue <= lesserValue {
			return fmt.Errorf("%s (%s) has a greater value than %s (%s). Should have been the opposite", otherData, lesserValue, dataWithGreaterValue, greaterValue)
		}

		return nil
	}
}
