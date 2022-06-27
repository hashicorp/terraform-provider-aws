package meta_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
)

func TestInvertStringSlice(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    []string
		Expected []string
	}{
		{
			Name:     "DNS Suffix",
			Input:    []string{"amazonaws", "com", "cn"},
			Expected: []string{"cn", "com", "amazonaws"},
		},
		{
			Name:     "Ordered List",
			Input:    []string{"abc", "bcd", "cde", "xyz", "zzz"},
			Expected: []string{"zzz", "xyz", "cde", "bcd", "abc"},
		},
		{
			Name:     "Unordered List",
			Input:    []string{"abc", "zzz", "bcd", "xyz", "cde"},
			Expected: []string{"cde", "xyz", "bcd", "zzz", "abc"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			if !reflect.DeepEqual(tfmeta.InvertStringSlice(testCase.Input), testCase.Expected) {
				t.Errorf("got %v, expected %v", tfmeta.InvertStringSlice(testCase.Input), testCase.Expected)
			}
		})
	}
}

func TestAccMetaService_basic(t *testing.T) {
	dataSourceName := "data.aws_service.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "dns_name", fmt.Sprintf("%s.%s.%s", ec2.EndpointsID, acctest.Region(), "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "partition", acctest.Partition()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "region", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", acctest.Region(), ec2.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", ec2.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "true"),
				),
			},
		},
	})
}

func TestAccMetaService_byReverseDNSName(t *testing.T) {
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_byReverseDNSName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "region", endpoints.CnNorth1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "cn.com.amazonaws", endpoints.CnNorth1RegionID, s3.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "cn.com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", s3.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "true"),
				),
			},
		},
	})
}

func TestAccMetaService_byDNSName(t *testing.T) {
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_byDNSName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "region", endpoints.UsEast1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", endpoints.UsEast1RegionID, rds.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", rds.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "true"),
				),
			},
		},
	})
}

func TestAccMetaService_byParts(t *testing.T) {
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_byPart(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "dns_name", fmt.Sprintf("%s.%s.%s", s3.EndpointsID, acctest.Region(), "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", acctest.Region(), s3.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "true"),
				),
			},
		},
	})
}

func TestAccMetaService_unsupported(t *testing.T) {
	dataSourceName := "data.aws_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_unsupported(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "dns_name", fmt.Sprintf("%s.%s.%s", waf.EndpointsID, endpoints.UsGovWest1RegionID, "amazonaws.com")),
					resource.TestCheckResourceAttr(dataSourceName, "partition", endpoints.AwsUsGovPartitionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_prefix", "com.amazonaws"),
					resource.TestCheckResourceAttr(dataSourceName, "region", endpoints.UsGovWest1RegionID),
					resource.TestCheckResourceAttr(dataSourceName, "reverse_dns_name", fmt.Sprintf("%s.%s.%s", "com.amazonaws", endpoints.UsGovWest1RegionID, waf.EndpointsID)),
					resource.TestCheckResourceAttr(dataSourceName, "service_id", waf.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "supported", "false"),
				),
			},
		},
	})
}

func testAccServiceDataSourceConfig_basic() string {
	return fmt.Sprintf(`
data "aws_service" "default" {
  service_id = %[1]q
}
`, ec2.EndpointsID)
}

func testAccServiceDataSourceConfig_byReverseDNSName() string {
	// lintignore:AWSAT003
	return `
data "aws_service" "test" {
  reverse_dns_name = "cn.com.amazonaws.cn-north-1.s3"
}
`
}

func testAccServiceDataSourceConfig_byDNSName() string {
	// lintignore:AWSAT003
	return `
data "aws_service" "test" {
  dns_name = "rds.us-east-1.amazonaws.com"
}
`
}

func testAccServiceDataSourceConfig_byPart() string {
	return `
data "aws_region" "current" {}

data "aws_service" "test" {
  reverse_dns_prefix = "com.amazonaws"
  region             = data.aws_region.current.name
  service_id         = "s3"
}
`
}

func testAccServiceDataSourceConfig_unsupported() string {
	// lintignore:AWSAT003
	return `
data "aws_service" "test" {
  reverse_dns_name = "com.amazonaws.us-gov-west-1.waf"
}
`
}
