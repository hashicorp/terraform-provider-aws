package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceMethodSettings() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMethodSettingsUpdate,
		ReadWithoutTimeout:   resourceMethodSettingsRead,
		UpdateWithoutTimeout: resourceMethodSettingsUpdate,
		DeleteWithoutTimeout: resourceMethodSettingsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceMethodSettingsImport,
		},

		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stage_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"method_path": {
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
						"metrics_enabled": {
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
						"data_trace_enabled": {
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
						"caching_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"cache_ttl_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"cache_data_encrypted": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"require_authorization_for_cache_control": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"unauthorized_cache_control_header_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(apigateway.UnauthorizedCacheControlHeaderStrategy_Values(), false),
							Computed:     true,
						},
					},
				},
			},
		},
	}
}

func flattenMethodSettings(settings *apigateway.MethodSetting) []interface{} {
	if settings == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"metrics_enabled":                            settings.MetricsEnabled,
			"logging_level":                              settings.LoggingLevel,
			"data_trace_enabled":                         settings.DataTraceEnabled,
			"throttling_burst_limit":                     settings.ThrottlingBurstLimit,
			"throttling_rate_limit":                      settings.ThrottlingRateLimit,
			"caching_enabled":                            settings.CachingEnabled,
			"cache_ttl_in_seconds":                       settings.CacheTtlInSeconds,
			"cache_data_encrypted":                       settings.CacheDataEncrypted,
			"require_authorization_for_cache_control":    settings.RequireAuthorizationForCacheControl,
			"unauthorized_cache_control_header_strategy": settings.UnauthorizedCacheControlHeaderStrategy,
		},
	}
}

func resourceMethodSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	stage, err := FindStageByName(ctx, conn, d.Get("rest_api_id").(string), d.Get("stage_name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Stage Method Settings (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting API Gateway Stage Method Settings (%s): %s", d.Id(), err)
	}

	methodPath := d.Get("method_path").(string)
	settings, ok := stage.MethodSettings[methodPath]

	if !d.IsNewResource() && !ok {
		log.Printf("[WARN] API Gateway Stage Method Settings (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err := d.Set("settings", flattenMethodSettings(settings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting settings: %s", err)
	}

	return diags
}

func resourceMethodSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	methodPath := d.Get("method_path").(string)
	prefix := fmt.Sprintf("/%s/", methodPath)

	ops := make([]*apigateway.PatchOperation, 0)
	if d.HasChange("settings.0.metrics_enabled") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + "metrics/enabled"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.metrics_enabled").(bool))),
		})
	}
	if d.HasChange("settings.0.logging_level") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + "logging/loglevel"),
			Value: aws.String(d.Get("settings.0.logging_level").(string)),
		})
	}
	if d.HasChange("settings.0.data_trace_enabled") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + "logging/dataTrace"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.data_trace_enabled").(bool))),
		})
	}

	if d.HasChange("settings.0.throttling_burst_limit") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + "throttling/burstLimit"),
			Value: aws.String(fmt.Sprintf("%d", d.Get("settings.0.throttling_burst_limit").(int))),
		})
	}
	if d.HasChange("settings.0.throttling_rate_limit") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + "throttling/rateLimit"),
			Value: aws.String(fmt.Sprintf("%f", d.Get("settings.0.throttling_rate_limit").(float64))),
		})
	}
	if d.HasChange("settings.0.caching_enabled") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + "caching/enabled"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.caching_enabled").(bool))),
		})
	}

	if v, ok := d.GetOkExists("settings.0.cache_ttl_in_seconds"); ok {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + "caching/ttlInSeconds"),
			Value: aws.String(fmt.Sprintf("%d", v.(int))),
		})
	}

	if d.HasChange("settings.0.cache_data_encrypted") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + "caching/dataEncrypted"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.cache_data_encrypted").(bool))),
		})
	}
	if d.HasChange("settings.0.require_authorization_for_cache_control") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + "caching/requireAuthorizationForCacheControl"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.require_authorization_for_cache_control").(bool))),
		})
	}
	if d.HasChange("settings.0.unauthorized_cache_control_header_strategy") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + "caching/unauthorizedCacheControlHeaderStrategy"),
			Value: aws.String(d.Get("settings.0.unauthorized_cache_control_header_strategy").(string)),
		})
	}

	restApiId := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	input := apigateway.UpdateStageInput{
		RestApiId:       aws.String(restApiId),
		StageName:       aws.String(stageName),
		PatchOperations: ops,
	}
	log.Printf("[DEBUG] Updating API Gateway Stage: %s", input)

	_, err := conn.UpdateStageWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Stage failed: %s", err)
	}

	d.SetId(restApiId + "-" + stageName + "-" + methodPath)

	return append(diags, resourceMethodSettingsRead(ctx, d, meta)...)
}

func resourceMethodSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	input := &apigateway.UpdateStageInput{
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
		StageName: aws.String(d.Get("stage_name").(string)),
		PatchOperations: []*apigateway.PatchOperation{
			{
				Op:   aws.String(apigateway.OpRemove),
				Path: aws.String(fmt.Sprintf("/%s", d.Get("method_path").(string))),
			},
		},
	}

	_, err := conn.UpdateStageWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return diags
	}

	// BadRequestException: Cannot remove method setting */* because there is no method setting for this method
	if tfawserr.ErrMessageContains(err, apigateway.ErrCodeBadRequestException, "no method setting for this method") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Stage Method Settings (%s): %s", d.Id(), err)
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
