package pinpoint_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

/**
 Before running this test, the following ENV variable must be set:

 GCM_API_KEY - Google Cloud Messaging Api Key
**/

func TestAccPinpointGCMChannel_basic(t *testing.T) {
	var channel pinpoint.GCMChannelResponse
	resourceName := "aws_pinpoint_gcm_channel.test_gcm_channel"

	if os.Getenv("GCM_API_KEY") == "" {
		t.Skipf("GCM_API_KEY env missing, skip test")
	}

	apiKey := os.Getenv("GCM_API_KEY")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGCMChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGCMChannelConfig_basic(apiKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGCMChannelExists(resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
			{
				Config: testAccGCMChannelConfig_basic(apiKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGCMChannelExists(resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func testAccCheckGCMChannelExists(n string, channel *pinpoint.GCMChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint GCM Channel with that application ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

		// Check if the app exists
		params := &pinpoint.GetGcmChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetGcmChannel(params)

		if err != nil {
			return err
		}

		*channel = *output.GCMChannelResponse

		return nil
	}
}

func testAccGCMChannelConfig_basic(apiKey string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_gcm_channel" "test_gcm_channel" {
  application_id = aws_pinpoint_app.test_app.application_id
  enabled        = "false"
  api_key        = "%s"
}
`, apiKey)
}

func testAccCheckGCMChannelDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_gcm_channel" {
			continue
		}

		// Check if the event stream exists
		params := &pinpoint.GetGcmChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetGcmChannel(params)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
				continue
			}
			return err
		}
		return fmt.Errorf("GCM Channel exists when it should be destroyed!")
	}

	return nil
}
