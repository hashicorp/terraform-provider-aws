package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestFindRegionByEc2Endpoint(t *testing.T) {
	var testCases = []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "does-not-exist",
			ErrCount: 1,
		},
		{
			Value:    "ec2.does-not-exist.amazonaws.com",
			ErrCount: 1,
		},
		{
			Value:    "us-east-1", // lintignore:AWSAT003
			ErrCount: 1,
		},
		{
			Value:    "ec2.us-east-1.amazonaws.com", // lintignore:AWSAT003
			ErrCount: 0,
		},
	}

	for _, tc := range testCases {
		_, err := tfec2.FindRegionByEndpoint(tc.Value)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Value, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Value)
		}
	}
}

func TestFindRegionByName(t *testing.T) {
	var testCases = []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "does-not-exist",
			ErrCount: 1,
		},
		{
			Value:    "ec2.us-east-1.amazonaws.com", // lintignore:AWSAT003
			ErrCount: 1,
		},
		{
			Value:    "us-east-1", // lintignore:AWSAT003
			ErrCount: 0,
		},
	}

	for _, tc := range testCases {
		_, err := tfec2.FindRegionByName(tc.Value)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Value, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Value)
		}
	}
}

func TestAccDataSourceAwsRegion_basic(t *testing.T) {
	dataSourceName := "data.aws_region.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "description", regexp.MustCompile(`^.+$`)),
					acctest.CheckResourceAttrRegionalHostnameService(dataSourceName, "endpoint", ec2.EndpointsID),
					resource.TestCheckResourceAttr(dataSourceName, "name", acctest.Region()),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRegion_endpoint(t *testing.T) {
	dataSourceName := "data.aws_region.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionConfig_endpoint(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "description", regexp.MustCompile(`^.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "endpoint", regexp.MustCompile(fmt.Sprintf("^ec2\\.[^.]+\\.%s$", acctest.PartitionDNSSuffix()))),
					resource.TestMatchResourceAttr(dataSourceName, "name", regexp.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRegion_endpointAndName(t *testing.T) {
	dataSourceName := "data.aws_region.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionConfig_endpointAndName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "description", regexp.MustCompile(`^.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "endpoint", regexp.MustCompile(fmt.Sprintf("^ec2\\.[^.]+\\.%s$", acctest.PartitionDNSSuffix()))),
					resource.TestMatchResourceAttr(dataSourceName, "name", regexp.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRegion_name(t *testing.T) {
	dataSourceName := "data.aws_region.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionConfig_name(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "description", regexp.MustCompile(`^.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "endpoint", regexp.MustCompile(fmt.Sprintf("^ec2\\.[^.]+\\.%s$", acctest.PartitionDNSSuffix()))),
					resource.TestMatchResourceAttr(dataSourceName, "name", regexp.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

const testAccDataSourceAwsRegionConfig_empty = `
data "aws_region" "test" {}
`

func testAccDataSourceAwsRegionConfig_endpoint() string {
	return `
data "aws_partition" "test" {}

data "aws_regions" "test" {
}

data "aws_region" "test" {
  endpoint = "ec2.${tolist(data.aws_regions.test.names)[0]}.${data.aws_partition.test.dns_suffix}"
}
`
}

func testAccDataSourceAwsRegionConfig_endpointAndName() string {
	return `
data "aws_partition" "test" {}

data "aws_regions" "test" {
}

data "aws_region" "test" {
  endpoint = "ec2.${tolist(data.aws_regions.test.names)[0]}.${data.aws_partition.test.dns_suffix}"
  name     = tolist(data.aws_regions.test.names)[0]
}
`
}

func testAccDataSourceAwsRegionConfig_name() string {
	return `
data "aws_regions" "test" {
}

data "aws_region" "test" {
  name = tolist(data.aws_regions.test.names)[0]
}
`
}
