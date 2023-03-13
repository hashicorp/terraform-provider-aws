package ec2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPCTrafficMirrorFilter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.TrafficMirrorFilter
	resourceName := "aws_ec2_traffic_mirror_filter.test"
	description := "test filter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorFilter(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorFilterDestroy(ctx),
		Steps: []resource.TestStep{
			//create
			{
				Config: testAccVPCTrafficMirrorFilterConfig_basic(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`traffic-mirror-filter/tmf-.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "network_services.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Test Disable DNS service
			{
				Config: testAccVPCTrafficMirrorFilterConfig_noDNS(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "network_services.#", "0"),
				),
			},
			// Test Enable DNS service
			{
				Config: testAccVPCTrafficMirrorFilterConfig_basic(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(ctx, resourceName, &v),
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
	ctx := acctest.Context(t)
	var v ec2.TrafficMirrorFilter
	resourceName := "aws_ec2_traffic_mirror_filter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorFilter(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorFilterConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(ctx, resourceName, &v),
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
				Config: testAccVPCTrafficMirrorFilterConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCTrafficMirrorFilterConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCTrafficMirrorFilter_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.TrafficMirrorFilter
	resourceName := "aws_ec2_traffic_mirror_filter.test"
	description := "test filter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTrafficMirrorFilter(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficMirrorFilterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCTrafficMirrorFilterConfig_basic(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficMirrorFilterExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTrafficMirrorFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckTrafficMirrorFilter(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()

	_, err := conn.DescribeTrafficMirrorFiltersWithContext(ctx, &ec2.DescribeTrafficMirrorFiltersInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skip("skipping traffic mirror filter acceprance test: ", err)
	}

	if err != nil {
		t.Fatal("Unexpected PreCheck error: ", err)
	}
}

func testAccCheckTrafficMirrorFilterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_traffic_mirror_filter" {
				continue
			}

			_, err := tfec2.FindTrafficMirrorFilterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Traffic Mirror Filter %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTrafficMirrorFilterExists(ctx context.Context, n string, v *ec2.TrafficMirrorFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Traffic Mirror Filter ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()

		output, err := tfec2.FindTrafficMirrorFilterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCTrafficMirrorFilterConfig_basic(description string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  description = %[1]q

  network_services = ["amazon-dns"]
}
`, description)
}

func testAccVPCTrafficMirrorFilterConfig_noDNS(description string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  description = %[1]q
}
`, description)
}

func testAccVPCTrafficMirrorFilterConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccVPCTrafficMirrorFilterConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_traffic_mirror_filter" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
