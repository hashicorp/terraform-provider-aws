package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2InstanceTypeOfferingDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_type_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstanceTypeOfferings(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingDataSourceConfig_filter(),
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
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstanceTypeOfferings(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingDataSourceConfig_location(),
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
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstanceTypeOfferings(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeOfferingDataSourceConfig_preferreds(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "instance_type", "t3.micro"),
				),
			},
		},
	})
}

func testAccInstanceTypeOfferingDataSourceConfig_filter() string {
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

func testAccInstanceTypeOfferingDataSourceConfig_location() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
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
`)
}

func testAccInstanceTypeOfferingDataSourceConfig_preferreds() string {
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
