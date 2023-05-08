package chimesdkvoice_test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfchimesdkvoice "github.com/hashicorp/terraform-provider-aws/internal/service/chimesdkvoice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"testing"
	"time"
)

const retryTimeout = 30 * time.Second

func TestAccChimeSdkVoiceTags_basic(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkvoice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChimeSdkVoiceTagConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTagsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccChimeSdkVoiceTags_update(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_tags.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkvoice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChimeSdkVoiceTagConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTagsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccChimeSdkVoiceTagConfig_updated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTagsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value12"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func TestAccChimeSdkVoiceTags_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkvoice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccChimeSdkVoiceTagConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTagsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfchimesdkvoice.ResourceTags(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})

}

func testAccCheckTagsExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			fmt.Errorf("not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceConn()

		input := &chimesdkvoice.ListTagsForResourceInput{
			ResourceARN: aws.String(rs.Primary.ID),
		}

		response := &chimesdkvoice.ListTagsForResourceOutput{}
		err := tfresource.Retry(ctx, retryTimeout, func() *retry.RetryError {
			var err error
			response, err = conn.ListTagsForResourceWithContext(ctx, input)
			if err == nil && len(response.Tags) != 0 {
				retry.RetryableError(nil)
			}
			return retry.NonRetryableError(err)
		})

		resp, err := conn.ListTagsForResourceWithContext(ctx, input)
		if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeBadRequestException) {
			return nil
		}

		if err != nil {
			return err
		}

		if resp.Tags == nil {
			return fmt.Errorf("Tags not found for the Vc")
		}

		return nil
	}
}

func testAccChimeSdkVoiceTagConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "test" {
  name               = "vc-%[1]s"
  require_encryption = true
}

resource "aws_chimesdkvoice_tags" "test" {
   resource_arn = aws_chime_voice_connector.test.voice_connector_arn
   tags = {
      "key1" : "value1"
   }
}
`, name)
}

func testAccChimeSdkVoiceTagConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "aws_chime_voice_connector" "test" {
 name               = "vc-%[1]s"
 require_encryption = true
}

resource "aws_chimesdkvoice_tags" "test" {
  resource_arn = aws_chime_voice_connector.test.voice_connector_arn
  tags = {
     "key1" : "value12",
     "key2" : "value2"
  }
}
`, name)
}

func testAccCheckTagsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chimesdkvoice_tags" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceConn()
			input := &chimesdkvoice.ListTagsForResourceInput{
				ResourceARN: aws.String(rs.Primary.ID),
			}

			tfresource.Retry(ctx, retryTimeout, func() *retry.RetryError {
				response, err := conn.ListTagsForResourceWithContext(ctx, input)
				if len(response.Tags) != 0 {
					return retry.RetryableError(nil)
				} else if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeBadRequestException) ||
					tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
					return nil
				}
				return nil
			})

		}
		return nil
	}
}
