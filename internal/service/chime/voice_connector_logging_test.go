package chime_test

import (
	"context"
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

func TestAccChimeVoiceConnectorLogging_basic(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, chime.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorLoggingConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_sip_logs", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_media_metric_logs", "true"),
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

func TestAccChimeVoiceConnectorLogging_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, chime.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorLoggingConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceConnectorLoggingExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfchime.ResourceVoiceConnectorLogging(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccChimeVoiceConnectorLogging_update(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chime_voice_connector_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, chime.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceConnectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceConnectorLoggingConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorLoggingExists(ctx, resourceName),
				),
			},
			{
				Config: testAccVoiceConnectorLoggingConfig_updated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVoiceConnectorLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_sip_logs", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_media_metric_logs", "false"),
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

func testAccVoiceConnectorLoggingConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_logging" "test" {
  voice_connector_id       = aws_chime_voice_connector.chime.id
  enable_sip_logs          = true
  enable_media_metric_logs = true
}
`, name)
}

func testAccVoiceConnectorLoggingConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "chime" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chime_voice_connector_logging" "test" {
  voice_connector_id       = aws_chime_voice_connector.chime.id
  enable_sip_logs          = false
  enable_media_metric_logs = false
}
`, name)
}

func testAccCheckVoiceConnectorLoggingExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Chime Voice Connector logging ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeConn()
		input := &chime.GetVoiceConnectorLoggingConfigurationInput{
			VoiceConnectorId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetVoiceConnectorLoggingConfigurationWithContext(ctx, input)
		if err != nil {
			return err
		}

		if resp == nil || resp.LoggingConfiguration == nil {
			return fmt.Errorf("no Chime Voice Connector logging configureation (%s) found", rs.Primary.ID)
		}

		return nil
	}
}
