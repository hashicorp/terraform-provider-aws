package chime_test

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
	tfchime "github.com/hashicorp/terraform-provider-aws/internal/service/chime"
)

func TestAccChimeVoiceConnector_basic(t *testing.T) {
	var voiceConnector *chime.VoiceConnector

	vcName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, chime.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorConfig_basic(vcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorExists(resourceName, voiceConnector),
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

func TestAccChimeVoiceConnector_disappears(t *testing.T) {
	var voiceConnector *chime.VoiceConnector

	vcName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, chime.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorConfig_basic(vcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceConnectorExists(resourceName, voiceConnector),
					acctest.CheckResourceDisappears(acctest.Provider, tfchime.ResourceVoiceConnector(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccChimeVoiceConnector_update(t *testing.T) {
	var voiceConnector *chime.VoiceConnector

	vcName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, chime.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorConfig_basic(vcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorExists(resourceName, voiceConnector),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("vc-%s", vcName)),
					resource.TestCheckResourceAttr(resourceName, "aws_region", chime.VoiceConnectorAwsRegionUsEast1),
					resource.TestCheckResourceAttr(resourceName, "require_encryption", "true"),
				),
			},
			{
				Config: testAccVoiceConnectorConfig_updated(vcName),
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

func testAccVoiceConnectorConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "test" {
  name               = "vc-%s"
  require_encryption = true
}
`, name)
}

func testAccVoiceConnectorConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "test" {
  name               = "vc-%s"
  require_encryption = false
}
`, name)
}

func testAccCheckVoiceConnectorExists(name string, vc *chime.VoiceConnector) resource.TestCheckFunc {
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

func testAccCheckVoiceConnectorDestroy(s *terraform.State) error {
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
