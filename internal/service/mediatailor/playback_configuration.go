package mediatailor

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"
)

func ResourcePlaybackConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePlaybackConfigurationCreate,
		ReadContext:   resourcePlaybackConfigurationRead,
		UpdateContext: resourcePlaybackConfigurationUpdate,
		DeleteContext: resourcePlaybackConfigurationDelete,
		Schema: map[string]*schema.Schema{
			"ad_decision_server_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"avail_suppression": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"OFF", "BEHIND_LIVE_EDGE"}, false),
						},
						"value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`), "must be valid HH:MM:SS string"),
						},
					},
				},
			},
			"bumper": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"end_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"start_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"cdn_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_segment_url_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"content_segment_url_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"configuration_aliases": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
			"dash_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"manifest_endpoint_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mpd_location": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"DISABLED", "EMT_DEFAULT"}, false),
						},
						"origin_manifest_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"SINGLE_PERIOD", "MULTI_PERIOD"}, false),
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
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_decision_server_url": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 25000),
						},
						"max_duration_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
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
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_marker_passthrough": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
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
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
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
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"transcode_profile_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"video_content_source_url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourcePlaybackConfigurationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

}

func resourcePlaybackConfigurationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

}

func resourcePlaybackConfigurationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

}

func resourcePlaybackConfigurationDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

}
