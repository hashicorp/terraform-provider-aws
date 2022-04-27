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
	"regexp"
	"strings"
	"testing"
)

func TestAccPlaybackConfigurationResource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_tailor_playback_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaybackConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ad_decision_server_url", "https://www.example.com/ads"),
					resource.TestCheckResourceAttrSet(resourceName, "playback_configuration_arn"),
					resource.TestCheckResourceAttr(resourceName, "video_content_source_url", "https://www.example.com/source"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateVerify: true,
				ImportState:       true,
			},
		},
	})
}

func TestAccPlaybackConfigurationResource_recreate(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_tailor_playback_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaybackConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ad_decision_server_url", "https://www.example.com/ads"),
					resource.TestCheckResourceAttrSet(resourceName, "playback_configuration_arn"),
					resource.TestCheckResourceAttr(resourceName, "video_content_source_url", "https://www.example.com/source"),
				),
			},
			{
				Taint:  []string{resourceName},
				Config: testAccResourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ad_decision_server_url", "https://www.example.com/ads"),
					resource.TestCheckResourceAttrSet(resourceName, "playback_configuration_arn"),
					resource.TestCheckResourceAttr(resourceName, "video_content_source_url", "https://www.example.com/source"),
				),
			},
		},
	})
}

func TestAccPlaybackConfigurationResource_validateAdDecisionServerURL(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaybackConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceConfig_AdDecisionServerURL(rName, "https://www.example.com/"+strings.Repeat("abcde12345", 2500)), // generate a string longer than 25000 characters
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected length of ad_decision_server_url to be in the range (1 - 25000)")),
			},
		},
	})
}

func TestAccPlaybackConfigurationResource_validateAvailSuppressionMode(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaybackConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceConfig_AvailSuppression(rName, "ON", "20:20:20"),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected avail_suppression_mode to be one of [OFF BEHIND_LIVE_EDGE]")),
			},
			{
				Config:      testAccResourceConfig_AvailSuppression(rName, "OFF", "202020"),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("invalid value for avail_suppression_value (must be valid HH:MM:SS string)")),
			},
		},
	})
}

func TestAccPlaybackConfigurationResource_validateDashConfiguration(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaybackConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceConfig_DashConfiguration(rName, "ENABLED", "EMT_DEFAULT"),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected dash_mpd_location to be one of [DISABLED EMT_DEFAULT]")),
			},
			{
				Config:      testAccResourceConfig_DashConfiguration(rName, "DISABLED", "UNEXPECTED"),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected dash_origin_manifest_type to be one of [SINGLE_PERIOD MULTI_PERIOD]")),
			},
		},
	})
}

func TestAccPlaybackConfigurationResource_validateLivePreRollConfiguration(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaybackConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceConfig_LivePreRollConfiguration(rName, "https://www.example.com/"+strings.Repeat("abcde12345", 2500), 1),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected length of live_pre_roll_configuration.0.ad_decision_server_url to be in the range (1 - 25000)")),
			},
			{
				Config:      testAccResourceConfig_LivePreRollConfiguration(rName, "https://www.example.com/ads", 0),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected live_pre_roll_configuration.0.max_duration_seconds to be at least (1)")),
			},
		},
	})
}

func TestAccPlaybackConfigurationResource_validatePersonalizationThresholdSeconds(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaybackConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceConfig_PersonalizationThresholdSeconds(rName, 0),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected personalization_threshold_seconds to be at least (1)")),
			},
		},
	})
}

func TestAccPlaybackConfigurationResource_validateVideoContentSourceURL(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaybackConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceConfig_VideoContentSourceURL(rName, "https://www.example.com/"+strings.Repeat("abcde12345", 52)), // generate a string longer than 512 characters
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected length of video_content_source_url to be in the range (1 - 512)")),
			},
		},
	})
}

func TestAccPlaybackConfigurationResource_Update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_tailor_playback_configuration.test"
	exampleURL := "https://www.example.com"
	updatedExampleURL := "https://www.updated.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaybackConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_completeResource(rName, exampleURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ad_decision_server_url", exampleURL),
					resource.TestCheckResourceAttr(resourceName, "avail_suppression_mode", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "avail_suppression_value", "10:10:10"),
					resource.TestCheckResourceAttr(resourceName, "bumper.0.end_url", exampleURL),
					resource.TestCheckResourceAttr(resourceName, "bumper.0.start_url", exampleURL),
					resource.TestCheckResourceAttr(resourceName, "cdn_configuration.0.ad_segment_url_prefix", exampleURL),
					resource.TestCheckResourceAttr(resourceName, "cdn_configuration.0.content_segment_url_prefix", exampleURL),
					resource.TestCheckResourceAttr(resourceName, "dash_mpd_location", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "dash_origin_manifest_type", "SINGLE_PERIOD"),
					resource.TestCheckResourceAttr(resourceName, "live_pre_roll_configuration.0.ad_decision_server_url", exampleURL),
					resource.TestCheckResourceAttr(resourceName, "live_pre_roll_configuration.0.max_duration_seconds", "1"),
					resource.TestCheckResourceAttr(resourceName, "manifest_processing_rules.0.ad_marker_passthrough.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "personalization_threshold_seconds", "1"),
					resource.TestCheckResourceAttr(resourceName, "slate_ad_url", exampleURL),
					resource.TestCheckResourceAttr(resourceName, "video_content_source_url", exampleURL),
				),
			},
			{
				Config: testAccResourceConfig_updatedCompleteResource(rName, updatedExampleURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ad_decision_server_url", updatedExampleURL),
					resource.TestCheckResourceAttr(resourceName, "avail_suppression_mode", "BEHIND_LIVE_EDGE"),
					resource.TestCheckResourceAttr(resourceName, "avail_suppression_value", "20:20:20"),
					resource.TestCheckResourceAttr(resourceName, "bumper.0.end_url", updatedExampleURL),
					resource.TestCheckResourceAttr(resourceName, "bumper.0.start_url", updatedExampleURL),
					resource.TestCheckResourceAttr(resourceName, "cdn_configuration.0.ad_segment_url_prefix", updatedExampleURL),
					resource.TestCheckResourceAttr(resourceName, "cdn_configuration.0.content_segment_url_prefix", updatedExampleURL),
					resource.TestCheckResourceAttr(resourceName, "dash_mpd_location", "EMT_DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "dash_origin_manifest_type", "MULTI_PERIOD"),
					resource.TestCheckResourceAttr(resourceName, "live_pre_roll_configuration.0.ad_decision_server_url", updatedExampleURL),
					resource.TestCheckResourceAttr(resourceName, "live_pre_roll_configuration.0.max_duration_seconds", "2"),
					resource.TestCheckResourceAttr(resourceName, "manifest_processing_rules.0.ad_marker_passthrough.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "personalization_threshold_seconds", "2"),
					resource.TestCheckResourceAttr(resourceName, "slate_ad_url", updatedExampleURL),
					resource.TestCheckResourceAttr(resourceName, "video_content_source_url", updatedExampleURL),
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

		var resourceName string

		if arn.IsARN(rs.Primary.ID) {
			resourceArn, err := arn.Parse(rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("error parsing resource arn: %s.\n%s", err, rs.Primary.ID)
			}
			arnSections := strings.Split(resourceArn.Resource, "/")
			resourceName = arnSections[len(arnSections)-1]
		} else {
			resourceName = rs.Primary.ID
		}

		input := &mediatailor.GetPlaybackConfigurationInput{Name: aws.String(resourceName)}
		_, err := conn.GetPlaybackConfiguration(input)

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
  ad_decision_server_url = "https://www.example.com/ads"
  name=%[1]q
  video_content_source_url = "https://www.example.com/source"
}
`, rName)
}

func testAccResourceConfig_AdDecisionServerURL(rName, adDecisionServerURL string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_playback_configuration" "test"{
  ad_decision_server_url = %[2]q
  name=%[1]q
  video_content_source_url = "https://www.example.com/source"
}
`, rName, adDecisionServerURL)
}

func testAccResourceConfig_AvailSuppression(rName, mode, value string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_playback_configuration" "test"{
  ad_decision_server_url = "https://www.example.com/ads"
  avail_suppression_mode = %[2]q
  avail_suppression_value = %[3]q
  name=%[1]q
  video_content_source_url = "https://www.example.com/source"
}
`, rName, mode, value)
}

func testAccResourceConfig_DashConfiguration(rName, mpdLocation, originManifestType string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_playback_configuration" "test"{
  ad_decision_server_url = "https://www.example.com/ads"
  dash_mpd_location = %[2]q
  dash_origin_manifest_type = %[3]q
  name=%[1]q
  video_content_source_url = "https://www.example.com/source"
}
`, rName, mpdLocation, originManifestType)
}

func testAccResourceConfig_LivePreRollConfiguration(rName, adDecisionServerUrl string, maxDurationSeconds int) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_playback_configuration" "test"{
  ad_decision_server_url = "https://www.example.com/ads"
  live_pre_roll_configuration {
    ad_decision_server_url=%[2]q
	max_duration_seconds=%[3]v
  }
  name=%[1]q
  video_content_source_url = "https://www.example.com/source"
}
`, rName, adDecisionServerUrl, maxDurationSeconds)
}

func testAccResourceConfig_PersonalizationThresholdSeconds(rName string, seconds int) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_playback_configuration" "test"{
  ad_decision_server_url = "https://www.example.com/ads"
  personalization_threshold_seconds = %[2]v
  name=%[1]q
  video_content_source_url = "https://www.example.com/source"
}
`, rName, seconds)
}

func testAccResourceConfig_VideoContentSourceURL(rName, videoContentSourceURL string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_playback_configuration" "test"{
  ad_decision_server_url = "https://www.example.com/ads"
  name=%[1]q
  video_content_source_url = %[2]q
}
`, rName, videoContentSourceURL)
}

func testAccResourceConfig_completeResource(rName, url string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_playback_configuration" "test"{
  ad_decision_server_url = %[2]q
  avail_suppression_mode = "OFF"
  avail_suppression_value = "10:10:10"
  bumper {
	end_url = %[2]q
	start_url = %[2]q
  }
  cdn_configuration {
    ad_segment_url_prefix = %[2]q
	content_segment_url_prefix = %[2]q
  }
  dash_mpd_location = "DISABLED"
  dash_origin_manifest_type = "SINGLE_PERIOD"
  live_pre_roll_configuration {
	ad_decision_server_url = %[2]q
	max_duration_seconds = 1
  }
  manifest_processing_rules {
	ad_marker_passthrough {
	  enabled = true
	}
  }
  name=%[1]q
  personalization_threshold_seconds = 1
  slate_ad_url = %[2]q
  video_content_source_url = %[2]q
}
`, rName, url)
}

func testAccResourceConfig_updatedCompleteResource(rName, url string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_playback_configuration" "test"{
  ad_decision_server_url = %[2]q
  avail_suppression_mode = "BEHIND_LIVE_EDGE"
  avail_suppression_value = "20:20:20"
  bumper {
	end_url = %[2]q
	start_url = %[2]q
  }
  cdn_configuration {
    ad_segment_url_prefix = %[2]q
	content_segment_url_prefix = %[2]q
  }
  dash_mpd_location = "EMT_DEFAULT"
  dash_origin_manifest_type = "MULTI_PERIOD"
  live_pre_roll_configuration {
	ad_decision_server_url = %[2]q
	max_duration_seconds = 2
  }
  manifest_processing_rules {
	ad_marker_passthrough {
	  enabled = false
	}
  }
  name=%[1]q
  personalization_threshold_seconds = 2
  slate_ad_url = %[2]q
  video_content_source_url = %[2]q
}
`, rName, url)
}
