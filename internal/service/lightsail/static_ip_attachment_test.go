package lightsail_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccLightsailStaticIPAttachment_basic(t *testing.T) {
	var staticIp lightsail.StaticIp
	staticIpName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	instanceName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	keypairName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStaticIPAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPAttachmentConfig_basic(staticIpName, instanceName, keypairName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPAttachmentExists("aws_lightsail_static_ip_attachment.test", &staticIp),
					resource.TestCheckResourceAttrSet("aws_lightsail_static_ip_attachment.test", "ip_address"),
				),
			},
		},
	})
}

func TestAccLightsailStaticIPAttachment_disappears(t *testing.T) {
	var staticIp lightsail.StaticIp
	staticIpName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	instanceName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))
	keypairName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))

	staticIpDestroy := func(*terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.DetachStaticIp(&lightsail.DetachStaticIpInput{
			StaticIpName: aws.String(staticIpName),
		})

		if err != nil {
			return fmt.Errorf("Error deleting Lightsail Static IP in disappear test")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStaticIPAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPAttachmentConfig_basic(staticIpName, instanceName, keypairName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPAttachmentExists("aws_lightsail_static_ip_attachment.test", &staticIp),
					staticIpDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStaticIPAttachmentExists(n string, staticIp *lightsail.StaticIp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Static IP Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetStaticIp(&lightsail.GetStaticIpInput{
			StaticIpName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if resp == nil || resp.StaticIp == nil {
			return fmt.Errorf("Static IP (%s) not found", rs.Primary.ID)
		}

		if !*resp.StaticIp.IsAttached {
			return fmt.Errorf("Static IP (%s) not attached", rs.Primary.ID)
		}

		*staticIp = *resp.StaticIp
		return nil
	}
}

func testAccCheckStaticIPAttachmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_static_ip_attachment" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		resp, err := conn.GetStaticIp(&lightsail.GetStaticIpInput{
			StaticIpName: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
			continue
		}

		if err == nil {
			if *resp.StaticIp.IsAttached {
				return fmt.Errorf("Lightsail Static IP %q is still attached (to %q)", rs.Primary.ID, *resp.StaticIp.AttachedTo)
			}
		}

		return err
	}

	return nil
}

func testAccStaticIPAttachmentConfig_basic(staticIpName, instanceName, keypairName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_static_ip_attachment" "test" {
  static_ip_name = aws_lightsail_static_ip.test.name
  instance_name  = aws_lightsail_instance.test.name
}

resource "aws_lightsail_static_ip" "test" {
  name = "%s"
}

resource "aws_lightsail_instance" "test" {
  name              = "%s"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "micro_1_0"
  key_pair_name     = aws_lightsail_key_pair.test.name
}

resource "aws_lightsail_key_pair" "test" {
  name = "%s"
}
`, staticIpName, instanceName, keypairName)
}
