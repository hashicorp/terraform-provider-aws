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

func TestAccChimeVoiceConnectorGroup_basic(t *testing.T) {
	var voiceConnectorGroup *chime.VoiceConnectorGroup

	vcgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, chime.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVoiceConnectorGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorGroupConfig(vcgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorGroupExists(resourceName, voiceConnectorGroup),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("vcg-%s", vcgName)),
					resource.TestCheckResourceAttr(resourceName, "connector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector.0.priority", "1"),
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

func TestAccChimeVoiceConnectorGroup_disappears(t *testing.T) {
	var voiceConnectorGroup *chime.VoiceConnectorGroup

	vcgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, chime.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVoiceConnectorGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorGroupConfig(vcgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceConnectorGroupExists(resourceName, voiceConnectorGroup),
					acctest.CheckResourceDisappears(acctest.Provider, tfchime.ResourceVoiceConnectorGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccChimeVoiceConnectorGroup_update(t *testing.T) {
	var voiceConnectorGroup *chime.VoiceConnectorGroup

	vcgName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, chime.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVoiceConnectorGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorGroupConfig(vcgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorGroupExists(resourceName, voiceConnectorGroup),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("vcg-%s", vcgName)),
					resource.TestCheckResourceAttr(resourceName, "connector.#", "1"),
				),
			},
			{
				Config: testAccVoiceConnectorGroupUpdated(vcgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("vcg-updated-%s", vcgName)),
					resource.TestCheckResourceAttr(resourceName, "connector.0.priority", "10"),
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

func testAccVoiceConnectorGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_group" "test" {
  name = "vcg-%[1]s"

  connector {
    voice_connector_id = aws_chime_voice_connector.chime.id
    priority           = 1
  }
}
`, name)
}

func testAccVoiceConnectorGroupUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = false
}

resource "aws_chime_voice_connector_group" "test" {
  name = "vcg-updated-%[1]s"

  connector {
    voice_connector_id = aws_chime_voice_connector.chime.id
    priority           = 10
  }
}
`, name)
}

func testAccCheckVoiceConnectorGroupExists(name string, vc *chime.VoiceConnectorGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime voice connector group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeConn
		input := &chime.GetVoiceConnectorGroupInput{
			VoiceConnectorGroupId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetVoiceConnectorGroup(input)
		if err != nil || resp.VoiceConnectorGroup == nil {
			return err
		}

		vc = resp.VoiceConnectorGroup
		return nil
	}
}

func testAccCheckVoiceConnectorGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_chime_voice_connector" {
			continue
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeConn
		input := &chime.GetVoiceConnectorGroupInput{
			VoiceConnectorGroupId: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetVoiceConnectorGroup(input)
		if err == nil {
			if resp.VoiceConnectorGroup != nil && aws.StringValue(resp.VoiceConnectorGroup.Name) != "" {
				return fmt.Errorf("error Chime Voice Connector still exists")
			}
		}
		return nil
	}
	return nil
}
