package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsRegion_basic(t *testing.T) {
	// Ensure we always get a consistent result
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resourceName := "data.aws_region.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsRegionConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", "ec2.us-east-1.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "name", "us-east-1"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRegion_endpoint(t *testing.T) {
	// Ensure we always get a consistent result
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	endpoint1 := "ec2.us-east-1.amazonaws.com"
	endpoint2 := "ec2.us-east-2.amazonaws.com"
	name1 := "us-east-1"
	name2 := "us-east-2"
	resourceName := "data.aws_region.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsRegionConfig_endpoint(endpoint1),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
				),
			},
			resource.TestStep{
				Config: testAccDataSourceAwsRegionConfig_endpoint(endpoint2),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "false"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
				),
			},
			resource.TestStep{
				Config: testAccDataSourceAwsRegionConfig_currentAndEndpoint(endpoint1),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
				),
			},
			resource.TestStep{
				Config:      testAccDataSourceAwsRegionConfig_endpoint("does-not-exist"),
				ExpectError: regexp.MustCompile(`region not found for endpoint: does-not-exist`),
			},
			resource.TestStep{
				Config:      testAccDataSourceAwsRegionConfig_currentAndEndpoint(endpoint2),
				ExpectError: regexp.MustCompile(`multiple regions matched`),
			},
		},
	})
}

func TestAccDataSourceAwsRegion_name(t *testing.T) {
	// Ensure we always get a consistent result
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	endpoint1 := "ec2.us-east-1.amazonaws.com"
	endpoint2 := "ec2.us-east-2.amazonaws.com"
	name1 := "us-east-1"
	name2 := "us-east-2"
	resourceName := "data.aws_region.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsRegionConfig_name(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
				),
			},
			resource.TestStep{
				Config: testAccDataSourceAwsRegionConfig_name(name2),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "false"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint2),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
				),
			},
			resource.TestStep{
				Config: testAccDataSourceAwsRegionConfig_currentAndName(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionCheck(resourceName),
					resource.TestCheckResourceAttr(resourceName, "current", "true"),
					resource.TestCheckResourceAttr(resourceName, "endpoint", endpoint1),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
				),
			},
			resource.TestStep{
				Config:      testAccDataSourceAwsRegionConfig_name("does-not-exist"),
				ExpectError: regexp.MustCompile(`region not found for name: does-not-exist`),
			},
			resource.TestStep{
				Config:      testAccDataSourceAwsRegionConfig_currentAndName(name2),
				ExpectError: regexp.MustCompile(`multiple regions matched`),
			},
		},
	})
}

func testAccDataSourceAwsRegionCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

const testAccDataSourceAwsRegionConfig_empty = `
data "aws_region" "test" {}
`

func testAccDataSourceAwsRegionConfig_currentAndEndpoint(endpoint string) string {
	return fmt.Sprintf(`
data "aws_region" "test" {
  current  = true
  endpoint = "%s"
}
`, endpoint)
}

func testAccDataSourceAwsRegionConfig_currentAndName(name string) string {
	return fmt.Sprintf(`
data "aws_region" "test" {
  current = true
  name    = "%s"
}
`, name)
}

func testAccDataSourceAwsRegionConfig_endpoint(endpoint string) string {
	return fmt.Sprintf(`
data "aws_region" "test" {
  endpoint = "%s"
}
`, endpoint)
}

func testAccDataSourceAwsRegionConfig_name(name string) string {
	return fmt.Sprintf(`
data "aws_region" "test" {
  name = "%s"
}
`, name)
}
