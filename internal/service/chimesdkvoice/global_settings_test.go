package chimesdkvoice_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"testing"
	"time"
)

func TestAccChimeSdkVoiceGlobalSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_chimesdkvoice_global_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkvoice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckChimeSdkVoiceGlobalSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChimeSdkVoiceGlobalSettingsConfig_basic(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "voice_connector.0.cdr_bucket", bucketName),
				),
			},
			{
				Config: testAccChimeSdkVoiceGlobalSettingsConfig_basic(bucketNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "voice_connector.0.cdr_bucket", bucketNameUpdated),
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

func testAccChimeSdkVoiceGlobalSettingsConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_chimesdkvoice_global_settings" "test" {
  voice_connector {
    cdr_bucket = %[1]q
  }
  depends_on = [aws_s3_bucket.test]
}
`, bucketName)
}

func testAccCheckChimeSdkVoiceGlobalSettingsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chimesdkvoice_global_settings" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceConn()
			input := &chimesdkvoice.GetGlobalSettingsInput{}

			const retryTimeout = 10 * time.Second
			response := &chimesdkvoice.GetGlobalSettingsOutput{}

			err := tfresource.Retry(ctx, retryTimeout, func() *retry.RetryError {
				var err error
				response, err = conn.GetGlobalSettingsWithContext(ctx, input)
				if err == nil && response.VoiceConnector.CdrBucket != nil {
					return retry.RetryableError(errors.New("error Chime Voice Connector Global settings still exists"))
				}
				return nil
			})

			if err != nil {
				return fmt.Errorf("error Chime Voice Connector Global settings still exists")
			}
		}
		return nil
	}
}
