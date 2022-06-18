package chime_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfchime "github.com/hashicorp/terraform-provider-aws/internal/service/chime"
)

func TestAccChimeVoiceConnectorOrigination_basic(t *testing.T) {
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_origination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, chime.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVoiceConnectorOriginationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorOriginationConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorOriginationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"protocol": "TCP",
						"priority": "1",
					}),
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

func TestAccChimeVoiceConnectorOrigination_disappears(t *testing.T) {
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_origination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, chime.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVoiceConnectorOriginationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorOriginationConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceConnectorOriginationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfchime.ResourceVoiceConnectorOrigination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccChimeVoiceConnectorOrigination_update(t *testing.T) {
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_origination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, chime.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckVoiceConnectorOriginationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorOriginationConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorOriginationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
				),
			},
			{
				Config: testAccVoiceConnectorOriginationConfig_updated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorOriginationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"protocol": "TCP",
						"port":     "5060",
						"priority": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"protocol": "UDP",
						"priority": "2",
					}),
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

func testAccCheckVoiceConnectorOriginationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime voice connector origination ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeConn
		input := &chime.GetVoiceConnectorOriginationInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetVoiceConnectorOrigination(input)
		if err != nil {
			return err
		}

		if resp == nil || resp.Origination == nil {
			return fmt.Errorf("Chime Voice Connector Origination (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVoiceConnectorOriginationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_chime_voice_connector_origination" {
			continue
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeConn
		input := &chime.GetVoiceConnectorOriginationInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetVoiceConnectorOrigination(input)

		if tfawserr.ErrCodeEquals(err, chime.ErrCodeNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.Origination != nil {
			return fmt.Errorf("error Chime Voice Connector Origination (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccVoiceConnectorOriginationConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "test" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_origination" "test" {
  route {
    host     = "200.100.12.1"
    port     = 5060
    protocol = "TCP"
    priority = 1
    weight   = 1
  }
  voice_connector_id = aws_chime_voice_connector.test.id
}
`, name)
}

func testAccVoiceConnectorOriginationConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "test" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_origination" "test" {
  voice_connector_id = aws_chime_voice_connector.test.id

  route {
    host     = "200.100.12.1"
    port     = 5060
    protocol = "TCP"
    priority = 1
    weight   = 1
  }

  route {
    host     = "209.166.124.147"
    protocol = "UDP"
    priority = 2
    weight   = 30
  }
}
`, name)
}
