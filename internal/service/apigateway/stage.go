// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// @SDKResource("aws_api_gateway_stage", name="Stage")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigateway;apigateway.GetStageOutput", serialize=true)
// @Testing(importStateIdFunc=testAccStageImportStateIdFunc)
func resourceStage() *schema.Resource {
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
						names.AttrDestinationARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrFormat: {
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.CacheClusterSize](),
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
			names.AttrDescription: {
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
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	deploymentID := d.Get("deployment_id").(string)
	input := &apigateway.CreateStageInput{
		RestApiId:    aws.String(apiID),
		StageName:    aws.String(stageName),
		DeploymentId: aws.String(deploymentID),
		Tags:         getTagsIn(ctx),
	}

	waitForCache := false
	if v, ok := d.GetOk("cache_cluster_enabled"); ok {
		input.CacheClusterEnabled = v.(bool)
		waitForCache = true
	}

	if v, ok := d.GetOk("cache_cluster_size"); ok {
		input.CacheClusterSize = types.CacheClusterSize(v.(string))
		waitForCache = true
	}

	if v, ok := d.GetOk("canary_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CanarySettings = expandCanarySettings(v.([]interface{})[0].(map[string]interface{}), deploymentID)
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("documentation_version"); ok {
		input.DocumentationVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("variables"); ok && len(v.(map[string]interface{})) > 0 {
		input.Variables = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("xray_tracing_enabled"); ok {
		input.TracingEnabled = v.(bool)
	}

	output, err := conn.CreateStage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Stage (%s): %s", stageName, err)
	}

	d.SetId(fmt.Sprintf("ags-%s-%s", apiID, stageName))

	if waitForCache && output.CacheClusterStatus != types.CacheClusterStatusNotAvailable {
		if _, err := waitStageCacheAvailable(ctx, conn, apiID, stageName); err != nil {
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
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	stage, err := findStageByTwoPartKey(ctx, conn, apiID, stageName)

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
	d.Set(names.AttrARN, stageARN(meta.(*conns.AWSClient), apiID, stageName))
	if stage.CacheClusterStatus == types.CacheClusterStatusDeleteInProgress {
		d.Set("cache_cluster_enabled", false)
		d.Set("cache_cluster_size", d.Get("cache_cluster_size"))
	} else {
		enabled := stage.CacheClusterEnabled
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
	d.Set(names.AttrDescription, stage.Description)
	d.Set("documentation_version", stage.DocumentationVersion)
	d.Set("execution_arn", stageInvokeARN(meta.(*conns.AWSClient), apiID, stageName))
	d.Set("invoke_url", meta.(*conns.AWSClient).APIGatewayInvokeURL(ctx, apiID, stageName))
	if err := d.Set("variables", stage.Variables); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting variables: %s", err)
	}
	d.Set("web_acl_arn", stage.WebAclArn)
	d.Set("xray_tracing_enabled", stage.TracingEnabled)

	setTagsOut(ctx, stage.Tags)

	return diags
}

func resourceStageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		apiID := d.Get("rest_api_id").(string)
		stageName := d.Get("stage_name").(string)
		operations := make([]types.PatchOperation, 0)
		waitForCache := false

		if d.HasChange("cache_cluster_enabled") {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/cacheClusterEnabled"),
				Value: aws.String(fmt.Sprintf("%t", d.Get("cache_cluster_enabled").(bool))),
			})
			waitForCache = true
		}
		if d.HasChange("cache_cluster_size") && d.Get("cache_cluster_enabled").(bool) {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/cacheClusterSize"),
				Value: aws.String(d.Get("cache_cluster_size").(string)),
			})
			waitForCache = true
		}
		if d.HasChange("client_certificate_id") {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
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
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/deploymentId"),
				Value: aws.String(d.Get("deployment_id").(string)),
			})

			if _, ok := d.GetOk("canary_settings"); ok {
				operations = append(operations, types.PatchOperation{
					Op:    types.OpReplace,
					Path:  aws.String("/canarySettings/deploymentId"),
					Value: aws.String(d.Get("deployment_id").(string)),
				})
			}
		}
		if d.HasChange(names.AttrDescription) {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/description"),
				Value: aws.String(d.Get(names.AttrDescription).(string)),
			})
		}
		if d.HasChange("xray_tracing_enabled") {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/tracingEnabled"),
				Value: aws.String(fmt.Sprintf("%t", d.Get("xray_tracing_enabled").(bool))),
			})
		}
		if d.HasChange("documentation_version") {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
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
					types.PatchOperation{
						Op:    types.OpReplace,
						Path:  aws.String("/accessLogSettings/destinationArn"),
						Value: aws.String(d.Get("access_log_settings.0.destination_arn").(string)),
					}, types.PatchOperation{
						Op:    types.OpReplace,
						Path:  aws.String("/accessLogSettings/format"),
						Value: aws.String(d.Get("access_log_settings.0.format").(string)),
					})
			} else if len(accessLogSettings) == 0 {
				operations = append(operations, types.PatchOperation{
					Op:   types.OpRemove,
					Path: aws.String("/accessLogSettings"),
				})
			}
		}

		input := &apigateway.UpdateStageInput{
			RestApiId:       aws.String(apiID),
			StageName:       aws.String(stageName),
			PatchOperations: operations,
		}

		output, err := conn.UpdateStage(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway Stage (%s): %s", d.Id(), err)
		}

		if waitForCache && output.CacheClusterStatus != types.CacheClusterStatusNotAvailable {
			if _, err := waitStageCacheUpdated(ctx, conn, apiID, stageName); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for API Gateway Stage (%s) cache update: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceStageRead(ctx, d, meta)...)
}

func resourceStageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Stage: %s", d.Id())
	_, err := conn.DeleteStage(ctx, &apigateway.DeleteStageInput{
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
		StageName: aws.String(d.Get("stage_name").(string)),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Stage (%s): %s", d.Id(), err)
	}

	return diags
}

func findStageByTwoPartKey(ctx context.Context, conn *apigateway.Client, apiID, stageName string) (*apigateway.GetStageOutput, error) {
	input := &apigateway.GetStageInput{
		RestApiId: aws.String(apiID),
		StageName: aws.String(stageName),
	}

	output, err := conn.GetStage(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
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

func stageCacheStatus(ctx context.Context, conn *apigateway.Client, restApiId, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findStageByTwoPartKey(ctx, conn, restApiId, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, string(output.CacheClusterStatus), nil
	}
}

func waitStageCacheAvailable(ctx context.Context, conn *apigateway.Client, apiID, name string) (*types.Stage, error) {
	const (
		timeout = 90 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.CacheClusterStatusCreateInProgress, types.CacheClusterStatusDeleteInProgress, types.CacheClusterStatusFlushInProgress),
		Target:  enum.Slice(types.CacheClusterStatusAvailable),
		Refresh: stageCacheStatus(ctx, conn, apiID, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Stage); ok {
		return output, err
	}

	return nil, err
}

func waitStageCacheUpdated(ctx context.Context, conn *apigateway.Client, apiID, name string) (*types.Stage, error) {
	const (
		timeout = 30 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.CacheClusterStatusCreateInProgress, types.CacheClusterStatusFlushInProgress),
		Target: enum.Slice(
			types.CacheClusterStatusAvailable,
			// There's an AWS API bug (raised & confirmed in Sep 2016 by support)
			// which causes the stage to remain in deletion state forever
			// TODO: Check if this bug still exists in AWS SDK v2
			types.CacheClusterStatusDeleteInProgress,
		),
		Refresh: stageCacheStatus(ctx, conn, apiID, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Stage); ok {
		return output, err
	}

	return nil, err
}

func diffVariablesOps(oldVars, newVars map[string]interface{}, prefix string) []types.PatchOperation {
	ops := make([]types.PatchOperation, 0)

	for k := range oldVars {
		if _, ok := newVars[k]; !ok {
			ops = append(ops, types.PatchOperation{
				Op:   types.OpRemove,
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
		ops = append(ops, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String(prefix + k),
			Value: aws.String(newValue),
		})
	}

	return ops
}

func flattenAccessLogSettings(accessLogSettings *types.AccessLogSettings) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)
	if accessLogSettings != nil {
		result = append(result, map[string]interface{}{
			names.AttrDestinationARN: aws.ToString(accessLogSettings.DestinationArn),
			names.AttrFormat:         aws.ToString(accessLogSettings.Format),
		})
	}
	return result
}

func expandCanarySettings(tfMap map[string]interface{}, deploymentId string) *types.CanarySettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CanarySettings{
		DeploymentId: aws.String(deploymentId),
	}

	if v, ok := tfMap["percent_traffic"].(float64); ok {
		apiObject.PercentTraffic = v
	}

	if v, ok := tfMap["stage_variable_overrides"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.StageVariableOverrides = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["use_stage_cache"].(bool); ok {
		apiObject.UseStageCache = v
	}

	return apiObject
}

func flattenCanarySettings(canarySettings *types.CanarySettings) []interface{} {
	settings := make(map[string]interface{})

	if canarySettings == nil {
		return nil
	}

	overrides := canarySettings.StageVariableOverrides

	if len(overrides) > 0 {
		settings["stage_variable_overrides"] = overrides
	}

	settings["percent_traffic"] = canarySettings.PercentTraffic
	settings["use_stage_cache"] = canarySettings.UseStageCache

	return []interface{}{settings}
}

func appendCanarySettingsPatchOperations(operations []types.PatchOperation, oldCanarySettingsRaw, newCanarySettingsRaw []interface{}) []types.PatchOperation {
	if len(newCanarySettingsRaw) == 0 { // Schema guarantees either 0 or 1
		return append(operations, types.PatchOperation{
			Op:   types.Op("remove"),
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
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/canarySettings/percentTraffic"),
			Value: aws.String(fmt.Sprintf("%f", newPercentTraffic)),
		})
	}

	oldUseStageCache := oldSettings["use_stage_cache"].(bool)
	newUseStageCache := newSettings["use_stage_cache"].(bool)
	if oldUseStageCache != newUseStageCache {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/canarySettings/useStageCache"),
			Value: aws.String(fmt.Sprintf("%t", newUseStageCache)),
		})
	}

	return operations
}

func stageARN(c *conns.AWSClient, apiID, stageName string) string {
	return arn.ARN{
		Partition: c.Partition,
		Region:    c.Region,
		Service:   "apigateway",
		Resource:  fmt.Sprintf("/restapis/%s/stages/%s", apiID, stageName),
	}.String()
}

func stageInvokeARN(c *conns.AWSClient, apiID, stageName string) string {
	return arn.ARN{
		Partition: c.Partition,
		Service:   "execute-api",
		Region:    c.Region,
		AccountID: c.AccountID,
		Resource:  fmt.Sprintf("%s/%s", apiID, stageName),
	}.String()
}
