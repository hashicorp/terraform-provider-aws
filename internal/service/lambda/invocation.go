// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

const defaultInvocationTerraformKey = "tf"

// @SDKResource("aws_lambda_invocation")
func ResourceInvocation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInvocationCreate,
		ReadWithoutTimeout:   resourceInvocationRead,
		DeleteWithoutTimeout: resourceInvocationDelete,
		UpdateWithoutTimeout: resourceInvocationUpdate,

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"input": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  FunctionVersionLatest,
			},
			"result": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"lifecycle_scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      lifecycleScopeCreateOnly,
				ValidateFunc: validation.StringInSlice(lifecycleScope_Values(), false),
			},
			"terraform_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  defaultInvocationTerraformKey,
			},
		},
		CustomizeDiff: customdiff.Sequence(
			customizeDiffValidateInput,
			customizeDiffInputChangeWithCreateOnlyScope,
		),
	}
}

func resourceInvocationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return invoke(ctx, invocationActionCreate, d, meta)
}

func resourceInvocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceInvocationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return invoke(ctx, invocationActionUpdate, d, meta)
}

func resourceInvocationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if !isCreateOnlyScope(d) {
		log.Printf("[DEBUG] Lambda Invocation (%s) \"deleted\" by invocation & removing from state", d.Id())
		return invoke(ctx, invocationActionDelete, d, meta)
	}
	var diags diag.Diagnostics
	log.Printf("[DEBUG] Lambda Invocation (%s) \"deleted\" by removing from state", d.Id())
	return diags
}

// buildInput makes sure that the user provided input is enriched for handling lifecycle events
//
// In order to make this a non-breaking change this function only manipulates input if
// the invocation is not only for creation of resources. In order for the lambda
// to understand the action it has to take we pass on the action that terraform wants to do
// on the invocation resource.
//
// Because Lambda functions by default are stateless we must pass the input from the previous
// invocation to allow implementation of delete/update at Lambda side.
func buildInput(d *schema.ResourceData, action string) ([]byte, error) {
	if isCreateOnlyScope(d) {
		jsonBytes := []byte(d.Get("input").(string))
		return jsonBytes, nil
	}
	oldInputMap, newInputMap, err := getInputChange(d)
	if err != nil {
		log.Printf("[DEBUG] input serialization %s", err)
		return nil, err
	}

	newInputMap[d.Get("terraform_key").(string)] = map[string]interface{}{
		"action":     action,
		"prev_input": oldInputMap,
	}
	return json.Marshal(&newInputMap)
}

func getObjectFromJSONString(s string) (map[string]interface{}, error) {
	if len(s) == 0 {
		return nil, nil
	}
	var mapObject map[string]interface{}
	if err := json.Unmarshal([]byte(s), &mapObject); err != nil {
		log.Printf("[ERROR] input JSON deserialization '%s'", s)
		return nil, err
	}
	return mapObject, nil
}

// getInputChange gets old an new input as maps
func getInputChange(d *schema.ResourceData) (map[string]interface{}, map[string]interface{}, error) {
	old, new := d.GetChange("input")
	oldMap, err := getObjectFromJSONString(old.(string))
	if err != nil {
		log.Printf("[ERROR] old input serialization '%s'", old.(string))
		return nil, nil, err
	}
	newMap, err := getObjectFromJSONString(new.(string))
	if err != nil {
		log.Printf("[ERROR] new input serialization '%s'", new.(string))
		return nil, nil, err
	}
	return oldMap, newMap, nil
}

// isCreateOnlyScope returns True if Lambda is only invoked when the resource is
// created or replaced.
//
// The original invocation logic only triggers on create.
func isCreateOnlyScope(d *schema.ResourceData) bool {
	return d.Get("lifecycle_scope").(string) == lifecycleScopeCreateOnly
}

func invoke(ctx context.Context, action string, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	functionName := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)
	input, err := buildInput(d, action)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Lambda Invocation (%s) input transformation failed for input (%s): %s", d.Id(), d.Get("input").(string), err)
	}

	res, err := conn.InvokeWithContext(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
		Payload:        input,
		Qualifier:      aws.String(qualifier),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Lambda Invocation (%s) failed: %s", d.Id(), err)
	}

	if res.FunctionError != nil {
		return sdkdiag.AppendErrorf(diags, "Lambda function (%s) returned error: (%s)", functionName, string(res.Payload))
	}

	d.SetId(fmt.Sprintf("%s_%s_%x", functionName, qualifier, md5.Sum(input)))
	d.Set("result", string(res.Payload))

	return diags
}
