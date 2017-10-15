package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSShieldProtection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShieldProtectionRoute53HostedZoneConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionExists("aws_shield_protection.hoge"),
				),
			},
			{
				Config: testAccShieldProtectionElbConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionExists("aws_shield_protection.hoge"),
				),
			},
		},
	})
}

func testAccCheckAWSShieldProtectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).shieldconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_shield_protection" {
			continue
		}

		input := &shield.DescribeProtectionInput{
			ProtectionId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeProtection(input)
		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSShieldProtectionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

// TODO: aws_shield_subscription
const testAccShieldProtectionRoute53HostedZoneConfig = `
resource "aws_route53_zone" "hoge" {
	name = "hashicorp.com."
	comment = "Custom comment"

	tags {
		foo = "bar"
		Name = "tf-route53-tag-test"
	}
}

resource "aws_shield_protection" "hoge" {
  name = "hoge"
  resource = "arn:aws:route53:::hostedzone/${aws_route53_zone.hoge.zone_id}"
}
`

// TODO: aws_shield_subscription
func testAccShieldProtectionElbConfig() string {
	return fmt.Sprintf(`
    resource "aws_elb" "hoge" {
      name = "hoge"
      availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

      listener {
        instance_port = 8000
        instance_protocol = "http"
        lb_port = 80
        // Protocol should be case insensitive
        lb_protocol = "HttP"
      }

    	tags {
    		bar = "baz"
    	}

      cross_zone_load_balancing   = true
    }

    resource "aws_shield_protection" "hoge" {
      name = "hoge"
      resource = "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/hoge"
    }
    `)
}
