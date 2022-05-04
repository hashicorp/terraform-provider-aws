package mediatailor_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/mediatailor"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"regexp"
	"testing"
)

func TestAccPlaybackConfigurationDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_media_tailor_playback_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaybackConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlaybackConfigurationDataSourceBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ad_decision_server_url", "https://www.example.com/ads"),
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "video_content_source_url", "https://www.example.com/source"),
					resource.TestMatchResourceAttr(dataSourceName, "dash_configuration.0.manifest_endpoint_prefix", regexp.MustCompile(`^https://(\w+).mediatailor.(-|\w)+.\w+.\w+/\w+/dash/\w+/(-|\w)+/$`)),
					resource.TestMatchResourceAttr(dataSourceName, "hls_configuration.0.manifest_endpoint_prefix", regexp.MustCompile(`^https://(\w+).mediatailor.(-|\w)+.\w+.\w+/\w+/master/\w+/(-|\w)+/$`)),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "mediatailor", regexp.MustCompile(`playbackConfiguration/.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "playback_endpoint_prefix", regexp.MustCompile(`^https://(\w+).mediatailor.(-|\w)+.\w+.\w+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "session_initialization_endpoint_prefix", regexp.MustCompile(`^https://(\w+).mediatailor.(-|\w)+.\w+.\w+/\w+/session/\w+/(-|\w)+/$`)),
				),
			},
		},
	})
}

func testAccPlaybackConfigurationDataSourceBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_playback_configuration" "test"{
  ad_decision_server_url = "https://www.example.com/ads"
  name = "%[1]s"
  video_content_source_url = "https://www.example.com/source"
}

data "aws_media_tailor_playback_configuration" "test" {
  name = aws_media_tailor_playback_configuration.test.name
}
`, rName)
}
