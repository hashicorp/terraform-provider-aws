package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsElbHostedZoneName(t *testing.T) {
	lbName := fmt.Sprintf("test-datasource-elb-zname-%s", acctest.RandString(6))
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsElbHostedZoneNameConfigNotFound(),
				ExpectError: regexp.MustCompile(`There is no ACTIVE Load Balancer`),
			},
			{
				Config: testAccDataSourceAwsElbHostedZoneNameConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsElbHostedZoneNameCheck("data.aws_elb_hosted_zone_name.test"),
					resource.TestMatchResourceAttr("data.aws_elb_hosted_zone_name.test", "hosted_zone_name", regexp.MustCompile("^"+lbName)),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataSourceAwsElbHostedZoneNameCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		elb, ok := s.RootModule().Resources["aws_elb.test"]
		if !ok {
			return fmt.Errorf("can't find aws_elb.test in state")
		}

		attr := rs.Primary.Attributes

		if attr["hosted_zone_id"] != elb.Primary.Attributes["zone_id"] {
			return fmt.Errorf(
				"zone_id is %s; want %s",
				attr["hosted_zone_id"],
				elb.Primary.Attributes["zone_id"],
			)
		}

		if attr["dns_name"] != elb.Primary.Attributes["dns_name"] {
			return fmt.Errorf(
				"dns_name is %s; want %s",
				attr["dns_name"],
				elb.Primary.Attributes["dns_name"],
			)
		}

		return nil
	}
}

func testAccDataSourceAwsElbHostedZoneNameConfig(lbName string) string {
	return fmt.Sprintf(`
resource "aws_elb" "test" {
  name               = "%s"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }
}

data "aws_elb_hosted_zone_name" "test" {
  name = "%s"
  depends_on = ["aws_elb.test"]
}
`, lbName, lbName)
}

func testAccDataSourceAwsElbHostedZoneNameConfigNotFound() string {
	return fmt.Sprintf(`
data "aws_elb_hosted_zone_name" "test_not_found" {
  name = "non-existent-%s"
}`, acctest.RandString(16))
}
