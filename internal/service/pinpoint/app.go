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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			names.AttrApplicationID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"campaign_hook": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lambda_function_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrMode: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(pinpoint.Mode_Values(), false),
						},
						"web_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			//"cloudwatch_metrics_enabled": {
			//	Type:     schema.TypeBool,
			//	Optional: true,
			//	Default:  false,
			//},
			"limits": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
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
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
			},
			"quiet_time": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &pinpoint.CreateAppInput{
		CreateApplicationRequest: &pinpoint.CreateApplicationRequest{
			Name: aws.String(name),
			Tags: getTagsIn(ctx),
		},
	}

	output, err := conn.CreateAppWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Pinpoint App (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ApplicationResponse.Id))

	return append(diags, resourceAppUpdate(ctx, d, meta)...)
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	app, err := FindAppByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Pinpoint App (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Pinpoint App (%s): %s", d.Id(), err)
	}

	settings, err := findAppSettingsByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Pinpoint App (%s) settings: %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, app.Id)
	d.Set(names.AttrARN, app.Arn)
	if err := d.Set("campaign_hook", flattenCampaignHook(settings.CampaignHook)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting campaign_hook: %s", err)
	}
	if err := d.Set("limits", flattenCampaignLimits(settings.Limits)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting limits: %s", err)
	}
	d.Set(names.AttrName, app.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.StringValue(app.Name)))
	if err := d.Set("quiet_time", flattenQuietTime(settings.QuietTime)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting quiet_time: %s", err)
	}

	return diags
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		appSettings := &pinpoint.WriteApplicationSettingsRequest{}

		if d.HasChange("campaign_hook") {
			appSettings.CampaignHook = expandCampaignHook(d.Get("campaign_hook").([]interface{}))
		}

		//if d.HasChange("cloudwatch_metrics_enabled") {
		//	appSettings.CloudWatchMetricsEnabled = aws.Bool(d.Get("cloudwatch_metrics_enabled").(bool));
		//}

		if d.HasChange("limits") {
			appSettings.Limits = expandCampaignLimits(d.Get("limits").([]interface{}))
		}

		if d.HasChange("quiet_time") {
			appSettings.QuietTime = expandQuietTime(d.Get("quiet_time").([]interface{}))
		}

		input := &pinpoint.UpdateApplicationSettingsInput{
			ApplicationId:                   aws.String(d.Id()),
			WriteApplicationSettingsRequest: appSettings,
		}

		_, err := conn.UpdateApplicationSettingsWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Pinpoint App (%s) settings: %s", d.Id(), err)
		}
	}

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	log.Printf("[DEBUG] Deleting Pinpoint App: %s", d.Id())
	_, err := conn.DeleteAppWithContext(ctx, &pinpoint.DeleteAppInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint App (%s): %s", d.Id(), err)
	}

	return diags
}

func FindAppByID(ctx context.Context, conn *pinpoint.Pinpoint, id string) (*pinpoint.ApplicationResponse, error) {
	input := &pinpoint.GetAppInput{
		ApplicationId: aws.String(id),
	}

	output, err := conn.GetAppWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationResponse == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ApplicationResponse, nil
}

func findAppSettingsByID(ctx context.Context, conn *pinpoint.Pinpoint, id string) (*pinpoint.ApplicationSettingsResource, error) {
	input := &pinpoint.GetApplicationSettingsInput{
		ApplicationId: aws.String(id),
	}

	output, err := conn.GetApplicationSettingsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationSettingsResource == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ApplicationSettingsResource, nil
}

func expandCampaignHook(configs []interface{}) *pinpoint.CampaignHook {
	if len(configs) == 0 || configs[0] == nil {
		return nil
	}

	m := configs[0].(map[string]interface{})

	ch := &pinpoint.CampaignHook{}

	if v, ok := m["lambda_function_name"]; ok {
		ch.LambdaFunctionName = aws.String(v.(string))
	}

	if v, ok := m[names.AttrMode]; ok {
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
	m[names.AttrMode] = aws.StringValue(ch.Mode)
	m["web_url"] = aws.StringValue(ch.WebUrl)

	l = append(l, m)

	return l
}

func expandCampaignLimits(configs []interface{}) *pinpoint.CampaignLimits {
	if len(configs) == 0 || configs[0] == nil {
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
	if len(configs) == 0 || configs[0] == nil {
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
