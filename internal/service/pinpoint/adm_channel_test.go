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
 Before running this test, the following two ENV variables must be set:

 ADM_CLIENT_ID     - Amazon ADM OAuth Credentials Client ID
 ADM_CLIENT_SECRET - Amazon ADM OAuth Credentials Client Secret
**/

type testAccADMChannelConfiguration struct {
	ClientID     string
	ClientSecret string
}

func testAccADMChannelConfigurationFromEnv(t *testing.T) *testAccADMChannelConfiguration {
	if os.Getenv("ADM_CLIENT_ID") == "" {
		t.Skipf("ADM_CLIENT_ID ENV is missing")
	}

	if os.Getenv("ADM_CLIENT_SECRET") == "" {
		t.Skipf("ADM_CLIENT_SECRET ENV is missing")
	}

	conf := testAccADMChannelConfiguration{
		ClientID:     os.Getenv("ADM_CLIENT_ID"),
		ClientSecret: os.Getenv("ADM_CLIENT_SECRET"),
	}

	return &conf
}

func TestAccPinpointADMChannel_basic(t *testing.T) {
	var channel pinpoint.ADMChannelResponse
	resourceName := "aws_pinpoint_adm_channel.channel"

	config := testAccADMChannelConfigurationFromEnv(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckADMChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccADMChannelConfig_basic(config),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADMChannelExists(resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_id", "client_secret"},
			},
			{
				Config: testAccADMChannelConfig_basic(config),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADMChannelExists(resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func testAccCheckADMChannelExists(n string, channel *pinpoint.ADMChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint ADM channel with that Application ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

		// Check if the ADM Channel exists
		params := &pinpoint.GetAdmChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetAdmChannel(params)

		if err != nil {
			return err
		}

		*channel = *output.ADMChannelResponse

		return nil
	}
}

func testAccADMChannelConfig_basic(conf *testAccADMChannelConfiguration) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_adm_channel" "channel" {
  application_id = aws_pinpoint_app.test_app.application_id

  client_id     = "%s"
  client_secret = "%s"
  enabled       = false
}
`, conf.ClientID, conf.ClientSecret)
}

func testAccCheckADMChannelDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_adm_channel" {
			continue
		}

		// Check if the ADM channel exists by fetching its attributes
		params := &pinpoint.GetAdmChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetAdmChannel(params)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
				continue
			}
			return err
		}
		return fmt.Errorf("ADM Channel exists when it should be destroyed!")
	}

	return nil
}
