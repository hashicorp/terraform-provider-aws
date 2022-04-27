package mediatailor

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediatailor"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourcePlaybackConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePlaybackConfigurationRead,
		Schema: map[string]*schema.Schema{
			"ad_decision_server_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"avail_suppression": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"bumper": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"end_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"start_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"cdn_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_segment_url_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"content_segment_url_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dash_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"manifest_endpoint_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mpd_location": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"origin_manifest_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"hls_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"manifest_endpoint_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"live_pre_roll_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_decision_server_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"max_duration_seconds": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"log_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"percent_enabled": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"manifest_processing_rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_marker_passthrough": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"personalization_threshold_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"playback_configuration_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"playback_endpoint_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"session_initialization_endpoint_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"slate_ad_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"transcode_profile_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"video_content_source_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePlaybackConfigurationRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaTailorConn

	configurationName := d.Get("name").(string)

	input := &mediatailor.GetPlaybackConfigurationInput{Name: &configurationName}

	result, err := conn.GetPlaybackConfiguration(input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while retrieving the configuration: %w", err))
	}

	d.SetId(aws.StringValue(result.Name))
	if err = d.Set("ad_decision_server_url", result.AdDecisionServerUrl); err != nil {
		return diag.FromErr(fmt.Errorf("error setting ad_decision_server_url: %s", err))
	}

	if err = d.Set("avail_suppression", flattenAvailSuppression(result.AvailSuppression)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting avail_suppression_mode: %s", err))
	}

	if err = d.Set("bumper", flattenBumper(result.Bumper)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting bumper: %s", err))
	}

	if err = d.Set("cdn_configuration", flattenCdnConfiguration(result.CdnConfiguration)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting cdn_configuration: %s", err))
	}

	if err = d.Set("dash_configuration", flattenDashConfiguration(result.DashConfiguration)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting dash_configuration: %s", err))
	}

	if err = d.Set("hls_configuration", flattenHlsConfiguration(result.HlsConfiguration)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting hls_configuration: %s", err))
	}

	if err = d.Set("live_pre_roll_configuration", flattenLivePreRollConfiguration(result.LivePreRollConfiguration)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting live_pre_roll_configuration: %s", err))
	}

	if result.LogConfiguration != nil {
		if err = d.Set("log_configuration", []interface{}{map[string]interface{}{
			"percent_enabled": result.LogConfiguration.PercentEnabled,
		}}); err != nil {
			return diag.FromErr(fmt.Errorf("error setting log_configuration: %s", err))
		}
	} else {
		if err = d.Set("log_configuration", []interface{}{map[string]interface{}{
			"percent_enabled": 0,
		}}); err != nil {
			return diag.FromErr(fmt.Errorf("error setting log_configuration: %s", err))
		}
	}

	if err = d.Set("manifest_processing_rules", flattenManifestProcessingRules(result.ManifestProcessingRules)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting manifest_processing_rules: %s", err))
	}

	if err = d.Set("name", result.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %s", err))
	}

	if err = d.Set("personalization_threshold_seconds", result.PersonalizationThresholdSeconds); err != nil {
		return diag.FromErr(fmt.Errorf("error setting personalization_threshold_seconds: %s", err))
	}

	if err = d.Set("playback_configuration_arn", result.PlaybackConfigurationArn); err != nil {
		return diag.FromErr(fmt.Errorf("error setting playback_configuration_arn: %s", err))
	}

	if err = d.Set("playback_endpoint_prefix", result.PlaybackEndpointPrefix); err != nil {
		return diag.FromErr(fmt.Errorf("error setting playback_endpoint_prefix: %s", err))
	}

	if err = d.Set("session_initialization_endpoint_prefix", result.SessionInitializationEndpointPrefix); err != nil {
		return diag.FromErr(fmt.Errorf("error setting session_initialization_endpoint_prefix: %s", err))
	}

	if err = d.Set("slate_ad_url", result.SlateAdUrl); err != nil {
		return diag.FromErr(fmt.Errorf("error setting slate_ad_url: %s", err))
	}

	if err = d.Set("tags", result.Tags); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}

	if err = d.Set("transcode_profile_name", result.TranscodeProfileName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting transcode_profile_name: %s", err))
	}

	if err = d.Set("video_content_source_url", result.VideoContentSourceUrl); err != nil {
		return diag.FromErr(fmt.Errorf("error setting video_content_source_url: %s", err))
	}

	return nil
}

func flattenAvailSuppression(s *mediatailor.AvailSuppression) []interface{} {
	if s == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"mode":  s.Mode,
		"value": s.Value,
	}
	return []interface{}{m}
}

func flattenBumper(s *mediatailor.Bumper) []interface{} {
	if s == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"end_url":   s.EndUrl,
		"start_url": s.StartUrl,
	}
	return []interface{}{m}
}

func flattenCdnConfiguration(s *mediatailor.CdnConfiguration) []interface{} {
	if s == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"ad_segment_url_prefix":      s.AdSegmentUrlPrefix,
		"content_segment_url_prefix": s.ContentSegmentUrlPrefix,
	}
	return []interface{}{m}
}

func flattenDashConfiguration(s *mediatailor.DashConfiguration) []interface{} {
	if s == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"manifest_endpoint_prefix": s.ManifestEndpointPrefix,
		"mpd_location":             s.MpdLocation,
		"origin_manifest_type":     s.OriginManifestType,
	}
	return []interface{}{m}
}

func flattenHlsConfiguration(s *mediatailor.HlsConfiguration) []interface{} {
	if s == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"manifest_endpoint_prefix": s.ManifestEndpointPrefix,
	}
	return []interface{}{m}
}

func flattenLivePreRollConfiguration(s *mediatailor.LivePreRollConfiguration) []interface{} {
	if s == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"ad_decision_server_url": s.AdDecisionServerUrl,
		"max_duration_seconds":   s.MaxDurationSeconds,
	}
	return []interface{}{m}
}

func flattenManifestProcessingRules(s *mediatailor.ManifestProcessingRules) []interface{} {
	if s == nil {
		return []interface{}{}
	}

	if s.AdMarkerPassthrough == nil {
		return []interface{}{
			map[string]interface{}{
				"ad_marker_passthrough": nil,
			},
		}
	}

	m := map[string]interface{}{
		"ad_marker_passthrough": []interface{}{map[string]interface{}{
			"enabled": s.AdMarkerPassthrough.Enabled,
		}},
	}
	return []interface{}{m}
}
