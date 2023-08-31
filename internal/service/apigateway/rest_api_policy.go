// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_api_gateway_rest_api_policy")
func ResourceRestAPIPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRestAPIPolicyPut,
		ReadWithoutTimeout:   resourceRestAPIPolicyRead,
		UpdateWithoutTimeout: resourceRestAPIPolicyPut,
		DeleteWithoutTimeout: resourceRestAPIPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"policy": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceRestAPIPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	restApiId := d.Get("rest_api_id").(string)
	log.Printf("[DEBUG] Setting API Gateway REST API Policy: %s", restApiId)

	operations := make([]*apigateway.PatchOperation, 0)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	operations = append(operations, &apigateway.PatchOperation{
		Op:    aws.String(apigateway.OpReplace),
		Path:  aws.String("/policy"),
		Value: aws.String(policy),
	})

	res, err := conn.UpdateRestApiWithContext(ctx, &apigateway.UpdateRestApiInput{
		RestApiId:       aws.String(restApiId),
		PatchOperations: operations,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting API Gateway REST API Policy %s", err)
	}

	log.Printf("[DEBUG] API Gateway REST API Policy Set: %s", restApiId)

	d.SetId(aws.StringValue(res.Id))

	return append(diags, resourceRestAPIPolicyRead(ctx, d, meta)...)
}

func resourceRestAPIPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	log.Printf("[DEBUG] Reading API Gateway REST API Policy %s", d.Id())

	api, err := conn.GetRestApiWithContext(ctx, &apigateway.GetRestApiInput{
		RestApiId: aws.String(d.Id()),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		log.Printf("[WARN] API Gateway REST API Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway REST API Policy (%s): %s", d.Id(), err)
	}

	normalizedPolicy, err := structure.NormalizeJsonString(`"` + aws.StringValue(api.Policy) + `"`)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "normalizing API Gateway REST API policy JSON: %s", err)
	}

	policy, err := strconv.Unquote(normalizedPolicy)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "unescaping API Gateway REST API policy: %s", err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), policy)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", policyToSet, err)
	}

	d.Set("policy", policyToSet)

	d.Set("rest_api_id", api.Id)

	return diags
}

func resourceRestAPIPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	restApiId := d.Get("rest_api_id").(string)
	log.Printf("[DEBUG] Deleting API Gateway REST API Policy: %s", restApiId)

	operations := make([]*apigateway.PatchOperation, 0)

	operations = append(operations, &apigateway.PatchOperation{
		Op:    aws.String(apigateway.OpReplace),
		Path:  aws.String("/policy"),
		Value: aws.String(""),
	})

	_, err := conn.UpdateRestApiWithContext(ctx, &apigateway.UpdateRestApiInput{
		RestApiId:       aws.String(restApiId),
		PatchOperations: operations,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway REST API policy: %s", err)
	}

	log.Printf("[DEBUG] API Gateway REST API Policy Deleted: %s", restApiId)

	return diags
}
