package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2InstanceTypeOfferingsDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstanceTypeOfferings(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingsFilterDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceTypeOfferingsInstanceTypes(dataSourceName),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypeOfferingsDataSource_locationType(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstanceTypeOfferings(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingsLocationTypeDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceTypeOfferingsInstanceTypes(dataSourceName),
					testAccCheckInstanceTypeOfferingsLocations(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckInstanceTypeOfferingsInstanceTypes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["instance_types.#"]; v == "0" {
			return fmt.Errorf("expected at least one instance_types result, got none")
		}

		if v := rs.Primary.Attributes["locations.#"]; v == "0" {
			return fmt.Errorf("expected at least one locations result, got none")
		}

		if v := rs.Primary.Attributes["location_types.#"]; v == "0" {
			return fmt.Errorf("expected at least one location_types result, got none")
		}

		return nil
	}
}

func testAccCheckInstanceTypeOfferingsLocations(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["locations.#"]; v == "0" {
			return fmt.Errorf("expected at least one locations result, got none")
		}

		return nil
	}
}

func testAccPreCheckInstanceTypeOfferings(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeInstanceTypeOfferingsInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeInstanceTypeOfferings(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccInstanceTypeOfferingsFilterDataSourceConfig() string {
	return `
data "aws_ec2_instance_type_offerings" "test" {
  filter {
    name   = "instance-type"
    values = ["t2.micro", "t3.micro"]
  }
}
`
}

func testAccInstanceTypeOfferingsLocationTypeDataSourceConfig() string {
	return acctest.ConfigAvailableAZsNoOptIn() + `
data "aws_ec2_instance_type_offerings" "test" {
  filter {
    name   = "location"
    values = [data.aws_availability_zones.available.names[0]]
  }

  location_type = "availability-zone"
}
`
}
