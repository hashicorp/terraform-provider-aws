package aws

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/atest"
	awsprovider "github.com/terraform-providers/terraform-provider-aws/provider"
)

func TestAccAwsEc2SpotPriceDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_spot_price.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { atest.PreCheck(t); testAccPreCheckAwsEc2SpotPrice(t) },
		ErrorCheck:   atest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2SpotPriceDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "spot_price", regexp.MustCompile(`^\d+\.\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "spot_price_timestamp", regexp.MustCompile(atest.RFC3339RegexPattern)),
				),
			},
		},
	})
}

func TestAccAwsEc2SpotPriceDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_spot_price.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { atest.PreCheck(t); testAccPreCheckAwsEc2SpotPrice(t) },
		ErrorCheck:   atest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2SpotPriceDataSourceFilterConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "spot_price", regexp.MustCompile(`^\d+\.\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "spot_price_timestamp", regexp.MustCompile(atest.RFC3339RegexPattern)),
				),
			},
		},
	})
}

func testAccPreCheckAwsEc2SpotPrice(t *testing.T) {
	conn := atest.Provider.Meta().(*awsprovider.AWSClient).EC2Conn

	input := &ec2.DescribeSpotPriceHistoryInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeSpotPriceHistory(input)

	if atest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAwsEc2SpotPriceDataSourceConfig() string {
	return atest.ComposeConfig(testAccAvailableAZsNoOptInConfig(), `
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

func testAccAwsEc2SpotPriceDataSourceFilterConfig() string {
	return atest.ComposeConfig(testAccAvailableAZsNoOptInConfig(), `
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
