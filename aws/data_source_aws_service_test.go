package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSService_basic(t *testing.T) {
	dataSourceName := "data.aws_service.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "dns", fmt.Sprintf("%s.%s.%s", ec2.ServiceID, testAccGetRegion(), "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", testAccGetRegion(), ec2.ServiceID)),
					resource.TestCheckResourceAttr(dataSourceName, "partition", testAccGetPartition()),
					resource.TestCheckResourceAttr(dataSourceName, "prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns", fmt.Sprintf("%s.%s.%s", "com.amazonaws", testAccGetRegion(), ec2.ServiceID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", ec2.ServiceID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "true"),
				),
			},
		},
	})
}

func TestAccAWSService_byName(t *testing.T) {
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceConfig_byServiceName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "dns", fmt.Sprintf("%s.%s.%s", s3.ServiceID, endpoints.CnNorth1RegionID, "amazonaws.com.cn")),
					resource.TestCheckResourceAttr(dataSourceName, "name", fmt.Sprintf("%s.%s.%s", "cn.com.amazonaws", endpoints.CnNorth1RegionID, s3.ServiceID)),
					resource.TestCheckResourceAttr(dataSourceName, "partition", endpoints.AwsCnPartitionID),
					resource.TestCheckResourceAttr(dataSourceName, "prefix", "cn.com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "region", endpoints.CnNorth1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", s3.ServiceID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "true"),
				),
			},
		},
	})
}

func TestAccAWSService_byPart(t *testing.T) {
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceConfig_byPart(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", testAccGetRegion(), ec2.ServiceID)),
					resource.TestCheckResourceAttr(dataSourceName, "prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", ec2.ServiceID),
				),
			},
		},
	})
}

func TestAccAWSService_unsupported(t *testing.T) {
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceConfig_unsupported(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "dns", fmt.Sprintf("%s.%s.%s", waf.EndpointsID, endpoints.UsGovWest1RegionID, "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", endpoints.UsGovWest1RegionID, waf.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "partition", endpoints.AwsUsGovPartitionID),
					resource.TestCheckResourceAttr(dataSourceName, "prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "region", endpoints.UsGovWest1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", waf.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "false"),
				),
			},
		},
	})
}

func testAccCheckAwsServiceConfig_basic() string {
	return fmt.Sprintf(`
data "aws_service" "default" {}
`)
}

func testAccCheckAwsServiceConfig_byServiceName() string {
	// lintignore:AWSAT003
	return fmt.Sprintf(`
data "aws_service" "test" {
  name = "cn.com.amazonaws.cn-north-1.s3"
}
`)
}

func testAccCheckAwsServiceConfig_byPart() string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_service" "test" {
  prefix     = "com.amazonaws"
  region     = data.aws_region.current.name
  service_id = "ec2"
}
`)
}

func testAccCheckAwsServiceConfig_unsupported() string {
	return fmt.Sprintf(`
data "aws_service" "test" {
  name = "com.amazonaws.us-gov-west-1.waf"
}
`)
}
