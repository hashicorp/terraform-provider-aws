// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rum

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/rum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rum/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rum_app_monitor", name="App Monitor")
// @Tags(identifierAttribute="arn")
func resourceAppMonitor() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAppMonitorCreate,
		ReadWithoutTimeout:   resourceAppMonitorRead,
		UpdateWithoutTimeout: resourceAppMonitorUpdate,
		DeleteWithoutTimeout: resourceAppMonitorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"app_monitor_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_cookies": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_xray": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"excluded_pages": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"favorite_pages": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"guest_role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"identity_pool_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"included_pages": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"session_sample_rate": {
							Type:         schema.TypeFloat,
							Optional:     true,
							Default:      0.1,
							ValidateFunc: validation.FloatBetween(0, 1),
						},
						"telemetries": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[awstypes.Telemetry](),
							},
						},
					},
				},
			},
			"app_monitor_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_events": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrStatus: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.CustomEventsStatusDisabled,
							ValidateDiagFunc: enum.Validate[awstypes.CustomEventsStatus](),
						},
					},
				},
			},
			"cw_log_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cw_log_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomain: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 253),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAppMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RUMClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &rum.CreateAppMonitorInput{
		Name:         aws.String(name),
		CwLogEnabled: aws.Bool(d.Get("cw_log_enabled").(bool)),
		Domain:       aws.String(d.Get(names.AttrDomain).(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("app_monitor_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AppMonitorConfiguration = expandAppMonitorConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("custom_events"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CustomEvents = expandCustomEvents(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.CreateAppMonitor(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch RUM App Monitor (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceAppMonitorRead(ctx, d, meta)...)
}

func resourceAppMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RUMClient(ctx)

	appMon, err := findAppMonitorByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch RUM App Monitor %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch RUM App Monitor (%s): %s", d.Id(), err)
	}

	if err := d.Set("app_monitor_configuration", []interface{}{flattenAppMonitorConfiguration(appMon.AppMonitorConfiguration)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting app_monitor_configuration: %s", err)
	}

	if err := d.Set("custom_events", []interface{}{flattenCustomEvents(appMon.CustomEvents)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting custom_events: %s", err)
	}

	d.Set("app_monitor_id", appMon.Id)
	name := aws.ToString(appMon.Name)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "rum",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "appmonitor/" + name,
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("cw_log_enabled", appMon.DataStorage.CwLog.CwLogEnabled)
	d.Set("cw_log_group", appMon.DataStorage.CwLog.CwLogGroup)
	d.Set(names.AttrDomain, appMon.Domain)
	d.Set(names.AttrName, name)

	setTagsOut(ctx, appMon.Tags)

	return diags
}

func resourceAppMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RUMClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &rum.UpdateAppMonitorInput{
			Name: aws.String(d.Id()),
		}

		if d.HasChange("app_monitor_configuration") {
			input.AppMonitorConfiguration = expandAppMonitorConfiguration(d.Get("app_monitor_configuration").([]interface{})[0].(map[string]interface{}))
		}

		if d.HasChange("custom_events") {
			input.CustomEvents = expandCustomEvents(d.Get("custom_events").([]interface{})[0].(map[string]interface{}))
		}

		if d.HasChange("cw_log_enabled") {
			input.CwLogEnabled = aws.Bool(d.Get("cw_log_enabled").(bool))
		}

		if d.HasChange(names.AttrDomain) {
			input.Domain = aws.String(d.Get(names.AttrDomain).(string))
		}

		_, err := conn.UpdateAppMonitor(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudWatch RUM App Monitor (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAppMonitorRead(ctx, d, meta)...)
}

func resourceAppMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RUMClient(ctx)

	log.Printf("[DEBUG] Deleting CloudWatch RUM App Monitor: %s", d.Id())
	_, err := conn.DeleteAppMonitor(ctx, &rum.DeleteAppMonitorInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch RUM App Monitor (%s): %s", d.Id(), err)
	}

	return diags
}

func findAppMonitorByName(ctx context.Context, conn *rum.Client, name string) (*awstypes.AppMonitor, error) {
	input := &rum.GetAppMonitorInput{
		Name: aws.String(name),
	}

	output, err := conn.GetAppMonitor(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AppMonitor == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AppMonitor, nil
}

func expandAppMonitorConfiguration(tfMap map[string]interface{}) *awstypes.AppMonitorConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &awstypes.AppMonitorConfiguration{}

	if v, ok := tfMap["guest_role_arn"].(string); ok && v != "" {
		config.GuestRoleArn = aws.String(v)
	}

	if v, ok := tfMap["identity_pool_id"].(string); ok && v != "" {
		config.IdentityPoolId = aws.String(v)
	}

	if v, ok := tfMap["session_sample_rate"].(float64); ok {
		config.SessionSampleRate = v
	}

	if v, ok := tfMap["allow_cookies"].(bool); ok {
		config.AllowCookies = aws.Bool(v)
	}

	if v, ok := tfMap["enable_xray"].(bool); ok {
		config.EnableXRay = aws.Bool(v)
	}

	if v, ok := tfMap["excluded_pages"].(*schema.Set); ok && v.Len() > 0 {
		config.ExcludedPages = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["favorite_pages"].(*schema.Set); ok && v.Len() > 0 {
		config.FavoritePages = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["included_pages"].(*schema.Set); ok && v.Len() > 0 {
		config.IncludedPages = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["telemetries"].(*schema.Set); ok && v.Len() > 0 {
		config.Telemetries = flex.ExpandStringyValueSet[awstypes.Telemetry](v)
	}

	return config
}

func flattenAppMonitorConfiguration(apiObject *awstypes.AppMonitorConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.GuestRoleArn; v != nil {
		tfMap["guest_role_arn"] = aws.ToString(v)
	}

	if v := apiObject.IdentityPoolId; v != nil {
		tfMap["identity_pool_id"] = aws.ToString(v)
	}

	tfMap["session_sample_rate"] = apiObject.SessionSampleRate

	if v := apiObject.AllowCookies; v != nil {
		tfMap["allow_cookies"] = aws.ToBool(v)
	}

	if v := apiObject.EnableXRay; v != nil {
		tfMap["enable_xray"] = aws.ToBool(v)
	}

	if v := apiObject.Telemetries; v != nil {
		tfMap["telemetries"] = apiObject.Telemetries
	}

	if v := apiObject.IncludedPages; v != nil {
		tfMap["included_pages"] = apiObject.IncludedPages
	}

	if v := apiObject.FavoritePages; v != nil {
		tfMap["favorite_pages"] = apiObject.FavoritePages
	}

	if v := apiObject.ExcludedPages; v != nil {
		tfMap["excluded_pages"] = apiObject.ExcludedPages
	}

	return tfMap
}

func expandCustomEvents(tfMap map[string]interface{}) *awstypes.CustomEvents {
	if tfMap == nil {
		return nil
	}

	config := &awstypes.CustomEvents{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		config.Status = awstypes.CustomEventsStatus(v)
	}

	return config
}

func flattenCustomEvents(apiObject *awstypes.CustomEvents) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrStatus: apiObject.Status,
	}

	return tfMap
}
