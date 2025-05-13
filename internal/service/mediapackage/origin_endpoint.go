// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackage

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediapackage"
	"github.com/aws/aws-sdk-go-v2/service/mediapackage/types"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mediapackage/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_media_package_origin_endpoint", name="Origin Endpoint")
func ResourceOriginEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOriginEndpointCreate,
		ReadWithoutTimeout:   resourceOriginEndpointRead,
		UpdateWithoutTimeout: resourceOriginEndpointUpdate,
		DeleteWithoutTimeout: resourceOriginEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"channel_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"origin_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorization": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cdn_identifier_secret": {
							Type:     schema.TypeString,
							Required: true,
						},
						"secrets_role_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"cmaf_package": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"speke_key_provider": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: spekeKeyProviderSchema,
										},
									},
									"constant_initialization_vector": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"encryption_method": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"key_rotation_interval_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"hls_manifests": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrID: {
										Type:     schema.TypeString,
										Required: true,
									},
									"ad_markers": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"ad_triggers": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"ads_on_delivery_restrictions": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"include_iframe_only_stream": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"manifest_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"playlist_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"playlist_window_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"program_date_time_interval_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"segment_duration_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"segment_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"stream_selection": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: streamSelectionSchema,
							},
						},
					},
				},
			},
			"dash_package": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_triggers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ads_on_delivery_restrictions": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"encryption": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"speke_key_provider": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: spekeKeyProviderSchema,
										},
									},
									"key_rotation_interval_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"include_iframe_only_stream": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"manifest_layout": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"manifest_window_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"min_buffer_time_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"min_update_period_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"period_triggers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"profile": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"segment_duration_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"segment_template_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"stream_selection": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: streamSelectionSchema,
							},
						},
						"suggested_presentation_delay_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"utc_timing": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"utc_timing_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"hls_package": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_markers": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ad_triggers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ads_on_delivery_restrictions": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"encryption": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"speke_key_provider": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: spekeKeyProviderSchema,
										},
									},
									"constant_initialization_vector": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"encryption_method": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"key_rotation_interval_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"repeat_ext_x_key": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"include_dvb_subtitles": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"include_iframe_only_stream": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"playlist_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"playlist_window_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"program_date_time_interval_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"segment_duration_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"stream_selection": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: streamSelectionSchema,
							},
						},
						"use_audio_rendition_group": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"include_iframe_only_stream": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"manifest_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"mss_package": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"speke_key_provider": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: spekeKeyProviderSchema,
										},
									},
								},
							},
						},
						"manifest_window_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"segment_duration_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"stream_selection": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: streamSelectionSchema,
							},
						},
					},
				},
			},
			"origination": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"start_over_window_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"time_delay_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"whitelist": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

const (
	ResNameOriginEndpoint = "Origin Endpoint"
)

func resourceOriginEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	// Because the IAM role is not evaluated until training time, we need to ensure we wait for the IAM propagation delay
	time.Sleep(10 * time.Second)

	in := &mediapackage.CreateOriginEndpointInput{
		ChannelId: aws.String(d.Get("channel_id").(string)),
		Id:        aws.String(d.Get("origin_endpoint_id").(string)),
	}

	if v, ok := d.GetOk("authorization"); ok && len(v.([]interface{})) > 0 {
		in.Authorization = expandAuthorization(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("cmaf_package"); ok && len(v.([]interface{})) > 0 {
		in.CmafPackage = expandCmafPackage(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("dash_package"); ok && len(v.([]interface{})) > 0 {
		in.DashPackage = expandDashPackage(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("hls_package"); ok {
		in.HlsPackage = expandHlsPackage(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("mss_package"); ok && len(v.([]interface{})) > 0 {
		in.MssPackage = expandMssPackage(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("manifest_name"); ok {
		in.ManifestName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("origination"); ok {
		in.Origination = types.Origination(v.(string))
	}

	if v, ok := d.GetOk("start_over_window_seconds"); ok {
		in.StartoverWindowSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("time_delay_seconds"); ok {
		in.TimeDelaySeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("whitelist"); ok {
		in.Whitelist = flex.ExpandStringValueList(v.([]interface{}))
	}

	out, err := conn.CreateOriginEndpoint(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionCreating, ResNameOriginEndpoint, d.Get(names.AttrARN).(string), err)
	}

	if out == nil || out.Origination == "" {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionCreating, ResNameOriginEndpoint, d.Get(names.AttrARN).(string), errors.New("empty output"))
	}

	if err := d.Set("url", out.Url); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionCreating, ResNameOriginEndpoint, d.Get(names.AttrARN).(string), err)
	}
	d.SetId(aws.ToString(out.Id))

	return append(diags, resourceOriginEndpointRead(ctx, d, meta)...)
}

func resourceOriginEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	out, err := findOriginEndpoint(ctx, conn, d.Get("id").(string), d.Get("channel_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaPackage OriginEndpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionReading, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("arn", out.Arn); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("origin_endpoint_id", out.Id); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("channel_id", out.ChannelId); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("authorization", flattenAuthorization(out.Authorization)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("cmaf_package", flattenCmafPackage(out.CmafPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("dash_package", flattenDashPackage(out.DashPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("description", out.Description); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("manifest_name", out.ManifestName); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("hls_package", flattenHlsPackage(out.HlsPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("mss_package", flattenMssPackage(out.MssPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("origination", out.Origination); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("start_over_window_seconds", out.StartoverWindowSeconds); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("time_delay_seconds", out.TimeDelaySeconds); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("whitelist", out.Whitelist); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	return diags
}

func resourceOriginEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	in := &mediapackage.UpdateOriginEndpointInput{
		Id: aws.String(d.Get("origin_endpoint_id").(string)),
	}

	if d.HasChanges() {
		if v, ok := d.GetOk("authorization"); ok && len(v.([]interface{})) > 0 {
			in.Authorization = expandAuthorization(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("cmaf_package"); ok && len(v.([]interface{})) > 0 {
			in.CmafPackage = expandCmafPackage(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("dash_package"); ok && len(v.([]interface{})) > 0 {
			in.DashPackage = expandDashPackage(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("description"); ok {
			in.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("hls_package"); ok && len(v.([]interface{})) > 0 {
			in.HlsPackage = expandHlsPackage(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("mss_package"); ok && len(v.([]interface{})) > 0 {
			in.MssPackage = expandMssPackage(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("manifest_name"); ok {
			in.ManifestName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("origination"); ok {
			in.Origination = types.Origination(v.(string))
		}

		if v, ok := d.GetOk("start_over_window_seconds"); ok {
			in.StartoverWindowSeconds = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("time_delay_seconds"); ok {
			in.TimeDelaySeconds = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("whitelist"); ok {
			in.Whitelist = flex.ExpandStringValueList(v.([]interface{}))
		}

		log.Printf("[DEBUG] Updating MediaPackage OriginEndpoint (%s): %#v", d.Id(), in)
		_, err := conn.UpdateOriginEndpoint(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionUpdating, ResNameOriginEndpoint, d.Id(), err)
		}
	}

	return append(diags, resourceOriginEndpointRead(ctx, d, meta)...)
}

func resourceOriginEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	log.Printf("[INFO] Deleting MediaPackage OriginEndpoint %s", d.Id())

	_, err := conn.DeleteOriginEndpoint(ctx, &mediapackage.DeleteOriginEndpointInput{
		Id: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionDeleting, ResNameOriginEndpoint, d.Id(), err)
	}

	return diags
}

func findOriginEndpoint(ctx context.Context, conn *mediapackage.Client, id, channelID string) (*types.OriginEndpoint, error) {
	in := &mediapackage.ListOriginEndpointsInput{
		ChannelId: aws.String(channelID),
	}

	out, err := conn.ListOriginEndpoints(ctx, in)
	if err != nil {
		return nil, err
	}

	if len(out.OriginEndpoints) == 0 {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	var ep *types.OriginEndpoint

	for _, e := range out.OriginEndpoints {
		if aws.ToString(e.Id) == id {
			ep = &e
		}
	}

	return ep, nil
}

func expandAuthorization(tfMap map[string]interface{}) *types.Authorization {
	if tfMap == nil {
		return nil
	}

	a := &types.Authorization{}

	if v, ok := tfMap["cdn_identifier_secret"].(string); ok && v != "" {
		a.CdnIdentifierSecret = aws.String(v)
	}

	if v, ok := tfMap["secrets_role_arn"].(string); ok && v != "" {
		a.SecretsRoleArn = aws.String(v)
	}

	return a
}

func flattenAuthorization(apiObject *types.Authorization) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.CdnIdentifierSecret; v != nil {
		m["cdn_identifier_secret"] = aws.ToString(v)
	}

	if v := apiObject.SecretsRoleArn; v != nil {
		m["secrets_role_arn"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func expandHlsPackage(tfMap map[string]interface{}) *types.HlsPackage {
	if tfMap == nil {
		return nil
	}

	h := &types.HlsPackage{}

	if v, ok := tfMap["ad_markers"].(string); ok && v != "" {
		h.AdMarkers = types.AdMarkers(v)
	}

	if v, ok := tfMap["ad_triggers"].([]interface{}); ok && len(v) > 0 {
		h.AdTriggers = expandAdTriggers(v)
	}

	if v, ok := tfMap["ads_on_delivery_restrictions"].(string); ok && v != "" {
		h.AdsOnDeliveryRestrictions = types.AdsOnDeliveryRestrictions(v)
	}

	if v, ok := tfMap["encryption"].([]interface{}); ok && len(v) > 0 {
		h.Encryption = expandHlsEncryption(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["include_dvb_subtitles"].(bool); ok {
		h.IncludeDvbSubtitles = aws.Bool(v)
	}

	if v, ok := tfMap["include_iframe_only_stream"].(bool); ok {
		h.IncludeIframeOnlyStream = aws.Bool(v)
	}

	if v, ok := tfMap["playlist_type"].(string); ok && v != "" {
		h.PlaylistType = types.PlaylistType(v)
	}

	if v, ok := tfMap["playlist_window_seconds"].(int); ok {
		h.PlaylistWindowSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["program_date_time_interval_seconds"].(int); ok {
		h.ProgramDateTimeIntervalSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["segment_duration_seconds"].(int); ok {
		h.SegmentDurationSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["stream_selection"].([]interface{}); ok && len(v) > 0 {
		h.StreamSelection = expandStreamSelection(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["use_audio_rendition_group"].(bool); ok {
		h.UseAudioRenditionGroup = aws.Bool(v)
	}

	return h
}

func flattenHlsPackage(apiObject *types.HlsPackage) []interface{} {
	if apiObject == nil {
		return nil
	}
	m := map[string]interface{}{}

	if v := apiObject.AdMarkers; v != "" {
		m["ad_markers"] = string(v)
	}

	if v := apiObject.AdTriggers; len(v) > 0 {
		m["ad_triggers"] = flex.FlattenStringyValueList(v)
	}

	if v := apiObject.AdsOnDeliveryRestrictions; v != "" {
		m["ads_on_delivery_restrictions"] = string(v)
	}

	if v := apiObject.Encryption; v != nil {
		m["encryption"] = flattenHlsEncryption(v)
	}

	if v := apiObject.IncludeDvbSubtitles; v != nil {
		m["include_dvb_subtitles"] = aws.ToBool(v)
	}

	if v := apiObject.IncludeIframeOnlyStream; v != nil {
		m["include_iframe_only_stream"] = aws.ToBool(v)
	}

	if v := apiObject.PlaylistType; v != "" {
		m["playlist_type"] = string(v)
	}

	if v := apiObject.PlaylistWindowSeconds; v != nil {
		m["playlist_window_seconds"] = int(aws.ToInt32(v))
	}

	if v := apiObject.ProgramDateTimeIntervalSeconds; v != nil {
		m["program_date_time_interval_seconds"] = int(aws.ToInt32(v))
	}

	if v := apiObject.SegmentDurationSeconds; v != nil {
		m["segment_duration_seconds"] = int(aws.ToInt32(v))
	}

	if v := apiObject.StreamSelection; v != nil {
		m["stream_selection"] = flattenStreamSelection(v)
	}

	if v := apiObject.UseAudioRenditionGroup; v != nil {
		m["use_audio_rendition_group"] = aws.ToBool(v)
	}

	return []interface{}{m}
}

func expandAdTriggers(tfList []interface{}) []types.AdTriggersElement {
	if len(tfList) == 0 {
		return nil
	}

	var ads []types.AdTriggersElement

	for _, elm := range tfList {
		ads = append(ads, types.AdTriggersElement(elm.(string)))
	}

	return ads
}

func expandHlsEncryption(tfMap map[string]interface{}) *types.HlsEncryption {
	if tfMap == nil {
		return nil
	}

	h := &types.HlsEncryption{}

	if v, ok := tfMap["speke_key_provider"].([]interface{}); ok && len(v) > 0 {
		h.SpekeKeyProvider = expandSpekeKeyProvider(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["constant_initialization_vector"].(string); ok && v != "" {
		h.ConstantInitializationVector = aws.String(v)
	}

	if v, ok := tfMap["encryption_method"].(string); ok && v != "" {
		h.EncryptionMethod = types.EncryptionMethod(v)
	}

	if v, ok := tfMap["key_rotation_interval_seconds"].(int); ok {
		h.KeyRotationIntervalSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["repeat_ext_x_key"].(bool); ok {
		h.RepeatExtXKey = aws.Bool(v)
	}

	return h
}

func flattenHlsEncryption(apiObject *types.HlsEncryption) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SpekeKeyProvider; v != nil {
		m["speke_key_provider"] = flattenSpekeKeyProvider(v)
	}

	if v := apiObject.ConstantInitializationVector; v != nil {
		m["constant_initialization_vector"] = aws.ToString(v)
	}

	if v := apiObject.EncryptionMethod; v != "" {
		m["encryption_method"] = string(v)
	}

	if v := apiObject.KeyRotationIntervalSeconds; v != nil {
		m["key_rotation_interval_seconds"] = int(*v)
	}

	if v := apiObject.RepeatExtXKey; v != nil {
		m["repeat_ext_x_key"] = aws.ToBool(v)
	}

	return []interface{}{m}
}

func expandSpekeKeyProvider(tfMap map[string]interface{}) *types.SpekeKeyProvider {
	s := &types.SpekeKeyProvider{
		ResourceId: aws.String(tfMap["resource_id"].(string)),
		RoleArn:    aws.String(tfMap["role_arn"].(string)),
		Url:        aws.String(tfMap["url"].(string)),
		SystemIds:  flex.ExpandStringValueList(tfMap["system_ids"].([]interface{})),
	}

	if v, ok := tfMap["certificate_arn"].(string); ok && v != "" {
		s.CertificateArn = aws.String(v)
	}

	if v, ok := tfMap["encryption_contract_configuration"].([]interface{}); ok && len(v) > 0 {
		c := &types.EncryptionContractConfiguration{
			PresetSpeke20Audio: types.PresetSpeke20Audio(v[0].(map[string]interface{})["preset_speke20_audio"].(string)),
			PresetSpeke20Video: types.PresetSpeke20Video(v[0].(map[string]interface{})["preset_speke20_video"].(string)),
		}
		s.EncryptionContractConfiguration = c
	}

	return s
}

func flattenSpekeKeyProvider(apiObject *types.SpekeKeyProvider) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.ResourceId; v != nil {
		m["resource_id"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		m["role_arn"] = aws.ToString(v)
	}

	if v := apiObject.Url; v != nil {
		m["url"] = aws.ToString(v)
	}

	if len(apiObject.SystemIds) > 0 {
		systemIDs := make([]interface{}, 0, len(apiObject.SystemIds))
		for _, sid := range apiObject.SystemIds {
			systemIDs = append(systemIDs, sid)
		}
		m["system_ids"] = systemIDs
	}

	if v := apiObject.CertificateArn; v != nil {
		m["certificate_arn"] = aws.ToString(v)
	}

	if v := apiObject.EncryptionContractConfiguration; v != nil {
		em := map[string]interface{}{}

		if v.PresetSpeke20Audio != "" {
			em["preset_speke20_audio"] = string(v.PresetSpeke20Audio)
		}
		if v.PresetSpeke20Video != "" {
			em["preset_speke20_video"] = string(v.PresetSpeke20Video)
		}

		m["encryption_contract_configuration"] = []interface{}{em}
	}

	return []interface{}{m}
}

func expandStreamSelection(tfMap map[string]interface{}) *types.StreamSelection {
	s := &types.StreamSelection{}

	if v, ok := tfMap["max_video_bits_per_second"].(int); ok {
		s.MaxVideoBitsPerSecond = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_video_bits_per_second"].(int); ok {
		s.MinVideoBitsPerSecond = aws.Int32(int32(v))
	}

	if v, ok := tfMap["stream_order"].(string); ok {
		s.StreamOrder = types.StreamOrder(v)
	}

	return s
}

func flattenStreamSelection(apiObject *types.StreamSelection) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.MaxVideoBitsPerSecond; v != nil {
		m["max_video_bits_per_second"] = int(aws.ToInt32(v))
	}

	if v := apiObject.MinVideoBitsPerSecond; v != nil {
		m["min_video_bits_per_second"] = int(aws.ToInt32(v))
	}

	if v := apiObject.StreamOrder; v != "" {
		m["stream_order"] = string(v)
	}

	return []interface{}{m}
}

func expandMssPackage(tfMap map[string]interface{}) *types.MssPackage {
	m := &types.MssPackage{}

	if v, ok := tfMap["encryption"].([]interface{}); ok && len(v) > 0 {
		m.Encryption = expandMssEncryption(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["manifest_window_seconds"].(int); ok {
		m.ManifestWindowSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["segment_duration_seconds"].(int); ok {
		m.SegmentDurationSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["stream_selection"].([]interface{}); ok && len(v) > 0 {
		m.StreamSelection = expandStreamSelection(v[0].(map[string]interface{}))
	}

	return m
}

func flattenMssPackage(apiObject *types.MssPackage) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Encryption; v != nil {
		m["encryption"] = flattenMssEncryption(v)
	}

	if v := apiObject.ManifestWindowSeconds; v != nil {
		m["manifest_window_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.SegmentDurationSeconds; v != nil {
		m["segment_duration_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.StreamSelection; v != nil {
		m["stream_selection"] = flattenStreamSelection(v)
	}

	return []interface{}{m}
}

func expandMssEncryption(mssEncryptionSettings map[string]interface{}) *types.MssEncryption {
	m := &types.MssEncryption{}

	if v, ok := mssEncryptionSettings["speke_key_provider"].([]interface{}); ok && len(v) > 0 {
		m.SpekeKeyProvider = expandSpekeKeyProvider(v[0].(map[string]interface{}))
	}

	return m
}

func flattenMssEncryption(apiObject *types.MssEncryption) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SpekeKeyProvider; v != nil {
		m["speke_key_provider"] = flattenSpekeKeyProvider(v)
	}

	return []interface{}{m}
}

func expandCmafPackage(cmafPackageSettings map[string]interface{}) *types.CmafPackageCreateOrUpdateParameters {
	c := &types.CmafPackageCreateOrUpdateParameters{}

	if v, ok := cmafPackageSettings["encryption"].([]interface{}); ok && len(v) > 0 {
		c.Encryption = expandCmafEncryption(v[0].(map[string]interface{}))
	}

	if v, ok := cmafPackageSettings["hls_manifests"].([]interface{}); ok && len(v) > 0 {
		c.HlsManifests = expandHlsManifests(v)
	}

	if v, ok := cmafPackageSettings["segment_duration_seconds"].(int); ok && v > 0 {
		c.SegmentDurationSeconds = aws.Int32(int32(v))
	}

	if v, ok := cmafPackageSettings["segment_prefix"].(string); ok && v != "" {
		c.SegmentPrefix = aws.String(v)
	}

	if v, ok := cmafPackageSettings["stream_selection"].([]interface{}); ok && len(v) > 0 {
		c.StreamSelection = expandStreamSelection(v[0].(map[string]interface{}))
	}

	return c
}

func flattenCmafPackage(apiObject *types.CmafPackage) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Encryption; v != nil {
		m["encryption"] = flattenCmafEncryption(v)
	}

	if v := apiObject.HlsManifests; v != nil {
		m["hls_manifests"] = flattenHlsManifests(v)
	}

	if v := apiObject.SegmentDurationSeconds; v != nil {
		m["segment_duration_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.SegmentPrefix; v != nil {
		m["segment_prefix"] = aws.ToString(v)
	}

	if v := apiObject.StreamSelection; v != nil {
		m["stream_selection"] = flattenStreamSelection(v)
	}

	return []interface{}{m}
}

func expandDashPackage(tfMap map[string]interface{}) *types.DashPackage {
	if tfMap == nil {
		return nil
	}

	d := &types.DashPackage{}

	if v, ok := tfMap["ad_triggers"].([]interface{}); ok && len(v) > 0 {
		d.AdTriggers = expandAdTriggers(v)
	}

	if v, ok := tfMap["ads_on_delivery_restrictions"].(string); ok && v != "" {
		d.AdsOnDeliveryRestrictions = types.AdsOnDeliveryRestrictions(v)
	}

	if v, ok := tfMap["encryption"].([]interface{}); ok && len(v) > 0 {
		d.Encryption = expandDashEncryption(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["include_iframe_only_stream"].(bool); ok {
		d.IncludeIframeOnlyStream = aws.Bool(v)
	}

	if v, ok := tfMap["manifest_layout"].(string); ok && v != "" {
		d.ManifestLayout = types.ManifestLayout(v)
	}

	if v, ok := tfMap["manifest_window_seconds"].(int); ok && v > 0 {
		d.ManifestWindowSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_buffer_time_seconds"].(int); ok && v > 0 {
		d.MinBufferTimeSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_update_period_seconds"].(int); ok && v > 0 {
		d.MinUpdatePeriodSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["period_triggers"].([]interface{}); ok && len(v) > 0 {
		d.PeriodTriggers = expandPeriodTriggers(v)
	}

	if v, ok := tfMap["profile"].(string); ok && v != "" {
		d.Profile = types.Profile(v)
	}

	if v, ok := tfMap["segment_duration_seconds"].(int); ok && v > 0 {
		d.SegmentDurationSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["stream_selection"].([]interface{}); ok && len(v) > 0 {
		d.StreamSelection = expandStreamSelection(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["suggested_presentation_delay_seconds"].(int); ok && v > 0 {
		d.SuggestedPresentationDelaySeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["utc_timing"].(string); ok && v != "" {
		d.UtcTiming = types.UtcTiming(v)
	}

	if v, ok := tfMap["utc_timing_uri"].(string); ok && v != "" {
		d.UtcTimingUri = aws.String(v)
	}

	return d
}

func flattenDashPackage(apiObject *types.DashPackage) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AdTriggers; len(v) > 0 {
		m["ad_triggers"] = flex.FlattenStringyValueList(v)
	}

	if v := apiObject.AdsOnDeliveryRestrictions; v != "" {
		m["ads_on_delivery_restrictions"] = string(v)
	}

	if v := apiObject.Encryption; v != nil {
		m["encryption"] = flattenDashEncryption(v)
	}

	if v := apiObject.IncludeIframeOnlyStream; v != nil {
		m["include_iframe_only_stream"] = aws.ToBool(v)
	}

	if v := apiObject.ManifestLayout; v != "" {
		m["manifest_layout"] = string(v)
	}

	if v := apiObject.ManifestWindowSeconds; v != nil {
		m["manifest_window_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.MinBufferTimeSeconds; v != nil {
		m["min_buffer_time_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.MinUpdatePeriodSeconds; v != nil {
		m["min_update_period_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.PeriodTriggers; len(v) > 0 {
		m["period_triggers"] = flex.FlattenStringyValueList(v)
	}

	if v := apiObject.Profile; v != "" {
		m["profile"] = string(v)
	}

	if v := apiObject.SegmentDurationSeconds; v != nil {
		m["segment_duration_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.StreamSelection; v != nil {
		m["stream_selection"] = flattenStreamSelection(v)
	}

	if v := apiObject.SuggestedPresentationDelaySeconds; v != nil {
		m["suggested_presentation_delay_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.UtcTiming; v != "" {
		m["utc_timing"] = string(v)
	}

	if v := apiObject.UtcTimingUri; v != nil {
		m["utc_timing_uri"] = v
	}

	return []interface{}{m}
}

func expandDashEncryption(tfMap map[string]interface{}) *types.DashEncryption {
	if tfMap == nil {
		return nil
	}

	d := &types.DashEncryption{}

	if v, ok := tfMap["speke_key_provider"].([]interface{}); ok && len(v) > 0 {
		d.SpekeKeyProvider = expandSpekeKeyProvider(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["key_rotation_interval_seconds"].(int); ok {
		d.KeyRotationIntervalSeconds = aws.Int32(int32(v))
	}

	return d
}

func flattenDashEncryption(apiObject *types.DashEncryption) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SpekeKeyProvider; v != nil {
		m["speke_key_provider"] = flattenSpekeKeyProvider(v)
	}

	if v := apiObject.KeyRotationIntervalSeconds; v != nil {
		m["key_rotation_interval_seconds"] = aws.ToInt32(v)
	}

	return []interface{}{m}
}

func expandPeriodTriggers(tfList []interface{}) []types.PeriodTriggersElement {
	if len(tfList) == 0 {
		return nil
	}

	var l []types.PeriodTriggersElement

	for _, r := range tfList {
		l = append(l, types.PeriodTriggersElement(r.(string)))
	}

	return l
}

func expandCmafEncryption(tfMap map[string]interface{}) *types.CmafEncryption {
	if tfMap == nil {
		return nil
	}

	e := &types.CmafEncryption{}

	if v, ok := tfMap["speke_key_provider"].([]interface{}); ok && len(v) > 0 {
		e.SpekeKeyProvider = expandSpekeKeyProvider(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["constant_initialization_vector"].(string); ok && v != "" {
		e.ConstantInitializationVector = aws.String(v)
	}

	if v, ok := tfMap["encryption_method"].(string); ok && v != "" {
		e.EncryptionMethod = types.CmafEncryptionMethod(v)
	}

	if v, ok := tfMap["key_rotation_interval_seconds"].(int); ok {
		e.KeyRotationIntervalSeconds = aws.Int32(int32(v))
	}

	return e
}

func flattenCmafEncryption(apiObject *types.CmafEncryption) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SpekeKeyProvider; v != nil {
		m["speke_key_provider"] = flattenSpekeKeyProvider(v)
	}

	if v := apiObject.ConstantInitializationVector; v != nil {
		m["constant_initialization_vector"] = aws.ToString(v)
	}

	if v := apiObject.EncryptionMethod; v != "" {
		m["encryption_method"] = string(v)
	}

	if v := apiObject.KeyRotationIntervalSeconds; v != nil {
		m["key_rotation_interval_seconds"] = aws.ToInt32(v)
	}

	return []interface{}{m}
}

func expandHlsManifests(tfList []interface{}) []types.HlsManifestCreateOrUpdateParameters {
	if len(tfList) == 0 {
		return nil
	}

	var hs []types.HlsManifestCreateOrUpdateParameters

	for _, tfManifest := range tfList {
		manifest := tfManifest.(map[string]interface{})
		m := types.HlsManifestCreateOrUpdateParameters{
			Id: aws.String(manifest["id"].(string)),
		}

		if v, ok := manifest["ad_markers"].(string); ok && v != "" {
			m.AdMarkers = types.AdMarkers(v)
		}

		if v, ok := manifest["ad_triggers"].([]interface{}); ok && len(v) > 0 {
			m.AdTriggers = expandAdTriggers(v)
		}

		if v, ok := manifest["ads_on_delivery_restrictions"].(string); ok && v != "" {
			m.AdsOnDeliveryRestrictions = types.AdsOnDeliveryRestrictions(v)
		}

		if v, ok := manifest["include_iframe_only_stream"].(bool); ok {
			m.IncludeIframeOnlyStream = aws.Bool(v)
		}

		if v, ok := manifest["manifest_name"].(string); ok && v != "" {
			m.ManifestName = aws.String(v)
		}

		if v, ok := manifest["playlist_type"].(string); ok && v != "" {
			m.PlaylistType = types.PlaylistType(v)
		}

		if v, ok := manifest["playlist_window_seconds"].(int); ok && v > 0 {
			m.PlaylistWindowSeconds = aws.Int32(int32(v))
		}

		if v, ok := manifest["program_date_time_interval_seconds"].(int); ok && v > 0 {
			m.ProgramDateTimeIntervalSeconds = aws.Int32(int32(v))
		}

		hs = append(hs, m)
	}

	return hs
}

func flattenHlsManifests(apiList []types.HlsManifest) []interface{} {
	if len(apiList) == 0 {
		return nil
	}

	var hs []interface{}

	for _, manifest := range apiList {
		m := map[string]interface{}{
			"id": manifest.Id,
		}

		if v := manifest.AdMarkers; v != "" {
			m["ad_markers"] = string(v)
		}

		if v := manifest.AdTriggers; len(v) > 0 {
			m["ad_triggers"] = flex.FlattenStringyValueList(v)
		}

		if v := manifest.AdsOnDeliveryRestrictions; v != "" {
			m["ads_on_delivery_restrictions"] = string(v)
		}

		if v := manifest.IncludeIframeOnlyStream; v != nil {
			m["include_iframe_only_stream"] = aws.ToBool(v)
		}

		if v := manifest.ManifestName; v != nil {
			m["manifest_name"] = aws.ToString(v)
		}

		if v := manifest.PlaylistType; v != "" {
			m["playlist_type"] = string(v)
		}

		if v := manifest.PlaylistWindowSeconds; v != nil {
			m["playlist_window_seconds"] = aws.ToInt32(v)
		}

		if v := manifest.ProgramDateTimeIntervalSeconds; v != nil {
			m["program_date_time_interval_seconds"] = aws.ToInt32(v)
		}

		hs = append(hs, m)
	}

	return hs
}

var spekeKeyProviderSchema = map[string]*schema.Schema{
	"resource_id": {
		Type:     schema.TypeString,
		Required: true,
	},
	"role_arn": {
		Type:     schema.TypeString,
		Required: true,
	},
	"system_ids": {
		Type:     schema.TypeList,
		Required: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	},
	"url": {
		Type:     schema.TypeString,
		Required: true,
	},
	"certificate_arn": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"encryption_contract_configuration": {
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"preset_speke20_audio": {
					Type:     schema.TypeString,
					Required: true,
				},
				"preset_speke20_video": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	},
}

var streamSelectionSchema = map[string]*schema.Schema{
	"max_video_bits_per_second": {
		Type:     schema.TypeInt,
		Optional: true,
	},
	"min_video_bits_per_second": {
		Type:     schema.TypeInt,
		Optional: true,
	},
	"stream_order": {
		Type:     schema.TypeString,
		Optional: true,
	},
}
