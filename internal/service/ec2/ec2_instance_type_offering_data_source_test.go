package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2InstanceTypeOfferingDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_type_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstanceTypeOffering(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingFilterDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "instance_type"),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypeOfferingDataSource_locationType(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_type_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstanceTypeOffering(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingLocationTypeDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "instance_type"),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypeOfferingDataSource_preferredInstanceTypes(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_type_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstanceTypeOffering(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingPreferredInstanceTypesDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "instance_type", "t3.micro"),
				),
			},
		},
	})
}

func testAccPreCheckInstanceTypeOffering(t *testing.T) {
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

func testAccInstanceTypeOfferingFilterDataSourceConfig() string {
	return `
# Rather than hardcode an instance type in the testing,
# use the first result from all available offerings.
data "aws_ec2_instance_type_offerings" "test" {}

data "aws_ec2_instance_type_offering" "test" {
  filter {
    name   = "instance-type"
    values = [tolist(data.aws_ec2_instance_type_offerings.test.instance_types)[0]]
  }
}
`
}

func testAccInstanceTypeOfferingLocationTypeDataSourceConfig() string {
	return acctest.ConfigAvailableAZsNoOptIn() + `
# Rather than hardcode an instance type in the testing,
# use the first result from all available offerings.
data "aws_ec2_instance_type_offerings" "test" {
  filter {
    name   = "location"
    values = [data.aws_availability_zones.available.names[0]]
  }

  location_type = "availability-zone"
}

data "aws_ec2_instance_type_offering" "test" {
  filter {
    name   = "instance-type"
    values = [tolist(data.aws_ec2_instance_type_offerings.test.instance_types)[0]]
  }

  filter {
    name   = "location"
    values = [data.aws_availability_zones.available.names[0]]
  }

  location_type = "availability-zone"
}
`
}

func testAccInstanceTypeOfferingPreferredInstanceTypesDataSourceConfig() string {
	return `
data "aws_ec2_instance_type_offering" "test" {
  filter {
    name   = "instance-type"
    values = ["t1.micro", "t2.micro", "t3.micro"]
  }

  preferred_instance_types = ["t3.micro", "t2.micro", "t1.micro"]
}
`
}
