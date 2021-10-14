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

func TestAccAWSChimeVoiceConnectorStreaming_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_streaming.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorStreamingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorStreamingConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorStreamingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_retention", "5"),
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "streaming_notification_targets.#", "1"),
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

func TestAccAWSChimeVoiceConnectorStreaming_disappears(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_streaming.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorStreamingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorStreamingConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorStreamingExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsChimeVoiceConnectorStreaming(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSChimeVoiceConnectorStreaming_update(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_chime_voice_connector_streaming.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, chime.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSChimeVoiceConnectorStreamingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSChimeVoiceConnectorStreamingConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorStreamingExists(resourceName),
				),
			},
			{
				Config: testAccAWSChimeVoiceConnectorStreamingUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSChimeVoiceConnectorStreamingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_retention", "2"),
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "streaming_notification_targets.#", "2"),
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

func testAccAWSChimeVoiceConnectorStreamingConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_streaming" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id

  disabled                       = false
  data_retention                 = 5
  streaming_notification_targets = ["SQS"]
}
`, name)
}

func testAccAWSChimeVoiceConnectorStreamingUpdated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_streaming" "test" {
  voice_connector_id = aws_chime_voice_connector.chime.id

  disabled                       = false
  data_retention                 = 2
  streaming_notification_targets = ["SQS", "SNS"]
}
`, name)
}

func testAccCheckAWSChimeVoiceConnectorStreamingExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime Voice Connector streaming configuration ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).chimeconn
		input := &chime.GetVoiceConnectorStreamingConfigurationInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetVoiceConnectorStreamingConfiguration(input)
		if err != nil {
			return err
		}

		if resp == nil || resp.StreamingConfiguration == nil {
			return fmt.Errorf("no Chime Voice Connector Streaming configuration (%s) found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAWSChimeVoiceConnectorStreamingDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_chime_voice_connector_termination" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).chimeconn
		input := &chime.GetVoiceConnectorStreamingConfigurationInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetVoiceConnectorStreamingConfiguration(input)

		if tfawserr.ErrMessageContains(err, chime.ErrCodeNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.StreamingConfiguration != nil {
			return fmt.Errorf("error Chime Voice Connector streaming configuration still exists")
		}
	}

	return nil
}
