// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_pinpoint_app", name="App")
// @Tags(identifierAttribute="arn")
func ResourceApp() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAppCreate,
		ReadWithoutTimeout:   resourceAppRead,
		UpdateWithoutTimeout: resourceAppUpdate,
		DeleteWithoutTimeout: resourceAppDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"application_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			//"cloudwatch_metrics_enabled": {
			//	Type:     schema.TypeBool,
			//	Optional: true,
			//	Default:  false,
			//},
			"campaign_hook": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lambda_function_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								pinpoint.ModeDelivery,
								pinpoint.ModeFilter,
							}, false),
						},
						"web_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"limits": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"daily": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100),
						},
						"maximum_duration": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(60),
						},
						"messages_per_second": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(50, 20000),
						},
						"total": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100),
						},
					},
				},
			},
			"quiet_time": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"end": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"start": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	var name string

	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = id.PrefixedUniqueId(v.(string))
	} else {
		name = id.UniqueId()
	}

	log.Printf("[DEBUG] Pinpoint create app: %s", name)

	req := &pinpoint.CreateAppInput{
		CreateApplicationRequest: &pinpoint.CreateApplicationRequest{
			Name: aws.String(name),
			Tags: getTagsIn(ctx),
		},
	}

	output, err := conn.CreateAppWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Pinpoint app: %s", err)
	}

	d.SetId(aws.StringValue(output.ApplicationResponse.Id))
	d.Set("arn", output.ApplicationResponse.Arn)

	return append(diags, resourceAppUpdate(ctx, d, meta)...)
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	log.Printf("[INFO] Reading Pinpoint App Attributes for %s", d.Id())

	app, err := conn.GetAppWithContext(ctx, &pinpoint.GetAppInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
			log.Printf("[WARN] Pinpoint App (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading Pinpoint Application (%s): %s", d.Id(), err)
	}

	settings, err := conn.GetApplicationSettingsWithContext(ctx, &pinpoint.GetApplicationSettingsInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
			log.Printf("[WARN] Pinpoint App (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading Pinpoint Application (%s) settings: %s", d.Id(), err)
	}

	arn := aws.StringValue(app.ApplicationResponse.Arn)
	d.Set("name", app.ApplicationResponse.Name)
	d.Set("application_id", app.ApplicationResponse.Id)
	d.Set("arn", arn)

	if err := d.Set("campaign_hook", flattenCampaignHook(settings.ApplicationSettingsResource.CampaignHook)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting campaign_hook: %s", err)
	}
	if err := d.Set("limits", flattenCampaignLimits(settings.ApplicationSettingsResource.Limits)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting limits: %s", err)
	}
	if err := d.Set("quiet_time", flattenQuietTime(settings.ApplicationSettingsResource.QuietTime)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting quiet_time: %s", err)
	}

	return diags
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	appSettings := &pinpoint.WriteApplicationSettingsRequest{}

	//if d.HasChange("cloudwatch_metrics_enabled") {
	//	appSettings.CloudWatchMetricsEnabled = aws.Bool(d.Get("cloudwatch_metrics_enabled").(bool));
	//}

	if d.HasChange("campaign_hook") {
		appSettings.CampaignHook = expandCampaignHook(d.Get("campaign_hook").([]interface{}))
	}

	if d.HasChange("limits") {
		appSettings.Limits = expandCampaignLimits(d.Get("limits").([]interface{}))
	}

	if d.HasChange("quiet_time") {
		appSettings.QuietTime = expandQuietTime(d.Get("quiet_time").([]interface{}))
	}

	req := pinpoint.UpdateApplicationSettingsInput{
		ApplicationId:                   aws.String(d.Id()),
		WriteApplicationSettingsRequest: appSettings,
	}

	_, err := conn.UpdateApplicationSettingsWithContext(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Pinpoint Application (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	log.Printf("[DEBUG] Deleting Pinpoint Application: %s", d.Id())
	_, err := conn.DeleteAppWithContext(ctx, &pinpoint.DeleteAppInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint Application (%s): %s", d.Id(), err)
	}

	return diags
}

func expandCampaignHook(configs []interface{}) *pinpoint.CampaignHook {
	if len(configs) == 0 {
		return nil
	}

	m := configs[0].(map[string]interface{})

	ch := &pinpoint.CampaignHook{}

	if v, ok := m["lambda_function_name"]; ok {
		ch.LambdaFunctionName = aws.String(v.(string))
	}

	if v, ok := m["mode"]; ok {
		ch.Mode = aws.String(v.(string))
	}

	if v, ok := m["web_url"]; ok {
		ch.WebUrl = aws.String(v.(string))
	}

	return ch
}

func flattenCampaignHook(ch *pinpoint.CampaignHook) []interface{} {
	l := make([]interface{}, 0)

	m := map[string]interface{}{}

	m["lambda_function_name"] = aws.StringValue(ch.LambdaFunctionName)
	m["mode"] = aws.StringValue(ch.Mode)
	m["web_url"] = aws.StringValue(ch.WebUrl)

	l = append(l, m)

	return l
}

func expandCampaignLimits(configs []interface{}) *pinpoint.CampaignLimits {
	if len(configs) == 0 {
		return nil
	}

	m := configs[0].(map[string]interface{})

	cl := pinpoint.CampaignLimits{}

	if v, ok := m["daily"]; ok {
		cl.Daily = aws.Int64(int64(v.(int)))
	}

	if v, ok := m["maximum_duration"]; ok {
		cl.MaximumDuration = aws.Int64(int64(v.(int)))
	}

	if v, ok := m["messages_per_second"]; ok {
		cl.MessagesPerSecond = aws.Int64(int64(v.(int)))
	}

	if v, ok := m["total"]; ok {
		cl.Total = aws.Int64(int64(v.(int)))
	}

	return &cl
}

func flattenCampaignLimits(cl *pinpoint.CampaignLimits) []interface{} {
	l := make([]interface{}, 0)

	m := map[string]interface{}{}

	m["daily"] = aws.Int64Value(cl.Daily)
	m["maximum_duration"] = aws.Int64Value(cl.MaximumDuration)
	m["messages_per_second"] = aws.Int64Value(cl.MessagesPerSecond)
	m["total"] = aws.Int64Value(cl.Total)

	l = append(l, m)

	return l
}

func expandQuietTime(configs []interface{}) *pinpoint.QuietTime {
	if len(configs) == 0 {
		return nil
	}

	m := configs[0].(map[string]interface{})

	qt := pinpoint.QuietTime{}

	if v, ok := m["end"]; ok {
		qt.End = aws.String(v.(string))
	}

	if v, ok := m["start"]; ok {
		qt.Start = aws.String(v.(string))
	}

	return &qt
}

func flattenQuietTime(qt *pinpoint.QuietTime) []interface{} {
	l := make([]interface{}, 0)

	m := map[string]interface{}{}

	m["end"] = aws.StringValue(qt.End)
	m["start"] = aws.StringValue(qt.Start)

	l = append(l, m)

	return l
}
