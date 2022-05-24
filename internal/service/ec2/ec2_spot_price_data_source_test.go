package ec2_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2SpotPriceDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_spot_price.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSpotPrice(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotPriceDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "spot_price", regexp.MustCompile(`^\d+\.\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "spot_price_timestamp", regexp.MustCompile(acctest.RFC3339RegexPattern)),
				),
			},
		},
	})
}

func TestAccEC2SpotPriceDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_spot_price.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSpotPrice(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotPriceDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "spot_price", regexp.MustCompile(`^\d+\.\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "spot_price_timestamp", regexp.MustCompile(acctest.RFC3339RegexPattern)),
				),
			},
		},
	})
}

func testAccPreCheckSpotPrice(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeSpotPriceHistoryInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeSpotPriceHistory(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSpotPriceDataSourceConfig_basic() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
data "aws_region" "current" {}

data "aws_ec2_instance_type_offering" "test" {
  filter {
    name   = "instance-type"
    values = ["m5.xlarge"]
  }
}

data "aws_ec2_spot_price" "test" {
  instance_type = data.aws_ec2_instance_type_offering.test.instance_type

  availability_zone = data.aws_availability_zones.available.names[0]

  filter {
    name   = "product-description"
    values = ["Linux/UNIX"]
  }
}
`)
}

func testAccSpotPriceDataSourceConfig_filter() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
data "aws_region" "current" {}

data "aws_ec2_instance_type_offering" "test" {
  filter {
    name   = "instance-type"
    values = ["m5.xlarge"]
  }
}

data "aws_ec2_spot_price" "test" {
  filter {
    name   = "product-description"
    values = ["Linux/UNIX"]
  }

  filter {
    name   = "instance-type"
    values = [data.aws_ec2_instance_type_offering.test.instance_type]
  }

  filter {
    name   = "availability-zone"
    values = [data.aws_availability_zones.available.names[0]]
  }
}
`)
}
