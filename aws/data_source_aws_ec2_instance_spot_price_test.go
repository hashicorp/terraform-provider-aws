package aws

import (
	"fmt"
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
					resource.TestCheckResourceAttrSet(dataSourceName, "instance_type"),
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
data "aws_ec2_instance_spot_price" "test" {
  instance_type = "t3.medium"

  availability_zone = "us-west-2a"

  filter {
    name   = "product-description"
    values = ["Linux/UNIX"]
  }
}
`)
}
