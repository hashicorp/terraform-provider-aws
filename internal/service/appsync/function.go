// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	functionVersion2018_05_29 = "2018-05-29"
)

// @SDKResource("aws_appsync_function", name="Function")
func resourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFunctionCreate,
		ReadWithoutTimeout:   resourceFunctionRead,
		UpdateWithoutTimeout: resourceFunctionUpdate,
		DeleteWithoutTimeout: resourceFunctionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"code": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"runtime"},
				ValidateFunc: validation.StringLenBetween(1, 32768),
			},
			"data_source": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"function_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					functionVersion2018_05_29,
				}, true),
			},
			"max_batch_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[A-Za-z_][0-9A-Za-z_]*`), "must match [A-Za-z_][0-9A-Za-z_]*"),
			},
			"request_mapping_template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"response_mapping_template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"runtime": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				RequiredWith: []string{"code"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.RuntimeName](),
						},
						"runtime_version": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"sync_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"conflict_detection": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ConflictDetectionType](),
						},
						"conflict_handler": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ConflictHandlerType](),
						},
						"lambda_conflict_handler_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"lambda_conflict_handler_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceFunctionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID := d.Get("api_id").(string)
	name := d.Get(names.AttrName).(string)
	input := &appsync.CreateFunctionInput{
		ApiId:           aws.String(apiID),
		DataSourceName:  aws.String(d.Get("data_source").(string)),
		FunctionVersion: aws.String(d.Get("function_version").(string)),
		Name:            aws.String(name),
	}

	if v, ok := d.GetOk("code"); ok {
		input.Code = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_batch_size"); ok {
		input.MaxBatchSize = int32(v.(int))
	}

	if v, ok := d.GetOk("request_mapping_template"); ok {
		input.RequestMappingTemplate = aws.String(v.(string))
		input.FunctionVersion = aws.String(functionVersion2018_05_29)
	}

	if v, ok := d.GetOk("response_mapping_template"); ok {
		input.ResponseMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("runtime"); ok && len(v.([]interface{})) > 0 {
		input.Runtime = expandRuntime(v.([]interface{}))
	}

	if v, ok := d.GetOk("sync_config"); ok && len(v.([]interface{})) > 0 {
		input.SyncConfig = expandSyncConfig(v.([]interface{}))
	}

	output, err := conn.CreateFunction(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppSync Function: %s", err)
	}

	d.SetId(functionCreateResourceID(apiID, aws.ToString(output.FunctionConfiguration.FunctionId)))

	return append(diags, resourceFunctionRead(ctx, d, meta)...)
}

func resourceFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, functionID, err := functionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	function, err := findFunctionByTwoPartKey(ctx, conn, apiID, functionID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync Function (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync Function (%s): %s", d.Id(), err)
	}

	d.Set("api_id", apiID)
	d.Set(names.AttrARN, function.FunctionArn)
	d.Set("code", function.Code)
	d.Set("data_source", function.DataSourceName)
	d.Set(names.AttrDescription, function.Description)
	d.Set("function_id", functionID)
	d.Set("function_version", function.FunctionVersion)
	d.Set("max_batch_size", function.MaxBatchSize)
	d.Set(names.AttrName, function.Name)
	d.Set("request_mapping_template", function.RequestMappingTemplate)
	d.Set("response_mapping_template", function.ResponseMappingTemplate)
	if err := d.Set("runtime", flattenRuntime(function.Runtime)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting runtime: %s", err)
	}
	if err := d.Set("sync_config", flattenSyncConfig(function.SyncConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sync_config: %s", err)
	}

	return diags
}

func resourceFunctionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, functionID, err := functionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &appsync.UpdateFunctionInput{
		ApiId:           aws.String(apiID),
		DataSourceName:  aws.String(d.Get("data_source").(string)),
		FunctionId:      aws.String(functionID),
		FunctionVersion: aws.String(d.Get("function_version").(string)),
		Name:            aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("code"); ok {
		input.Code = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_batch_size"); ok {
		input.MaxBatchSize = int32(v.(int))
	}

	if v, ok := d.GetOk("request_mapping_template"); ok {
		input.RequestMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("response_mapping_template"); ok {
		input.ResponseMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("runtime"); ok && len(v.([]interface{})) > 0 {
		input.Runtime = expandRuntime(v.([]interface{}))
	}

	if v, ok := d.GetOk("sync_config"); ok && len(v.([]interface{})) > 0 {
		input.SyncConfig = expandSyncConfig(v.([]interface{}))
	}

	_, err = conn.UpdateFunction(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AppSync Function (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFunctionRead(ctx, d, meta)...)
}

func resourceFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, functionID, err := functionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Appsync Function: %s", d.Id())
	_, err = conn.DeleteFunction(ctx, &appsync.DeleteFunctionInput{
		ApiId:      aws.String(apiID),
		FunctionId: aws.String(functionID),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppSync Function (%s): %s", d.Id(), err)
	}

	return diags
}

const functionResourceIDSeparator = "-"

func functionCreateResourceID(apiID, functionID string) string {
	parts := []string{apiID, functionID}
	id := strings.Join(parts, functionResourceIDSeparator)

	return id
}

func functionParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, functionResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected API-ID%[2]sFUNCTION-ID", id, functionResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findFunctionByTwoPartKey(ctx context.Context, conn *appsync.Client, apiID, functionID string) (*awstypes.FunctionConfiguration, error) {
	input := &appsync.GetFunctionInput{
		ApiId:      aws.String(apiID),
		FunctionId: aws.String(functionID),
	}

	output, err := conn.GetFunction(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FunctionConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.FunctionConfiguration, nil
}

func expandRuntime(tfList []interface{}) *awstypes.AppSyncRuntime {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.AppSyncRuntime{}

	if v, ok := tfMap[names.AttrName].(string); ok {
		apiObject.Name = awstypes.RuntimeName(v)
	}

	if v, ok := tfMap["runtime_version"].(string); ok {
		apiObject.RuntimeVersion = aws.String(v)
	}

	return apiObject
}

func flattenRuntime(apiObject *awstypes.AppSyncRuntime) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrName:    apiObject.Name,
		"runtime_version": aws.ToString(apiObject.RuntimeVersion),
	}

	return []interface{}{tfMap}
}

func expandSyncConfig(tfList []interface{}) *awstypes.SyncConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.SyncConfig{}

	if v, ok := tfMap["conflict_detection"].(string); ok {
		apiObject.ConflictDetection = awstypes.ConflictDetectionType(v)
	}

	if v, ok := tfMap["conflict_handler"].(string); ok {
		apiObject.ConflictHandler = awstypes.ConflictHandlerType(v)
	}

	if v, ok := tfMap["lambda_conflict_handler_config"].([]interface{}); ok && len(v) > 0 {
		apiObject.LambdaConflictHandlerConfig = expandLambdaConflictHandlerConfig(v)
	}

	return apiObject
}

func flattenSyncConfig(apiObject *awstypes.SyncConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"conflict_detection":             apiObject.ConflictDetection,
		"conflict_handler":               apiObject.ConflictHandler,
		"lambda_conflict_handler_config": flattenLambdaConflictHandlerConfig(apiObject.LambdaConflictHandlerConfig),
	}

	return []interface{}{tfMap}
}

func expandLambdaConflictHandlerConfig(tfList []interface{}) *awstypes.LambdaConflictHandlerConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.LambdaConflictHandlerConfig{}

	if v, ok := tfMap["lambda_conflict_handler_arn"].(string); ok {
		apiObject.LambdaConflictHandlerArn = aws.String(v)
	}

	return apiObject
}

func flattenLambdaConflictHandlerConfig(apiObject *awstypes.LambdaConflictHandlerConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"lambda_conflict_handler_arn": aws.ToString(apiObject.LambdaConflictHandlerArn),
	}

	return []interface{}{tfMap}
}
