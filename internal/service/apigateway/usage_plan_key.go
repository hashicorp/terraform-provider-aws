// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_usage_plan_key", name="Usage Plan Key")
func resourceUsagePlanKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUsagePlanKeyCreate,
		ReadWithoutTimeout:   resourceUsagePlanKeyRead,
		DeleteWithoutTimeout: resourceUsagePlanKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected USAGE-PLAN-ID/USAGE-PLAN-KEY-ID", d.Id())
				}
				usagePlanId := idParts[0]
				usagePlanKeyId := idParts[1]
				d.Set("usage_plan_id", usagePlanId)
				d.Set(names.AttrKeyID, usagePlanKeyId)
				d.SetId(usagePlanKeyId)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrKeyID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"usage_plan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrValue: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUsagePlanKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.CreateUsagePlanKeyInput{
		KeyId:       aws.String(d.Get(names.AttrKeyID).(string)),
		KeyType:     aws.String(d.Get("key_type").(string)),
		UsagePlanId: aws.String(d.Get("usage_plan_id").(string)),
	}

	output, err := conn.CreateUsagePlanKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Usage Plan Key: %s", err)
	}

	d.SetId(aws.ToString(output.Id))

	return append(diags, resourceUsagePlanKeyRead(ctx, d, meta)...)
}

func resourceUsagePlanKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	upk, err := findUsagePlanKeyByTwoPartKey(ctx, conn, d.Get("usage_plan_id").(string), d.Get(names.AttrKeyID).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Usage Plan Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Usage Plan Key (%s): %s", d.Id(), err)
	}

	d.Set("key_type", upk.Type)
	d.Set(names.AttrName, upk.Name)
	d.Set(names.AttrValue, upk.Value)

	return diags
}

func resourceUsagePlanKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Usage Plan Key: %s", d.Id())
	_, err := conn.DeleteUsagePlanKey(ctx, &apigateway.DeleteUsagePlanKeyInput{
		KeyId:       aws.String(d.Get(names.AttrKeyID).(string)),
		UsagePlanId: aws.String(d.Get("usage_plan_id").(string)),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Usage Plan Key (%s): %s", d.Id(), err)
	}

	return diags
}

func findUsagePlanKeyByTwoPartKey(ctx context.Context, conn *apigateway.Client, usagePlanID, keyID string) (*apigateway.GetUsagePlanKeyOutput, error) {
	input := &apigateway.GetUsagePlanKeyInput{
		KeyId:       aws.String(keyID),
		UsagePlanId: aws.String(usagePlanID),
	}

	output, err := conn.GetUsagePlanKey(ctx, input)

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
