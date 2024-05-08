// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lambda_alias", name="Alias")
func resourceAlias() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAliasCreate,
		ReadWithoutTimeout:   resourceAliasRead,
		UpdateWithoutTimeout: resourceAliasUpdate,
		DeleteWithoutTimeout: resourceAliasDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAliasImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"function_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressEquivalentFunctionNameOrARN,
			},
			"function_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"invoke_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"routing_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_version_weights": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeFloat},
						},
					},
				},
			},
		},
	}
}

func resourceAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &lambda.CreateAliasInput{
		Description:     aws.String(d.Get(names.AttrDescription).(string)),
		FunctionName:    aws.String(d.Get("function_name").(string)),
		FunctionVersion: aws.String(d.Get("function_version").(string)),
		Name:            aws.String(name),
		RoutingConfig:   expandAliasRoutingConfiguration(d.Get("routing_config").([]interface{})),
	}

	output, err := conn.CreateAlias(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lambda Alias (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.AliasArn))

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

func resourceAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	output, err := findAliasByTwoPartKey(ctx, conn, d.Get("function_name").(string), d.Get(names.AttrName).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Alias %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Alias (%s): %s", d.Id(), err)
	}

	aliasARN := aws.ToString(output.AliasArn)
	d.SetId(aliasARN) // For import.
	d.Set(names.AttrARN, aliasARN)
	d.Set(names.AttrDescription, output.Description)
	d.Set("function_version", output.FunctionVersion)
	d.Set("invoke_arn", invokeARN(meta.(*conns.AWSClient), aliasARN))
	d.Set(names.AttrName, output.Name)
	if err := d.Set("routing_config", flattenAliasRoutingConfiguration(output.RoutingConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting routing_config: %s", err)
	}

	return diags
}

func resourceAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	input := &lambda.UpdateAliasInput{
		Description:     aws.String(d.Get(names.AttrDescription).(string)),
		FunctionName:    aws.String(d.Get("function_name").(string)),
		FunctionVersion: aws.String(d.Get("function_version").(string)),
		Name:            aws.String(d.Get(names.AttrName).(string)),
		RoutingConfig:   expandAliasRoutingConfiguration(d.Get("routing_config").([]interface{})),
	}

	_, err := conn.UpdateAlias(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lambda Alias (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

func resourceAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	log.Printf("[INFO] Deleting Lambda Alias: %s", d.Id())
	_, err := conn.DeleteAlias(ctx, &lambda.DeleteAliasInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
		Name:         aws.String(d.Get(names.AttrName).(string)),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Alias (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAliasImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("Unexpected format of ID (%q), expected FUNCTION_NAME/ALIAS", d.Id())
	}

	functionName := idParts[0]
	alias := idParts[1]

	d.Set("function_name", functionName)
	d.Set(names.AttrName, alias)
	return []*schema.ResourceData{d}, nil
}

func findAliasByTwoPartKey(ctx context.Context, conn *lambda.Client, functionName, aliasName string) (*lambda.GetAliasOutput, error) {
	input := &lambda.GetAliasInput{
		FunctionName: aws.String(functionName),
		Name:         aws.String(aliasName),
	}

	return findAlias(ctx, conn, input)
}

func findAlias(ctx context.Context, conn *lambda.Client, input *lambda.GetAliasInput) (*lambda.GetAliasOutput, error) {
	output, err := conn.GetAlias(ctx, input)

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

func expandAliasRoutingConfiguration(tfList []interface{}) *awstypes.AliasRoutingConfiguration {
	apiObject := &awstypes.AliasRoutingConfiguration{}

	if len(tfList) == 0 || tfList[0] == nil {
		return apiObject
	}

	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["additional_version_weights"]; ok {
		apiObject.AdditionalVersionWeights = flex.ExpandFloat64ValueMap(v.(map[string]interface{}))
	}

	return apiObject
}

func flattenAliasRoutingConfiguration(apiObject *awstypes.AliasRoutingConfiguration) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"additional_version_weights": apiObject.AdditionalVersionWeights,
	}

	return []interface{}{tfMap}
}

func suppressEquivalentFunctionNameOrARN(k, old, new string, d *schema.ResourceData) bool {
	// Using function name or ARN should not be shown as a diff.
	// Try to convert the old and new values from ARN to function name
	oldFunctionName, oldFunctionNameErr := getFunctionNameFromARN(old)
	newFunctionName, newFunctionNameErr := getFunctionNameFromARN(new)
	return (oldFunctionName == new && oldFunctionNameErr == nil) || (newFunctionName == old && newFunctionNameErr == nil)
}
