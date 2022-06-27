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

func TestAccLightsailStaticIP_basic(t *testing.T) {
	var staticIp lightsail.StaticIp
	staticIpName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStaticIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPConfig_basic(staticIpName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPExists("aws_lightsail_static_ip.test", &staticIp),
				),
			},
		},
	})
}

func TestAccLightsailStaticIP_disappears(t *testing.T) {
	var staticIp lightsail.StaticIp
	staticIpName := fmt.Sprintf("tf-test-lightsail-%s", sdkacctest.RandString(5))

	staticIpDestroy := func(*terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn
		_, err := conn.ReleaseStaticIp(&lightsail.ReleaseStaticIpInput{
			StaticIpName: aws.String(staticIpName),
		})

		if err != nil {
			return fmt.Errorf("Error deleting Lightsail Static IP in disapear test")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStaticIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStaticIPConfig_basic(staticIpName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStaticIPExists("aws_lightsail_static_ip.test", &staticIp),
					staticIpDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStaticIPExists(n string, staticIp *lightsail.StaticIp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Static IP ID is set")
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
		*staticIp = *resp.StaticIp
		return nil
	}
}

func testAccCheckStaticIPDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_static_ip" {
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
			if resp.StaticIp != nil {
				return fmt.Errorf("Lightsail Static IP %q still exists", rs.Primary.ID)
			}
		}

		return err
	}

	return nil
}

func testAccStaticIPConfig_basic(staticIpName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_static_ip" "test" {
  name = "%s"
}
`, staticIpName)
}
