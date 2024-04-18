// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_lambda_alias")
func ResourceAlias() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAliasCreate,
		ReadWithoutTimeout:   resourceAliasRead,
		UpdateWithoutTimeout: resourceAliasUpdate,
		DeleteWithoutTimeout: resourceAliasDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAliasImport,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Using function name or ARN should not be shown as a diff.
					// Try to convert the old and new values from ARN to function name
					oldFunctionName, oldFunctionNameErr := GetFunctionNameFromARN(old)
					newFunctionName, newFunctionNameErr := GetFunctionNameFromARN(new)
					return (oldFunctionName == new && oldFunctionNameErr == nil) || (newFunctionName == old && newFunctionNameErr == nil)
				},
			},
			"function_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invoke_arn": {
				Type:     schema.TypeString,
				Computed: true,
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

// resourceAliasCreate maps to:
// CreateAlias in the API / SDK
func resourceAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	functionName := d.Get("function_name").(string)
	aliasName := d.Get("name").(string)

	log.Printf("[DEBUG] Creating Lambda alias: alias %s for function %s", aliasName, functionName)

	params := &lambda.CreateAliasInput{
		Description:     aws.String(d.Get("description").(string)),
		FunctionName:    aws.String(functionName),
		FunctionVersion: aws.String(d.Get("function_version").(string)),
		Name:            aws.String(aliasName),
		RoutingConfig:   expandAliasRoutingConfiguration(d.Get("routing_config").([]interface{})),
	}

	aliasConfiguration, err := conn.CreateAliasWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lambda alias: %s", err)
	}

	d.SetId(aws.StringValue(aliasConfiguration.AliasArn))

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

// resourceAliasRead maps to:
// GetAlias in the API / SDK
func resourceAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	log.Printf("[DEBUG] Fetching Lambda alias: %s:%s", d.Get("function_name"), d.Get("name"))

	params := &lambda.GetAliasInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
		Name:         aws.String(d.Get("name").(string)),
	}

	aliasConfiguration, err := conn.GetAliasWithContext(ctx, params)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Lambda Alias (%s): %s", d.Id(), err)
	}

	d.Set("description", aliasConfiguration.Description)
	d.Set("function_version", aliasConfiguration.FunctionVersion)
	d.Set("name", aliasConfiguration.Name)
	d.Set("arn", aliasConfiguration.AliasArn)
	d.SetId(aws.StringValue(aliasConfiguration.AliasArn))

	invokeArn := functionInvokeARN(*aliasConfiguration.AliasArn, meta)
	d.Set("invoke_arn", invokeArn)

	if err := d.Set("routing_config", flattenAliasRoutingConfiguration(aliasConfiguration.RoutingConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting routing_config: %s", err)
	}

	return diags
}

// resourceAliasDelete maps to:
// DeleteAlias in the API / SDK
func resourceAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	log.Printf("[INFO] Deleting Lambda alias: %s:%s", d.Get("function_name"), d.Get("name"))

	params := &lambda.DeleteAliasInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
		Name:         aws.String(d.Get("name").(string)),
	}

	_, err := conn.DeleteAliasWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda alias: %s", err)
	}

	return diags
}

// resourceAliasUpdate maps to:
// UpdateAlias in the API / SDK
func resourceAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	log.Printf("[DEBUG] Updating Lambda alias: %s:%s", d.Get("function_name"), d.Get("name"))

	params := &lambda.UpdateAliasInput{
		Description:     aws.String(d.Get("description").(string)),
		FunctionName:    aws.String(d.Get("function_name").(string)),
		FunctionVersion: aws.String(d.Get("function_version").(string)),
		Name:            aws.String(d.Get("name").(string)),
		RoutingConfig:   expandAliasRoutingConfiguration(d.Get("routing_config").([]interface{})),
	}

	_, err := conn.UpdateAliasWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lambda alias: %s", err)
	}

	return diags
}

func expandAliasRoutingConfiguration(l []interface{}) *lambda.AliasRoutingConfiguration {
	aliasRoutingConfiguration := &lambda.AliasRoutingConfiguration{}

	if len(l) == 0 || l[0] == nil {
		return aliasRoutingConfiguration
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["additional_version_weights"]; ok {
		aliasRoutingConfiguration.AdditionalVersionWeights = expandFloat64Map(v.(map[string]interface{}))
	}

	return aliasRoutingConfiguration
}

func resourceAliasImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("Unexpected format of ID (%q), expected FUNCTION_NAME/ALIAS", d.Id())
	}

	functionName := idParts[0]
	alias := idParts[1]

	d.Set("function_name", functionName)
	d.Set("name", alias)
	return []*schema.ResourceData{d}, nil
}
