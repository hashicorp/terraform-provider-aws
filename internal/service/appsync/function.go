package appsync

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFunction() *schema.Resource {
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
			"arn": {
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
			"description": {
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
					"2018-05-29",
				}, true),
			},
			"max_batch_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[_A-Za-z][_0-9A-Za-z]*`), "must match [_A-Za-z][_0-9A-Za-z]*"),
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
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appsync.RuntimeName_Values(), false),
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(appsync.ConflictDetectionType_Values(), false),
						},
						"conflict_handler": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(appsync.ConflictHandlerType_Values(), false),
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
	conn := meta.(*conns.AWSClient).AppSyncConn()

	apiID := d.Get("api_id").(string)

	input := &appsync.CreateFunctionInput{
		ApiId:           aws.String(apiID),
		DataSourceName:  aws.String(d.Get("data_source").(string)),
		FunctionVersion: aws.String(d.Get("function_version").(string)),
		Name:            aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("code"); ok {
		input.Code = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("request_mapping_template"); ok {
		input.RequestMappingTemplate = aws.String(v.(string))
		input.FunctionVersion = aws.String("2018-05-29")
	}

	if v, ok := d.GetOk("response_mapping_template"); ok {
		input.ResponseMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOkExists("max_batch_size"); ok {
		input.MaxBatchSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("sync_config"); ok && len(v.([]interface{})) > 0 {
		input.SyncConfig = expandSyncConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("runtime"); ok && len(v.([]interface{})) > 0 {
		input.Runtime = expandRuntime(v.([]interface{}))
	}

	resp, err := conn.CreateFunctionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppSync Function: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s", apiID, aws.StringValue(resp.FunctionConfiguration.FunctionId)))

	return append(diags, resourceFunctionRead(ctx, d, meta)...)
}

func resourceFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn()

	apiID, functionID, err := DecodeFunctionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppSync Function (%s): %s", d.Id(), err)
	}

	input := &appsync.GetFunctionInput{
		ApiId:      aws.String(apiID),
		FunctionId: aws.String(functionID),
	}

	resp, err := conn.GetFunctionWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] AppSync Function (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppSync Function (%s): %s", d.Id(), err)
	}

	function := resp.FunctionConfiguration
	d.Set("api_id", apiID)
	d.Set("function_id", functionID)
	d.Set("data_source", function.DataSourceName)
	d.Set("description", function.Description)
	d.Set("arn", function.FunctionArn)
	d.Set("function_version", function.FunctionVersion)
	d.Set("name", function.Name)
	d.Set("request_mapping_template", function.RequestMappingTemplate)
	d.Set("response_mapping_template", function.ResponseMappingTemplate)
	d.Set("max_batch_size", function.MaxBatchSize)
	d.Set("code", function.Code)

	if err := d.Set("sync_config", flattenSyncConfig(function.SyncConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sync_config: %s", err)
	}

	if err := d.Set("runtime", flattenRuntime(function.Runtime)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting runtime: %s", err)
	}

	return diags
}

func resourceFunctionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn()

	apiID, functionID, err := DecodeFunctionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AppSync Function (%s): %s", d.Id(), err)
	}

	input := &appsync.UpdateFunctionInput{
		ApiId:           aws.String(apiID),
		DataSourceName:  aws.String(d.Get("data_source").(string)),
		FunctionId:      aws.String(functionID),
		FunctionVersion: aws.String(d.Get("function_version").(string)),
		Name:            aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("code"); ok {
		input.Code = aws.String(v.(string))
	}

	if v, ok := d.GetOk("request_mapping_template"); ok {
		input.RequestMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("response_mapping_template"); ok {
		input.ResponseMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_batch_size"); ok {
		input.MaxBatchSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("sync_config"); ok && len(v.([]interface{})) > 0 {
		input.SyncConfig = expandSyncConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("runtime"); ok && len(v.([]interface{})) > 0 {
		input.Runtime = expandRuntime(v.([]interface{}))
	}

	_, err = conn.UpdateFunctionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AppSync Function (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFunctionRead(ctx, d, meta)...)
}

func resourceFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn()

	apiID, functionID, err := DecodeFunctionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppSync Function (%s): %s", d.Id(), err)
	}

	input := &appsync.DeleteFunctionInput{
		ApiId:      aws.String(apiID),
		FunctionId: aws.String(functionID),
	}

	_, err = conn.DeleteFunctionWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppSync Function (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeFunctionID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "-", 2)
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format ApiID-FunctionID, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}

func expandRuntime(l []interface{}) *appsync.AppSyncRuntime {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.AppSyncRuntime{}

	if v, ok := configured["name"].(string); ok {
		result.Name = aws.String(v)
	}

	if v, ok := configured["runtime_version"].(string); ok {
		result.RuntimeVersion = aws.String(v)
	}

	return result
}

func flattenRuntime(config *appsync.AppSyncRuntime) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"name":            aws.StringValue(config.Name),
		"runtime_version": aws.StringValue(config.RuntimeVersion),
	}

	return []map[string]interface{}{result}
}

func expandSyncConfig(l []interface{}) *appsync.SyncConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.SyncConfig{}

	if v, ok := configured["conflict_detection"].(string); ok {
		result.ConflictDetection = aws.String(v)
	}

	if v, ok := configured["conflict_handler"].(string); ok {
		result.ConflictHandler = aws.String(v)
	}

	if v, ok := configured["lambda_conflict_handler_config"].([]interface{}); ok && len(v) > 0 {
		result.LambdaConflictHandlerConfig = expandLambdaConflictHandlerConfig(v)
	}

	return result
}

func flattenSyncConfig(config *appsync.SyncConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"conflict_detection":             aws.StringValue(config.ConflictDetection),
		"conflict_handler":               aws.StringValue(config.ConflictHandler),
		"lambda_conflict_handler_config": flattenLambdaConflictHandlerConfig(config.LambdaConflictHandlerConfig),
	}

	return []map[string]interface{}{result}
}

func expandLambdaConflictHandlerConfig(l []interface{}) *appsync.LambdaConflictHandlerConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.LambdaConflictHandlerConfig{}

	if v, ok := configured["lambda_conflict_handler_arn"].(string); ok {
		result.LambdaConflictHandlerArn = aws.String(v)
	}

	return result
}

func flattenLambdaConflictHandlerConfig(config *appsync.LambdaConflictHandlerConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"lambda_conflict_handler_arn": aws.StringValue(config.LambdaConflictHandlerArn),
	}

	return []map[string]interface{}{result}
}
