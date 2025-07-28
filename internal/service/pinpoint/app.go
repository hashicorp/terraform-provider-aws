// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpoint/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_pinpoint_app", name="App")
// @Tags(identifierAttribute="arn")
func resourceApp() *schema.Resource {
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.Mode](),
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
	}
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &pinpoint.CreateAppInput{
		CreateApplicationRequest: &awstypes.CreateApplicationRequest{
			Name: aws.String(name),
			Tags: getTagsIn(ctx),
		},
	}

	output, err := conn.CreateApp(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Pinpoint App (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ApplicationResponse.Id))

	return append(diags, resourceAppUpdate(ctx, d, meta)...)
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	app, err := findAppByID(ctx, conn, d.Id())

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
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(app.Name)))
	if err := d.Set("quiet_time", flattenQuietTime(settings.QuietTime)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting quiet_time: %s", err)
	}

	return diags
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		appSettings := &awstypes.WriteApplicationSettingsRequest{}

		if d.HasChange("campaign_hook") {
			appSettings.CampaignHook = expandCampaignHook(d.Get("campaign_hook").([]any))
		}

		//if d.HasChange("cloudwatch_metrics_enabled") {
		//	appSettings.CloudWatchMetricsEnabled = aws.Bool(d.Get("cloudwatch_metrics_enabled").(bool));
		//}

		if d.HasChange("limits") {
			appSettings.Limits = expandCampaignLimits(d.Get("limits").([]any))
		}

		if d.HasChange("quiet_time") {
			appSettings.QuietTime = expandQuietTime(d.Get("quiet_time").([]any))
		}

		input := &pinpoint.UpdateApplicationSettingsInput{
			ApplicationId:                   aws.String(d.Id()),
			WriteApplicationSettingsRequest: appSettings,
		}

		_, err := conn.UpdateApplicationSettings(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Pinpoint App (%s) settings: %s", d.Id(), err)
		}
	}

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[DEBUG] Deleting Pinpoint App: %s", d.Id())
	_, err := conn.DeleteApp(ctx, &pinpoint.DeleteAppInput{
		ApplicationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint App (%s): %s", d.Id(), err)
	}

	return diags
}

func findAppByID(ctx context.Context, conn *pinpoint.Client, id string) (*awstypes.ApplicationResponse, error) {
	input := &pinpoint.GetAppInput{
		ApplicationId: aws.String(id),
	}

	output, err := conn.GetApp(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func findAppSettingsByID(ctx context.Context, conn *pinpoint.Client, id string) (*awstypes.ApplicationSettingsResource, error) {
	input := &pinpoint.GetApplicationSettingsInput{
		ApplicationId: aws.String(id),
	}

	output, err := conn.GetApplicationSettings(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func expandCampaignHook(configs []any) *awstypes.CampaignHook {
	if len(configs) == 0 || configs[0] == nil {
		return nil
	}

	m := configs[0].(map[string]any)

	ch := &awstypes.CampaignHook{}

	if v, ok := m["lambda_function_name"]; ok {
		ch.LambdaFunctionName = aws.String(v.(string))
	}

	if v, ok := m[names.AttrMode]; ok {
		ch.Mode = awstypes.Mode(v.(string))
	}

	if v, ok := m["web_url"]; ok {
		ch.WebUrl = aws.String(v.(string))
	}

	return ch
}

func flattenCampaignHook(ch *awstypes.CampaignHook) []any {
	l := make([]any, 0)

	m := map[string]any{}

	m["lambda_function_name"] = aws.ToString(ch.LambdaFunctionName)
	m[names.AttrMode] = ch.Mode
	m["web_url"] = aws.ToString(ch.WebUrl)

	l = append(l, m)

	return l
}

func expandCampaignLimits(configs []any) *awstypes.CampaignLimits {
	if len(configs) == 0 || configs[0] == nil {
		return nil
	}

	m := configs[0].(map[string]any)

	cl := awstypes.CampaignLimits{}

	if v, ok := m["daily"]; ok {
		cl.Daily = aws.Int32(int32(v.(int)))
	}

	if v, ok := m["maximum_duration"]; ok {
		cl.MaximumDuration = aws.Int32(int32(v.(int)))
	}

	if v, ok := m["messages_per_second"]; ok {
		cl.MessagesPerSecond = aws.Int32(int32(v.(int)))
	}

	if v, ok := m["total"]; ok {
		cl.Total = aws.Int32(int32(v.(int)))
	}

	return &cl
}

func flattenCampaignLimits(cl *awstypes.CampaignLimits) []any {
	l := make([]any, 0)

	m := map[string]any{}

	m["daily"] = aws.ToInt32(cl.Daily)
	m["maximum_duration"] = aws.ToInt32(cl.MaximumDuration)
	m["messages_per_second"] = aws.ToInt32(cl.MessagesPerSecond)
	m["total"] = aws.ToInt32(cl.Total)

	l = append(l, m)

	return l
}

func expandQuietTime(configs []any) *awstypes.QuietTime {
	if len(configs) == 0 || configs[0] == nil {
		return nil
	}

	m := configs[0].(map[string]any)

	qt := awstypes.QuietTime{}

	if v, ok := m["end"]; ok {
		qt.End = aws.String(v.(string))
	}

	if v, ok := m["start"]; ok {
		qt.Start = aws.String(v.(string))
	}

	return &qt
}

func flattenQuietTime(qt *awstypes.QuietTime) []any {
	l := make([]any, 0)

	m := map[string]any{}

	m["end"] = aws.ToString(qt.End)
	m["start"] = aws.ToString(qt.Start)

	l = append(l, m)

	return l
}
