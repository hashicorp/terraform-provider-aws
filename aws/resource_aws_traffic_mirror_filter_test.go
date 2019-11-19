package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSTrafficMirrorFilter_basic(t *testing.T) {
	resourceName := "aws_traffic_mirror_filter.filter"
	description := "test filter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSTrafficMirrorFilter(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsTrafficMirrorFilterDestroy,
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccTrafficMirrorFilterConfig(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsTrafficMirrorFilterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "network_services.#", "1"),
				),
			},
			// Test Disable DNS service
			{
				Config: testAccTrafficMirrorFilterConfigWithoutDNS(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsTrafficMirrorFilterExists(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "network_services"),
				),
			},
			// Test Enable DNS service
			{
				Config: testAccTrafficMirrorFilterConfig(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsTrafficMirrorFilterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "network_services.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsTrafficMirrorFilterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		out, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{
			TrafficMirrorFilterIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return err
		}

		if 0 == len(out.TrafficMirrorFilters) {
			return fmt.Errorf("Traffic mirror filter %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTrafficMirrorFilterConfig(description string) string {
	return fmt.Sprintf(`
resource "aws_traffic_mirror_filter" "filter" {
  description = "%s"

  network_services = ["amazon-dns"]
}
`, description)
}

func testAccTrafficMirrorFilterConfigWithoutDNS(description string) string {
	return fmt.Sprintf(`
resource "aws_traffic_mirror_filter" "filter" {
  description = "%s"
}
`, description)
}

func testAccPreCheckAWSTrafficMirrorFilter(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	_, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{})

	if testAccPreCheckSkipError(err) {
		t.Skip("skipping traffic mirror filter acceprance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckAwsTrafficMirrorFilterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_traffic_mirror_filter" {
			continue
		}

		out, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{
			TrafficMirrorFilterIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if isAWSErr(err, "InvalidTrafficMirrorFilterId.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if len(out.TrafficMirrorFilters) != 0 {
			return fmt.Errorf("Traffic mirror filter %s still not destroyed", rs.Primary.ID)
		}
	}

	return nil
}
