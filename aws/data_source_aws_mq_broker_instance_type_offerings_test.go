package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/mq"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsMqInstanceBrokerTypeOfferingsDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_mq_broker_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2InstanceTypeOfferings(t) },
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqInstanceBrokerTypeOfferingsInstanceTypes(dataSourceName),
				),
			},
		},
	})
}

func TestAccAwsMqInstanceBrokerTypeOfferingsDataSource_LocationType(t *testing.T) {
	dataSourceName := "data.aws_mq_broker_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsMqInstanceBrokerTypeOfferings(t) },
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigLocationType(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqInstanceBrokerTypeOfferingsInstanceTypes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckAwsMqInstanceBrokerTypeOfferingsInstanceTypes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["instance_types.#"]; v == "0" {
			return fmt.Errorf("expected at least one instance_types result, got none")
		}

		return nil
	}
}

func testAccPreCheckAwsMqInstanceBrokerTypeOfferings(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).mqconn

	input := &mq.DescribeBrokerInstanceOptionsInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeBrokerInstanceOptions(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigFilter() string {
	return `
data "aws_mq_broker_instance_type_offerings" "test" {
  filter {
    name   = "instance-type"
    values = ["t2.micro", "t3.micro"]
  }
}
`
}

func testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigLocationType() string {
	return testAccAvailableAZsNoOptInConfig() + `
data "aws_mq_broker_instance_type_offerings" "test" {
  filter {
    name   = "location"
    values = [data.aws_availability_zones.available.names[0]]
  }

  location_type = "availability-zone"
}
`
}
