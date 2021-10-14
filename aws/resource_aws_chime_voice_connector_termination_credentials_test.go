package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSChimeVoiceConnectorTerminationCredentials_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, chime.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorTerminationCredentialsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorTerminationCredentialsConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorTerminationCredentialsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credentials"},
			},
		},
	})
}

func TestAccAWSChimeVoiceConnectorTerminationCredentials_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, chime.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorTerminationCredentialsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorTerminationCredentialsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorTerminationCredentialsExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceVoiceConnectorTerminationCredentials(), resourceName),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccAWSChimeVoiceConnectorTerminationCredentials_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_termination_credentials.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, chime.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorTerminationCredentialsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorTerminationCredentialsConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorTerminationCredentialsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", "1"),
				),
			},
			{
				Config: testAccAWSChimeVoiceConnectorTerminationCredentialsConfigUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorTerminationCredentialsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "credentials.#", "2"),
				),
			},
		},
	})
}

func testAccCheckAWSChimeVoiceConnectorTerminationCredentialsExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime Voice Connector termination credentials ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeConn
		input := &chime.ListVoiceConnectorTerminationCredentialsInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.ListVoiceConnectorTerminationCredentials(input)
		if err != nil {
			return err
		}

		if resp == nil || resp.Usernames == nil {
			return fmt.Errorf("no Chime Voice Connector Termintation credentials (%s) found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAWSChimeVoiceConnectorTerminationCredentialsDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_chime_voice_connector_termination_credentials" {
			continue
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeConn
		input := &chime.ListVoiceConnectorTerminationCredentialsInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}
		resp, err := conn.ListVoiceConnectorTerminationCredentials(input)

		if tfawserr.ErrMessageContains(err, chime.ErrCodeNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.Usernames != nil {
			return fmt.Errorf("error Chime Voice Connector Termination credentials still exists")
		}
	}

	return nil
}

func testAccAWSChimeVoiceConnectorTerminationCredentialsConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_termination" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id

  calling_regions = ["US"]
  cidr_allow_list = ["50.35.78.0/27"]
}
`, rName)
}

func testAccAWSChimeVoiceConnectorTerminationCredentialsConfig(rName string) string {
	return acctest.ConfigCompose(testAccAWSChimeVoiceConnectorTerminationCredentialsConfigBase(rName), `
resource "aws_chime_voice_connector_termination_credentials" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id

  credentials {
    username = "test1"
    password = "test1!"
  }

  depends_on = [aws_chime_voice_connector_termination.test]
}
`)
}

func testAccAWSChimeVoiceConnectorTerminationCredentialsConfigUpdated(rName string) string {
	return acctest.ConfigCompose(testAccAWSChimeVoiceConnectorTerminationCredentialsConfigBase(rName), `
resource "aws_chime_voice_connector_termination_credentials" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id

  credentials {
    username = "test1"
    password = "test1!"
  }

  credentials {
    username = "test2"
    password = "test2!"
  }

  depends_on = [aws_chime_voice_connector_termination.test]
}
`)
}
