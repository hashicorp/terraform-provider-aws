// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lambda_invocation", name="Invocation")
func resourceInvocation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInvocationCreate,
		ReadWithoutTimeout:   schema.NoopContext,
		UpdateWithoutTimeout: resourceInvocationUpdate,
		DeleteWithoutTimeout: resourceInvocationDelete,

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
			"lifecycle_scope": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          lifecycleScopeCreateOnly,
				ValidateDiagFunc: enum.Validate[lifecycleScope](),
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
			"terraform_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "tf",
			},
			names.AttrTriggers: {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			customizeDiffValidateInput,
			customizeDiffInputChangeWithCreateOnlyScope,
		),
	}
}

func resourceInvocationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	return append(diags, invoke(ctx, conn, d, invocationActionCreate)...)
}

func resourceInvocationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	return append(diags, invoke(ctx, conn, d, invocationActionUpdate)...)
}

func resourceInvocationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	if !isCreateOnlyScope(d) {
		return append(diags, invoke(ctx, conn, d, invocationActionDelete)...)
	}

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
func buildInput(d *schema.ResourceData, action invocationAction) ([]byte, error) {
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
		names.AttrAction: action,
		"prev_input":     oldInputMap,
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
	return lifecycleScope(d.Get("lifecycle_scope").(string)) == lifecycleScopeCreateOnly
}

func invoke(ctx context.Context, conn *lambda.Client, d *schema.ResourceData, action invocationAction) diag.Diagnostics {
	var diags diag.Diagnostics

	functionName := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)
	payload, err := buildInput(d, action)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Lambda Invocation (%s) input transformation failed for input (%s): %s", d.Id(), d.Get("input").(string), err)
	}

	input := &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: awstypes.InvocationTypeRequestResponse,
		Payload:        payload,
		Qualifier:      aws.String(qualifier),
	}

	output, err := conn.Invoke(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "invoking Lambda Function (%s): %s", functionName, err)
	}

	if output.FunctionError != nil {
		return sdkdiag.AppendErrorf(diags, "invoking Lambda Function (%s): %s", functionName, string(output.Payload))
	}

	d.SetId(fmt.Sprintf("%s_%s_%x", functionName, qualifier, md5.Sum(payload)))
	d.Set("result", string(output.Payload))

	return diags
}

// customizeDiffValidateInput validates that `input` is JSON object when
// `lifecycle_scope` is not "CREATE_ONLY"
func customizeDiffValidateInput(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if lifecycleScope(diff.Get("lifecycle_scope").(string)) == lifecycleScopeCreateOnly {
		return nil
	}

	if !diff.GetRawPlan().GetAttr("input").IsWhollyKnown() {
		return nil
	}

	// input is validated to be valid JSON in the schema already.
	inputNoSpaces := strings.TrimSpace(diff.Get("input").(string))
	if strings.HasPrefix(inputNoSpaces, "{") && strings.HasSuffix(inputNoSpaces, "}") {
		return nil
	}

	return errors.New(`lifecycle_scope other than "CREATE_ONLY" requires input to be a JSON object`)
}

// customizeDiffInputChangeWithCreateOnlyScope forces a new resource when `input` has
// a change and `lifecycle_scope` is set to "CREATE_ONLY"
func customizeDiffInputChangeWithCreateOnlyScope(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.HasChange("input") && lifecycleScope(diff.Get("lifecycle_scope").(string)) == lifecycleScopeCreateOnly {
		return diff.ForceNew("input")
	}
	return nil
}
