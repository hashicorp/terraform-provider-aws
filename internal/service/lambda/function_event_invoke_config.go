// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lambda_function_event_invoke_config", name="Function Event Invoke Config")
func resourceFunctionEventInvokeConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFunctionEventInvokeConfigCreate,
		ReadWithoutTimeout:   resourceFunctionEventInvokeConfigRead,
		UpdateWithoutTimeout: resourceFunctionEventInvokeConfigUpdate,
		DeleteWithoutTimeout: resourceFunctionEventInvokeConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"destination_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"on_failure": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDestination: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"on_success": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDestination: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
			},
			"function_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"maximum_event_age_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(60, 21600),
			},
			"maximum_retry_attempts": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      2,
				ValidateFunc: validation.IntBetween(0, 2),
			},
			"qualifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceFunctionEventInvokeConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)
	id := functionName
	if qualifier != "" {
		id = fmt.Sprintf("%s:%s", functionName, qualifier)
	}
	input := &lambda.PutFunctionEventInvokeConfigInput{
		DestinationConfig:    expandFunctionEventInvokeConfigDestinationConfig(d.Get("destination_config").([]interface{})),
		FunctionName:         aws.String(functionName),
		MaximumRetryAttempts: aws.Int32(int32(d.Get("maximum_retry_attempts").(int))),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	if v, ok := d.GetOk("maximum_event_age_in_seconds"); ok {
		input.MaximumEventAgeInSeconds = aws.Int32(int32(v.(int)))
	}

	// Retry for destination validation eventual consistency errors.
	_, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
		func() (interface{}, error) {
			return conn.PutFunctionEventInvokeConfig(ctx, input)
		},
		func(err error) (bool, error) {
			// InvalidParameterValueException: The destination ARN arn:PARTITION:SERVICE:REGION:ACCOUNT:RESOURCE is invalid.
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "destination ARN") {
				return true, err
			}

			// InvalidParameterValueException: The function's execution role does not have permissions to call Publish on arn:...
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "does not have permissions") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lambda Function Event Invoke Config (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceFunctionEventInvokeConfigRead(ctx, d, meta)...)
}

func resourceFunctionEventInvokeConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName, qualifier, err := functionEventInvokeConfigParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findFunctionEventInvokeConfigByTwoPartKey(ctx, conn, functionName, qualifier)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Function Event Invoke Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Function Event Invoke Config (%s): %s", d.Id(), err)
	}

	if err := d.Set("destination_config", flattenFunctionEventInvokeConfigDestinationConfig(output.DestinationConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting destination_config: %s", err)
	}
	d.Set("function_name", functionName)
	d.Set("maximum_event_age_in_seconds", output.MaximumEventAgeInSeconds)
	d.Set("maximum_retry_attempts", output.MaximumRetryAttempts)
	d.Set("qualifier", qualifier)

	return diags
}

func resourceFunctionEventInvokeConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName, qualifier, err := functionEventInvokeConfigParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &lambda.PutFunctionEventInvokeConfigInput{
		DestinationConfig:    expandFunctionEventInvokeConfigDestinationConfig(d.Get("destination_config").([]interface{})),
		FunctionName:         aws.String(functionName),
		MaximumRetryAttempts: aws.Int32(int32(d.Get("maximum_retry_attempts").(int))),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	if v, ok := d.GetOk("maximum_event_age_in_seconds"); ok {
		input.MaximumEventAgeInSeconds = aws.Int32(int32(v.(int)))
	}

	// Retry for destination validation eventual consistency errors.
	_, err = tfresource.RetryWhen(ctx, iamPropagationTimeout,
		func() (interface{}, error) {
			return conn.PutFunctionEventInvokeConfig(ctx, input)
		},
		func(err error) (bool, error) {
			// InvalidParameterValueException: The destination ARN arn:PARTITION:SERVICE:REGION:ACCOUNT:RESOURCE is invalid.
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "destination ARN") {
				return true, err
			}

			// InvalidParameterValueException: The function's execution role does not have permissions to call Publish on arn:...
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "does not have permissions") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lambda Function Event Invoke Config (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFunctionEventInvokeConfigRead(ctx, d, meta)...)
}

func resourceFunctionEventInvokeConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName, qualifier, err := functionEventInvokeConfigParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &lambda.DeleteFunctionEventInvokeConfigInput{
		FunctionName: aws.String(functionName),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	log.Printf("[INFO] Deleting Lambda Function Event Invoke Config: %s", d.Id())
	_, err = conn.DeleteFunctionEventInvokeConfig(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Function Event Invoke Config (%s): %s", d.Id(), err)
	}

	return diags
}

func functionEventInvokeConfigParseResourceID(id string) (string, string, error) {
	if arn.IsARN(id) {
		parsedARN, err := arn.Parse(id)

		if err != nil {
			return "", "", fmt.Errorf("parsing ARN (%s): %s", id, err)
		}

		function := strings.TrimPrefix(parsedARN.Resource, "function:")

		if !strings.Contains(function, ":") {
			// Return ARN for function name to match configuration
			return id, "", nil
		}

		functionParts := strings.Split(function, ":")

		if len(functionParts) != 2 || functionParts[0] == "" || functionParts[1] == "" {
			return "", "", fmt.Errorf("unexpected format of function resource (%s), expected name:qualifier", id)
		}

		qualifier := functionParts[1]
		// Return ARN minus qualifier for function name to match configuration
		functionName := strings.TrimSuffix(id, fmt.Sprintf(":%s", qualifier))

		return functionName, qualifier, nil
	}

	if !strings.Contains(id, ":") {
		return id, "", nil
	}

	idParts := strings.Split(id, ":")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected name or name:qualifier", id)
	}

	return idParts[0], idParts[1], nil
}

func findFunctionEventInvokeConfig(ctx context.Context, conn *lambda.Client, input *lambda.GetFunctionEventInvokeConfigInput) (*lambda.GetFunctionEventInvokeConfigOutput, error) {
	output, err := conn.GetFunctionEventInvokeConfig(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findFunctionEventInvokeConfigByTwoPartKey(ctx context.Context, conn *lambda.Client, functionName, qualifier string) (*lambda.GetFunctionEventInvokeConfigOutput, error) {
	input := &lambda.GetFunctionEventInvokeConfigInput{
		FunctionName: aws.String(functionName),
	}
	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	return findFunctionEventInvokeConfig(ctx, conn, input)
}

func expandFunctionEventInvokeConfigDestinationConfig(tfList []interface{}) *awstypes.DestinationConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	destinationConfig := &awstypes.DestinationConfig{}

	if v, ok := tfMap["on_failure"].([]interface{}); ok {
		destinationConfig.OnFailure = expandFunctionEventInvokeConfigDestinationConfigOnFailure(v)
	}

	if v, ok := tfMap["on_success"].([]interface{}); ok {
		destinationConfig.OnSuccess = expandFunctionEventInvokeConfigDestinationConfigOnSuccess(v)
	}

	return destinationConfig
}

func expandFunctionEventInvokeConfigDestinationConfigOnFailure(tfList []interface{}) *awstypes.OnFailure {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	onFailure := &awstypes.OnFailure{}

	if v, ok := tfMap[names.AttrDestination].(string); ok {
		onFailure.Destination = aws.String(v)
	}

	return onFailure
}

func expandFunctionEventInvokeConfigDestinationConfigOnSuccess(tfList []interface{}) *awstypes.OnSuccess {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	onSuccess := &awstypes.OnSuccess{}

	if v, ok := tfMap[names.AttrDestination].(string); ok {
		onSuccess.Destination = aws.String(v)
	}

	return onSuccess
}

func flattenFunctionEventInvokeConfigDestinationConfig(apiObject *awstypes.DestinationConfig) []interface{} {
	// The API will respond with empty OnFailure and OnSuccess destinations when unconfigured:
	// "DestinationConfig":{"OnFailure":{"Destination":null},"OnSuccess":{"Destination":null}}
	// Return no destination configuration to prevent Terraform state difference

	if apiObject == nil {
		return []interface{}{}
	}

	if apiObject.OnFailure == nil && apiObject.OnSuccess == nil {
		return []interface{}{}
	}

	if (apiObject.OnFailure != nil && apiObject.OnFailure.Destination == nil) && (apiObject.OnSuccess != nil && apiObject.OnSuccess.Destination == nil) {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"on_failure": flattenFunctionEventInvokeConfigDestinationConfigOnFailure(apiObject.OnFailure),
		"on_success": flattenFunctionEventInvokeConfigDestinationConfigOnSuccess(apiObject.OnSuccess),
	}

	return []interface{}{tfMap}
}

func flattenFunctionEventInvokeConfigDestinationConfigOnFailure(apiObject *awstypes.OnFailure) []interface{} {
	// The API will respond with empty OnFailure destination when unconfigured:
	// "DestinationConfig":{"OnFailure":{"Destination":null},"OnSuccess":{"Destination":null}}
	// Return no on failure configuration to prevent Terraform state difference

	if apiObject == nil || apiObject.Destination == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrDestination: aws.ToString(apiObject.Destination),
	}

	return []interface{}{tfMap}
}

func flattenFunctionEventInvokeConfigDestinationConfigOnSuccess(apiObject *awstypes.OnSuccess) []interface{} {
	// The API will respond with empty OnSuccess destination when unconfigured:
	// "DestinationConfig":{"OnFailure":{"Destination":null},"OnSuccess":{"Destination":null}}
	// Return no on success configuration to prevent Terraform state difference

	if apiObject == nil || apiObject.Destination == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrDestination: aws.ToString(apiObject.Destination),
	}

	return []interface{}{m}
}
