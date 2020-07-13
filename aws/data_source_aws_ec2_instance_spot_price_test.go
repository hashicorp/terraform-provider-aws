package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSEc2InstanceSpotPriceDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_spot_price.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2InstanceSpotPrice(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2InstanceSpotPriceDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "spot_price", regexp.MustCompile(`^\d+\.\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "spot_price_timestamp", regexp.MustCompile(rfc3339RegexPattern)),
				),
			},
		},
	})
}

func testAccPreCheckAWSEc2InstanceSpotPrice(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	input := &ec2.DescribeSpotPriceHistoryInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeSpotPriceHistory(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSEc2InstanceSpotPriceDataSourceConfig() string {
	return fmt.Sprintf(`
# Rather than hardcode an instance type in the testing,
# use the first result from all available offerings.
data "aws_ec2_instance_type_offerings" "test" {}

data "aws_ec2_instance_type_offering" "test" {
  filter {
    name   = "instance-type"
	values = [tolist(data.aws_ec2_instance_type_offerings.test.instance_types)[0]]
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ec2_instance_spot_price" "test" {
  instance_type = data.aws_ec2_instance_type_offering.test.instance_type

  availability_zone = data.aws_availability_zones.available.names[0]

  filter {
    name   = "product-description"
    values = ["Linux/UNIX"]
  }
}
`)
}
