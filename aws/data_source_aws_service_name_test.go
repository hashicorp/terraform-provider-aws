package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSServiceName_basic(t *testing.T) {
	dataSourceName := "data.aws_service_name.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceNameConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttr(dataSourceName, "service", endpoints.Ec2ServiceID),
					resource.TestCheckResourceAttr(dataSourceName, "name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", testAccGetRegion(), endpoints.Ec2ServiceID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_prefix", "com.amazonaws"),
				),
			},
		},
	})
}

func TestAccAWSServiceName_byServiceName(t *testing.T) {
	dataSourceName := "data.aws_service_name.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceNameConfig_byServiceName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "region", endpoints.ApNortheast1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "service", endpoints.S3ServiceID),
					resource.TestCheckResourceAttr(dataSourceName, "name", fmt.Sprintf("%s.%s.%s", "cn.com.amazonaws", endpoints.ApNortheast1RegionID, endpoints.S3ServiceID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_prefix", "cn.com.amazonaws"),
				),
			},
		},
	})
}

func TestAccAWSServiceName_byPart(t *testing.T) {
	dataSourceName := "data.aws_service_name.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceNameConfig_byPart(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttr(dataSourceName, "service", endpoints.Ec2ServiceID),
					resource.TestCheckResourceAttr(dataSourceName, "name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", testAccGetRegion(), endpoints.Ec2ServiceID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_prefix", "com.amazonaws"),
				),
			},
		},
	})
}

func testAccCheckAwsServiceNameConfig_basic() string {
	return fmt.Sprintf(`
data "aws_service_name" "default" {}
`)
}

func testAccCheckAwsServiceNameConfig_byServiceName() string {
	// lintignore:AWSAT003
	return fmt.Sprintf(`
data "aws_service_name" "test" {
  name = "cn.com.amazonaws.ap-northeast-1.s3"
}
`)
}

func testAccCheckAwsServiceNameConfig_byPart() string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_service_name" "test" {
  service        = "ec2"
  region         = data.aws_region.current.name
  service_prefix = "com.amazonaws"
}
`)
}
