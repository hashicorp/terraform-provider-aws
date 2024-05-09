// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_rest_api_policy", name="REST API Policy")
func resourceRestAPIPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRestAPIPolicyPut,
		ReadWithoutTimeout:   resourceRestAPIPolicyRead,
		UpdateWithoutTimeout: resourceRestAPIPolicyPut,
		DeleteWithoutTimeout: resourceRestAPIPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: {
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
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRestAPIPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID := d.Get("rest_api_id").(string)
	operations := make([]types.PatchOperation, 0)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	operations = append(operations, types.PatchOperation{
		Op:    types.OpReplace,
		Path:  aws.String("/policy"),
		Value: aws.String(policy),
	})
	input := &apigateway.UpdateRestApiInput{
		PatchOperations: operations,
		RestApiId:       aws.String(apiID),
	}

	output, err := conn.UpdateRestApi(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway REST API Policy (%s): %s", apiID, err)
	}

	if d.IsNewResource() {
		d.SetId(aws.ToString(output.Id))
	}

	return append(diags, resourceRestAPIPolicyRead(ctx, d, meta)...)
}

func resourceRestAPIPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	api, err := findRestAPIByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway REST API (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway REST API (%s): %s", d.Id(), err)
	}

	policy, err := flattenAPIPolicy(api.Policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policy)
	d.Set("rest_api_id", api.Id)

	return diags
}

func resourceRestAPIPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	operations := make([]types.PatchOperation, 0)
	operations = append(operations, types.PatchOperation{
		Op:    types.OpReplace,
		Path:  aws.String("/policy"),
		Value: aws.String(""),
	})

	log.Printf("[DEBUG] Deleting API Gateway REST API Policy: %s", d.Id())
	_, err := conn.UpdateRestApi(ctx, &apigateway.UpdateRestApiInput{
		PatchOperations: operations,
		RestApiId:       aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway REST API Policy (%s): %s", d.Id(), err)
	}

	return diags
}
