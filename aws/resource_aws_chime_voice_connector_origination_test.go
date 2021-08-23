package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSChimeVoiceConnectorOrigination_basic(t *testing.T) {
	var voiceConnector *chime.VoiceConnector
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_origination.test"
	vcResourceName := "aws_chime_voice_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorOriginationConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorExists(vcResourceName, voiceConnector),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "route.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "route.0.priority", "1"),
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

func TestAccAWSChimeVoiceConnectorOrigination_disappears(t *testing.T) {
	var voiceConnector *chime.VoiceConnector
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_origination.test"
	vcResourceName := "aws_chime_voice_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorOriginationConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorExists(vcResourceName, voiceConnector),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsChimeVoiceConnectorOrigination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSChimeVoiceConnectorOrigination_update(t *testing.T) {
	var voiceConnector *chime.VoiceConnector
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_origination.test"
	vcResourceName := "aws_chime_voice_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorOriginationConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorExists(vcResourceName, voiceConnector),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
				),
			},
			{
				Config: testAccAWSChimeVoiceConnectorOriginationUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "route.1.port", "5060"),
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

func testAccAWSChimeVoiceConnectorOriginationConfig(name string) string {
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

func testAccAWSChimeVoiceConnectorOriginationUpdated(name string) string {
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
