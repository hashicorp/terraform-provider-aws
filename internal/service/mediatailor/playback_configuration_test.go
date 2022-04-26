package mediatailor_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/mediatailor"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"strings"
	"testing"
)

func TestAccPlaybackConfigurationResource_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_media_tailor_playback_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaybackConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
		},
	})
}

func testAccCheckPlaybackConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaTailorConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_tailor_playback_configuration" {
			continue
		}

		resourceArn, err := arn.Parse(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing resource arn: %s", err)
		}
		arnSections := strings.Split(resourceArn.Resource, "/")
		resourceName := arnSections[len(arnSections)-1]

		input := &mediatailor.GetPlaybackConfigurationInput{Name: aws.String(resourceName)}
		_, err = conn.GetPlaybackConfiguration(input)

		if tfawserr.ErrCodeContains(err, "NotFound") {
			continue
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccResourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_playback_configuration" "test"{
  ad_decision_server_url = "https://www.emample.com/ads"
  name=%[1]q
  video_content_source_url = "https://www.example.com/source"
}
`, rName)
}
