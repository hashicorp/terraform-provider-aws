package pinpoint_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccPinpointBaiduChannel_basic(t *testing.T) {
	var channel pinpoint.BaiduChannelResponse
	resourceName := "aws_pinpoint_baidu_channel.channel"

	apiKey := "123"
	apikeyUpdated := "234"
	secretKey := "456"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBaiduChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBaiduChannelConfig_basic(apiKey, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaiduChannelExists(resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "api_key", apiKey),
					resource.TestCheckResourceAttr(resourceName, "secret_key", secretKey),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key", "secret_key"},
			},
			{
				Config: testAccBaiduChannelConfig_basic(apikeyUpdated, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaiduChannelExists(resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "api_key", apikeyUpdated),
					resource.TestCheckResourceAttr(resourceName, "secret_key", secretKey),
				),
			},
		},
	})
}

func testAccCheckBaiduChannelExists(n string, channel *pinpoint.BaiduChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint Baidu channel with that Application ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

		// Check if the Baidu Channel exists
		params := &pinpoint.GetBaiduChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetBaiduChannel(params)

		if err != nil {
			return err
		}

		*channel = *output.BaiduChannelResponse

		return nil
	}
}

func testAccBaiduChannelConfig_basic(apiKey, secretKey string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_baidu_channel" "channel" {
  application_id = aws_pinpoint_app.test_app.application_id

  enabled    = "false"
  api_key    = "%s"
  secret_key = "%s"
}
`, apiKey, secretKey)
}

func testAccCheckBaiduChannelDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_baidu_channel" {
			continue
		}

		// Check if the Baidu channel exists by fetching its attributes
		params := &pinpoint.GetBaiduChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetBaiduChannel(params)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
				continue
			}
			return err
		}
		return fmt.Errorf("Baidu Channel exists when it should be destroyed!")
	}

	return nil
}
