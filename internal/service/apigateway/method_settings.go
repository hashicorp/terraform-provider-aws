// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_api_gateway_method_settings", name="Method Settings")
func resourceMethodSettings() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMethodSettingsUpdate,
		ReadWithoutTimeout:   resourceMethodSettingsRead,
		UpdateWithoutTimeout: resourceMethodSettingsUpdate,
		DeleteWithoutTimeout: resourceMethodSettingsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceMethodSettingsImport,
		},

		Schema: map[string]*schema.Schema{
			"method_path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cache_data_encrypted": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"cache_ttl_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"caching_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"data_trace_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"logging_level": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"OFF",
								"ERROR",
								"INFO",
							}, false),
						},
						"metrics_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"require_authorization_for_cache_control": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"throttling_burst_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  -1,
						},
						"throttling_rate_limit": {
							Type:     schema.TypeFloat,
							Optional: true,
							Default:  -1,
						},
						"unauthorized_cache_control_header_strategy": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.UnauthorizedCacheControlHeaderStrategy](),
							Computed:         true,
						},
					},
				},
			},
			"stage_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func flattenMethodSettings(apiObject *types.MethodSetting) []interface{} {
	if apiObject == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"metrics_enabled":                            apiObject.MetricsEnabled,
			"logging_level":                              apiObject.LoggingLevel,
			"data_trace_enabled":                         apiObject.DataTraceEnabled,
			"throttling_burst_limit":                     apiObject.ThrottlingBurstLimit,
			"throttling_rate_limit":                      apiObject.ThrottlingRateLimit,
			"caching_enabled":                            apiObject.CachingEnabled,
			"cache_ttl_in_seconds":                       apiObject.CacheTtlInSeconds,
			"cache_data_encrypted":                       apiObject.CacheDataEncrypted,
			"require_authorization_for_cache_control":    apiObject.RequireAuthorizationForCacheControl,
			"unauthorized_cache_control_header_strategy": apiObject.UnauthorizedCacheControlHeaderStrategy,
		},
	}
}

func resourceMethodSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	settings, err := findMethodSettingsByThreePartKey(ctx, conn, d.Get("rest_api_id").(string), d.Get("stage_name").(string), d.Get("method_path").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Method Settings (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Method Settings (%s): %s", d.Id(), err)
	}

	if err := d.Set("settings", flattenMethodSettings(settings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting settings: %s", err)
	}

	return diags
}

func resourceMethodSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	methodPath := d.Get("method_path").(string)
	prefix := fmt.Sprintf("/%s/", methodPath)

	ops := make([]types.PatchOperation, 0)
	if d.HasChange("settings.0.metrics_enabled") {
		ops = append(ops, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String(prefix + "metrics/enabled"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.metrics_enabled").(bool))),
		})
	}
	if d.HasChange("settings.0.logging_level") {
		ops = append(ops, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String(prefix + "logging/loglevel"),
			Value: aws.String(d.Get("settings.0.logging_level").(string)),
		})
	}
	if d.HasChange("settings.0.data_trace_enabled") {
		ops = append(ops, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String(prefix + "logging/dataTrace"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.data_trace_enabled").(bool))),
		})
	}
	if d.HasChange("settings.0.throttling_burst_limit") {
		ops = append(ops, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String(prefix + "throttling/burstLimit"),
			Value: aws.String(fmt.Sprintf("%d", d.Get("settings.0.throttling_burst_limit").(int))),
		})
	}
	if d.HasChange("settings.0.throttling_rate_limit") {
		ops = append(ops, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String(prefix + "throttling/rateLimit"),
			Value: aws.String(fmt.Sprintf("%f", d.Get("settings.0.throttling_rate_limit").(float64))),
		})
	}
	if d.HasChange("settings.0.caching_enabled") {
		ops = append(ops, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String(prefix + "caching/enabled"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.caching_enabled").(bool))),
		})
	}
	if v, ok := d.GetOkExists("settings.0.cache_ttl_in_seconds"); ok {
		ops = append(ops, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String(prefix + "caching/ttlInSeconds"),
			Value: aws.String(fmt.Sprintf("%d", v.(int))),
		})
	}
	if d.HasChange("settings.0.cache_data_encrypted") {
		ops = append(ops, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String(prefix + "caching/dataEncrypted"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.cache_data_encrypted").(bool))),
		})
	}
	if d.HasChange("settings.0.require_authorization_for_cache_control") {
		ops = append(ops, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String(prefix + "caching/requireAuthorizationForCacheControl"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.require_authorization_for_cache_control").(bool))),
		})
	}
	if d.HasChange("settings.0.unauthorized_cache_control_header_strategy") {
		ops = append(ops, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String(prefix + "caching/unauthorizedCacheControlHeaderStrategy"),
			Value: aws.String(d.Get("settings.0.unauthorized_cache_control_header_strategy").(string)),
		})
	}

	apiID := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	id := apiID + "-" + stageName + "-" + methodPath
	input := &apigateway.UpdateStageInput{
		PatchOperations: ops,
		RestApiId:       aws.String(apiID),
		StageName:       aws.String(stageName),
	}

	_, err := conn.UpdateStage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Stage (%s): %s", id, err)
	}

	if d.IsNewResource() {
		d.SetId(id)
	}

	return append(diags, resourceMethodSettingsRead(ctx, d, meta)...)
}

func resourceMethodSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.UpdateStageInput{
		PatchOperations: []types.PatchOperation{
			{
				Op:   types.OpRemove,
				Path: aws.String(fmt.Sprintf("/%s", d.Get("method_path").(string))),
			},
		},
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
		StageName: aws.String(d.Get("stage_name").(string)),
	}

	_, err := conn.UpdateStage(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	// BadRequestException: Cannot remove method setting */* because there is no method setting for this method
	if errs.IsAErrorMessageContains[*types.BadRequestException](err, "no method setting for this method") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Stage (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceMethodSettingsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 3)
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/STAGE-NAME/METHOD-PATH", d.Id())
	}
	restApiID := idParts[0]
	stageName := idParts[1]
	methodPath := idParts[2]
	d.Set("rest_api_id", restApiID)
	d.Set("stage_name", stageName)
	d.Set("method_path", methodPath)
	d.SetId(fmt.Sprintf("%s-%s-%s", restApiID, stageName, methodPath))
	return []*schema.ResourceData{d}, nil
}

func findMethodSettingsByThreePartKey(ctx context.Context, conn *apigateway.Client, apiID, stageName, methodPath string) (*types.MethodSetting, error) {
	stage, err := findStageByTwoPartKey(ctx, conn, apiID, stageName)

	if err != nil {
		return nil, err
	}

	output, ok := stage.MethodSettings[methodPath]

	if !ok {
		return nil, tfresource.NewEmptyResultError(methodPath)
	}

	return &output, nil
}
