// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_stage", name="Stage")
// @Tags(identifierAttribute="arn")
func ResourceStage() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStageCreate,
		ReadWithoutTimeout:   resourceStageRead,
		UpdateWithoutTimeout: resourceStageUpdate,
		DeleteWithoutTimeout: resourceStageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/STAGE-NAME", d.Id())
				}
				restApiID := idParts[0]
				stageName := idParts[1]
				d.Set("stage_name", stageName)
				d.Set("rest_api_id", restApiID)
				d.SetId(fmt.Sprintf("ags-%s-%s", restApiID, stageName))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"access_log_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"format": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cache_cluster_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cache_cluster_size": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(apigateway.CacheClusterSize_Values(), true),
			},
			"canary_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"percent_traffic": {
							Type:     schema.TypeFloat,
							Optional: true,
							Default:  0.0,
						},
						"stage_variable_overrides": {
							Type:     schema.TypeMap,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"use_stage_cache": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"client_certificate_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"deployment_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"documentation_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invoke_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
			"variables": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"xray_tracing_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"web_acl_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	restAPIID := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	deploymentID := d.Get("deployment_id").(string)
	input := &apigateway.CreateStageInput{
		RestApiId:    aws.String(restAPIID),
		StageName:    aws.String(stageName),
		DeploymentId: aws.String(deploymentID),
		Tags:         getTagsIn(ctx),
	}

	waitForCache := false
	if v, ok := d.GetOk("cache_cluster_enabled"); ok {
		input.CacheClusterEnabled = aws.Bool(v.(bool))
		waitForCache = true
	}

	if v, ok := d.GetOk("cache_cluster_size"); ok {
		input.CacheClusterSize = aws.String(v.(string))
		waitForCache = true
	}

	if v, ok := d.GetOk("canary_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CanarySettings = expandCanarySettings(v.([]interface{})[0].(map[string]interface{}), deploymentID)
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("documentation_version"); ok {
		input.DocumentationVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("variables"); ok && len(v.(map[string]interface{})) > 0 {
		input.Variables = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("xray_tracing_enabled"); ok {
		input.TracingEnabled = aws.Bool(v.(bool))
	}

	output, err := conn.CreateStageWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Stage (%s): %s", stageName, err)
	}

	d.SetId(fmt.Sprintf("ags-%s-%s", restAPIID, stageName))

	if waitForCache && aws.StringValue(output.CacheClusterStatus) != apigateway.CacheClusterStatusNotAvailable {
		if _, err := waitStageCacheAvailable(ctx, conn, restAPIID, stageName); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for API Gateway Stage (%s) cache create: %s", d.Id(), err)
		}
	}

	_, certOk := d.GetOk("client_certificate_id")
	_, logsOk := d.GetOk("access_log_settings")

	if certOk || logsOk {
		return append(diags, resourceStageUpdate(ctx, d, meta)...)
	}

	return append(diags, resourceStageRead(ctx, d, meta)...)
}

func resourceStageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	restAPIID := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	stage, err := FindStageByTwoPartKey(ctx, conn, restAPIID, stageName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Stage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Stage (%s): %s", d.Id(), err)
	}

	if err := d.Set("access_log_settings", flattenAccessLogSettings(stage.AccessLogSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting access_log_settings: %s", err)
	}
	stageARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "apigateway",
		Resource:  fmt.Sprintf("/restapis/%s/stages/%s", restAPIID, stageName),
	}.String()
	d.Set("arn", stageARN)
	if aws.StringValue(stage.CacheClusterStatus) == apigateway.CacheClusterStatusDeleteInProgress {
		d.Set("cache_cluster_enabled", false)
		d.Set("cache_cluster_size", d.Get("cache_cluster_size"))
	} else {
		enabled := aws.BoolValue(stage.CacheClusterEnabled)
		d.Set("cache_cluster_enabled", enabled)
		if enabled {
			d.Set("cache_cluster_size", stage.CacheClusterSize)
		} else {
			d.Set("cache_cluster_size", d.Get("cache_cluster_size"))
		}
	}
	if err := d.Set("canary_settings", flattenCanarySettings(stage.CanarySettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting canary_settings: %s", err)
	}
	d.Set("client_certificate_id", stage.ClientCertificateId)
	d.Set("deployment_id", stage.DeploymentId)
	d.Set("description", stage.Description)
	d.Set("documentation_version", stage.DocumentationVersion)
	executionARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "execute-api",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("%s/%s", restAPIID, stageName),
	}.String()
	d.Set("execution_arn", executionARN)
	d.Set("invoke_url", meta.(*conns.AWSClient).APIGatewayInvokeURL(restAPIID, stageName))
	if err := d.Set("variables", aws.StringValueMap(stage.Variables)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting variables: %s", err)
	}
	d.Set("web_acl_arn", stage.WebAclArn)
	d.Set("xray_tracing_enabled", stage.TracingEnabled)

	setTagsOut(ctx, stage.Tags)

	return diags
}

func resourceStageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	restAPIID := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)

	if d.HasChangesExcept("tags", "tags_all") {
		operations := make([]*apigateway.PatchOperation, 0)
		waitForCache := false
		if d.HasChange("cache_cluster_enabled") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/cacheClusterEnabled"),
				Value: aws.String(fmt.Sprintf("%t", d.Get("cache_cluster_enabled").(bool))),
			})
			waitForCache = true
		}
		if d.HasChange("cache_cluster_size") && d.Get("cache_cluster_enabled").(bool) {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/cacheClusterSize"),
				Value: aws.String(d.Get("cache_cluster_size").(string)),
			})
			waitForCache = true
		}
		if d.HasChange("client_certificate_id") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/clientCertificateId"),
				Value: aws.String(d.Get("client_certificate_id").(string)),
			})
		}
		if d.HasChange("canary_settings") {
			oldCanarySettingsRaw, newCanarySettingsRaw := d.GetChange("canary_settings")
			operations = appendCanarySettingsPatchOperations(operations,
				oldCanarySettingsRaw.([]interface{}),
				newCanarySettingsRaw.([]interface{}),
			)
		}
		if d.HasChange("deployment_id") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/deploymentId"),
				Value: aws.String(d.Get("deployment_id").(string)),
			})

			if _, ok := d.GetOk("canary_settings"); ok {
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String(apigateway.OpReplace),
					Path:  aws.String("/canarySettings/deploymentId"),
					Value: aws.String(d.Get("deployment_id").(string)),
				})
			}
		}
		if d.HasChange("description") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/description"),
				Value: aws.String(d.Get("description").(string)),
			})
		}
		if d.HasChange("xray_tracing_enabled") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/tracingEnabled"),
				Value: aws.String(fmt.Sprintf("%t", d.Get("xray_tracing_enabled").(bool))),
			})
		}
		if d.HasChange("documentation_version") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/documentationVersion"),
				Value: aws.String(d.Get("documentation_version").(string)),
			})
		}
		if d.HasChange("variables") {
			o, n := d.GetChange("variables")
			oldV := o.(map[string]interface{})
			newV := n.(map[string]interface{})
			operations = append(operations, diffVariablesOps(oldV, newV, "/variables/")...)
		}
		if d.HasChange("access_log_settings") {
			accessLogSettings := d.Get("access_log_settings").([]interface{})
			if len(accessLogSettings) == 1 {
				operations = append(operations,
					&apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpReplace),
						Path:  aws.String("/accessLogSettings/destinationArn"),
						Value: aws.String(d.Get("access_log_settings.0.destination_arn").(string)),
					}, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpReplace),
						Path:  aws.String("/accessLogSettings/format"),
						Value: aws.String(d.Get("access_log_settings.0.format").(string)),
					})
			} else if len(accessLogSettings) == 0 {
				operations = append(operations, &apigateway.PatchOperation{
					Op:   aws.String(apigateway.OpRemove),
					Path: aws.String("/accessLogSettings"),
				})
			}
		}

		input := &apigateway.UpdateStageInput{
			RestApiId:       aws.String(restAPIID),
			StageName:       aws.String(stageName),
			PatchOperations: operations,
		}

		output, err := conn.UpdateStageWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway Stage (%s): %s", d.Id(), err)
		}

		if waitForCache && aws.StringValue(output.CacheClusterStatus) != apigateway.CacheClusterStatusNotAvailable {
			_, err := waitStageCacheUpdated(ctx, conn, restAPIID, stageName)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for API Gateway Stage (%s) to be updated: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceStageRead(ctx, d, meta)...)
}

func resourceStageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Stage: %s", d.Id())
	_, err := conn.DeleteStageWithContext(ctx, &apigateway.DeleteStageInput{
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
		StageName: aws.String(d.Get("stage_name").(string)),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Stage (%s): %s", d.Id(), err)
	}

	return diags
}

func FindStageByTwoPartKey(ctx context.Context, conn *apigateway.APIGateway, restAPIID, stageName string) (*apigateway.Stage, error) {
	input := &apigateway.GetStageInput{
		RestApiId: aws.String(restAPIID),
		StageName: aws.String(stageName),
	}

	output, err := conn.GetStageWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func diffVariablesOps(oldVars, newVars map[string]interface{}, prefix string) []*apigateway.PatchOperation {
	ops := make([]*apigateway.PatchOperation, 0)

	for k := range oldVars {
		if _, ok := newVars[k]; !ok {
			ops = append(ops, &apigateway.PatchOperation{
				Op:   aws.String(apigateway.OpRemove),
				Path: aws.String(prefix + k),
			})
		}
	}

	for k, v := range newVars {
		newValue := v.(string)

		if oldV, ok := oldVars[k]; ok {
			oldValue := oldV.(string)
			if oldValue == newValue {
				continue
			}
		}
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + k),
			Value: aws.String(newValue),
		})
	}

	return ops
}

func flattenAccessLogSettings(accessLogSettings *apigateway.AccessLogSettings) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)
	if accessLogSettings != nil {
		result = append(result, map[string]interface{}{
			"destination_arn": aws.StringValue(accessLogSettings.DestinationArn),
			"format":          aws.StringValue(accessLogSettings.Format),
		})
	}
	return result
}

func expandCanarySettings(tfMap map[string]interface{}, deploymentId string) *apigateway.CanarySettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &apigateway.CanarySettings{
		DeploymentId: aws.String(deploymentId),
	}

	if v, ok := tfMap["percent_traffic"].(float64); ok {
		apiObject.PercentTraffic = aws.Float64(v)
	}

	if v, ok := tfMap["stage_variable_overrides"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.StageVariableOverrides = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["use_stage_cache"].(bool); ok {
		apiObject.UseStageCache = aws.Bool(v)
	}

	return apiObject
}

func flattenCanarySettings(canarySettings *apigateway.CanarySettings) []interface{} {
	settings := make(map[string]interface{})

	if canarySettings == nil {
		return nil
	}

	overrides := aws.StringValueMap(canarySettings.StageVariableOverrides)

	if len(overrides) > 0 {
		settings["stage_variable_overrides"] = overrides
	}

	settings["percent_traffic"] = canarySettings.PercentTraffic
	settings["use_stage_cache"] = canarySettings.UseStageCache

	return []interface{}{settings}
}

func appendCanarySettingsPatchOperations(operations []*apigateway.PatchOperation, oldCanarySettingsRaw, newCanarySettingsRaw []interface{}) []*apigateway.PatchOperation {
	if len(newCanarySettingsRaw) == 0 { // Schema guarantees either 0 or 1
		return append(operations, &apigateway.PatchOperation{
			Op:   aws.String("remove"),
			Path: aws.String("/canarySettings"),
		})
	}
	newSettings := newCanarySettingsRaw[0].(map[string]interface{})

	var oldSettings map[string]interface{}
	if len(oldCanarySettingsRaw) == 1 { // Schema guarantees either 0 or 1
		oldSettings = oldCanarySettingsRaw[0].(map[string]interface{})
	} else {
		oldSettings = map[string]interface{}{
			"percent_traffic":          0.0,
			"stage_variable_overrides": make(map[string]interface{}),
			"use_stage_cache":          false,
		}
	}

	oldOverrides := oldSettings["stage_variable_overrides"].(map[string]interface{})
	newOverrides := newSettings["stage_variable_overrides"].(map[string]interface{})
	operations = append(operations, diffVariablesOps(oldOverrides, newOverrides, "/canarySettings/stageVariableOverrides/")...)

	oldPercentTraffic := oldSettings["percent_traffic"].(float64)
	newPercentTraffic := newSettings["percent_traffic"].(float64)
	if oldPercentTraffic != newPercentTraffic {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/canarySettings/percentTraffic"),
			Value: aws.String(fmt.Sprintf("%f", newPercentTraffic)),
		})
	}

	oldUseStageCache := oldSettings["use_stage_cache"].(bool)
	newUseStageCache := newSettings["use_stage_cache"].(bool)
	if oldUseStageCache != newUseStageCache {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/canarySettings/useStageCache"),
			Value: aws.String(fmt.Sprintf("%t", newUseStageCache)),
		})
	}

	return operations
}
