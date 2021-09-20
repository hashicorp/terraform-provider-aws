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
)

func TestAccAWSChimeVoiceConnector_basic(t *testing.T) {
	var voiceConnector *chime.VoiceConnector

	vcName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, chime.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorConfig(vcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorExists(resourceName, voiceConnector),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("vc-%s", vcName)),
					resource.TestCheckResourceAttr(resourceName, "aws_region", chime.VoiceConnectorAwsRegionUsEast1),
					resource.TestCheckResourceAttr(resourceName, "require_encryption", "true"),
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

func TestAccAWSChimeVoiceConnector_disappears(t *testing.T) {
	var voiceConnector *chime.VoiceConnector

	vcName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, chime.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorConfig(vcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorExists(resourceName, voiceConnector),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceVoiceConnector(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSChimeVoiceConnector_update(t *testing.T) {
	var voiceConnector *chime.VoiceConnector

	vcName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, chime.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorConfig(vcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorExists(resourceName, voiceConnector),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("vc-%s", vcName)),
					resource.TestCheckResourceAttr(resourceName, "aws_region", chime.VoiceConnectorAwsRegionUsEast1),
					resource.TestCheckResourceAttr(resourceName, "require_encryption", "true"),
				),
			},
			{
				Config: testAccAWSChimeVoiceConnectorUpdated(vcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "require_encryption", "false"),
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

func testAccAWSChimeVoiceConnectorConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "test" {
  name               = "vc-%s"
  require_encryption = true
}
`, name)
}

func testAccAWSChimeVoiceConnectorUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "test" {
  name               = "vc-%s"
  require_encryption = false
}
`, name)
}

func testAccCheckAWSChimeVoiceConnectorExists(name string, vc *chime.VoiceConnector) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime voice connector ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeConn
		input := &chime.GetVoiceConnectorInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetVoiceConnector(input)
		if err != nil {
			return err
		}

		vc = resp.VoiceConnector

		return nil
	}
}

func testAccCheckAWSChimeVoiceConnectorDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_chime_voice_connector" {
			continue
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeConn
		input := &chime.GetVoiceConnectorInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetVoiceConnector(input)
		if err == nil {
			if resp.VoiceConnector != nil && aws.StringValue(resp.VoiceConnector.Name) != "" {
				return fmt.Errorf("error Chime Voice Connector still exists")
			}
		}
		return nil
	}
	return nil
}
