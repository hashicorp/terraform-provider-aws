package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSChimeVoiceConnectorTermination_basic(t *testing.T) {
	vcName := acctest.RandomWithPrefix("voice-connector-test")
	resourceName := "aws_chime_voice_connector_termination.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorTerminationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorTerminationConfig(vcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorTerminationExists(resourceName),
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

func testAccCheckAWSChimeVoiceConnectorTerminationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).chimeconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_chime_voice_connector_termination" {
			continue
		}

		_, err := conn.GetVoiceConnectorTermination(&chime.GetVoiceConnectorTerminationInput{
			VoiceConnectorId: aws.String(rs.Primary.Attributes["voice_connector_id"]),
		})
		if err != nil {
			if isAWSErr(err, chime.ErrorCodeNotFound, "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccAWSChimeVoiceConnectorTerminationConfig(vcName string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector_termination" "t" {
  disabled            = true
  cps_limit           = 1
  cidr_allow_list     = ["50.35.78.96/31"]
  calling_regions     = ["US", "CA"]
  voice_connector_id  = aws_chime_voice_connector.chime.id
}

resource "aws_chime_voice_connector" "chime" {
  name                = %s
  require_encryption  = true
  aws_region          = "us-east-1"
}
`, vcName)
}

func testAccCheckAWSChimeVoiceConnectorTerminationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}
