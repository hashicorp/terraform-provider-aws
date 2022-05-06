package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccVPCTrafficMirrorFilter_basic(t *testing.T) {
	var v ec2.TrafficMirrorFilter
	resourceName := "aws_ec2_traffic_mirror_filter.test"
	description := "test filter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorFilter(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorFilterDestroy,
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccTrafficMirrorFilterConfig(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`traffic-mirror-filter/tmf-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "network_services.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Test Disable DNS service
			{
				Config: testAccTrafficMirrorFilterConfigWithoutDNS(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(resourceName, &v),
					resource.TestCheckNoResourceAttr(resourceName, "network_services"),
				),
			},
			// Test Enable DNS service
			{
				Config: testAccTrafficMirrorFilterConfig(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(resourceName, &v),
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

func TestAccVPCTrafficMirrorFilter_tags(t *testing.T) {
	var v ec2.TrafficMirrorFilter
	resourceName := "aws_ec2_traffic_mirror_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorFilter(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficMirrorFilterConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTrafficMirrorFilterConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTrafficMirrorFilterConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCTrafficMirrorFilter_disappears(t *testing.T) {
	var v ec2.TrafficMirrorFilter
	resourceName := "aws_ec2_traffic_mirror_filter.test"
	description := "test filter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorFilter(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTrafficMirrorFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficMirrorFilterConfig(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTrafficMirrorFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTrafficMirrorFilterExists(name string, traffic *ec2.TrafficMirrorFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
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

		*traffic = *out.TrafficMirrorFilters[0]

		return nil
	}
}

func testAccTrafficMirrorFilterConfig(description string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  description = "%s"

  network_services = ["amazon-dns"]
}
`, description)
}

func testAccTrafficMirrorFilterConfigWithoutDNS(description string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  description = "%s"
}
`, description)
}

func testAccTrafficMirrorFilterConfigTags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccTrafficMirrorFilterConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccPreCheckTrafficMirrorFilter(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	_, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skip("skipping traffic mirror filter acceprance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckTrafficMirrorFilterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_traffic_mirror_filter" {
			continue
		}

		out, err := conn.DescribeTrafficMirrorFilters(&ec2.DescribeTrafficMirrorFiltersInput{
			TrafficMirrorFilterIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if tfawserr.ErrCodeEquals(err, "InvalidTrafficMirrorFilterId.NotFound") {
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
