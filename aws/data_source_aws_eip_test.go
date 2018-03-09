package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsEip_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsEipConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsEipCheck("data.aws_eip.by_id"),
					testAccDataSourceAwsEipCheck("data.aws_eip.by_public_ip"),
					testAccDataSourceAwsEipCheck("data.aws_eip.by_public_dns"),
					resource.TestCheckResourceAttrSet("data.aws_eip.by_public_dns", "public_dns"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEipCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		eipRs, ok := s.RootModule().Resources["aws_eip.test"]
		if !ok {
			return fmt.Errorf("can't find aws_eip.test in state")
		}

		for _, attr := range []string{"id", "public_ip", "public_dns"} {
			if rs.Primary.Attributes[attr] != eipRs.Primary.Attributes[attr] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attr,
					rs.Primary.Attributes[attr],
					eipRs.Primary.Attributes[attr],
				)
			}
		}

		return nil
	}
}

const testAccDataSourceAwsEipConfig = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_eip" "test" {}

data "aws_eip" "by_id" {
  id = "${aws_eip.test.id}"
}

data "aws_eip" "by_public_ip" {
  public_ip = "${aws_eip.test.public_ip}"
}

data "aws_eip" "by_public_dns" {
	public_dns = "${aws_eip.test.public_dns}"
}
`
