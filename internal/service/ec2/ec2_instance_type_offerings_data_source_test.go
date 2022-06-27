package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				Config: testAccInstanceTypeOfferingsDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "instance_types.#", "0"),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "locations.#", "0"),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "location_types.#", "0"),
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
				Config: testAccInstanceTypeOfferingsDataSourceConfig_location(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "instance_types.#", "0"),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "locations.#", "0"),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "location_types.#", "0"),
				),
			},
		},
	})
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

func testAccInstanceTypeOfferingsDataSourceConfig_filter() string {
	return `
data "aws_ec2_instance_type_offerings" "test" {
  filter {
    name   = "instance-type"
    values = ["t2.micro", "t3.micro"]
  }
}
`
}

func testAccInstanceTypeOfferingsDataSourceConfig_location() string {
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
