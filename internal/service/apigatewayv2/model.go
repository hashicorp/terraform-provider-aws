// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_apigatewayv2_model")
func ResourceModel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceModelCreate,
		ReadWithoutTimeout:   resourceModelRead,
		UpdateWithoutTimeout: resourceModelUpdate,
		DeleteWithoutTimeout: resourceModelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceModelImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "must be alphanumeric"),
				),
			},
			"schema": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 32768),
					validation.StringIsJSON,
				),
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceModelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	req := &apigatewayv2.CreateModelInput{
		ApiId:       aws.String(d.Get("api_id").(string)),
		ContentType: aws.String(d.Get("content_type").(string)),
		Name:        aws.String(d.Get("name").(string)),
		Schema:      aws.String(d.Get("schema").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 model: %+v", req)
	resp, err := conn.CreateModel(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 model: %s", err)
	}

	d.SetId(aws.ToString(resp.ModelId))

	return append(diags, resourceModelRead(ctx, d, meta)...)
}

func resourceModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	resp, err := conn.GetModel(ctx, &apigatewayv2.GetModelInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		ModelId: aws.String(d.Id()),
	})
	if errs.IsA[*awstypes.NotFoundException](err) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 model (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 model: %s", err)
	}

	d.Set("content_type", resp.ContentType)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("schema", resp.Schema)

	return diags
}

func resourceModelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	req := &apigatewayv2.UpdateModelInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		ModelId: aws.String(d.Id()),
	}
	if d.HasChange("content_type") {
		req.ContentType = aws.String(d.Get("content_type").(string))
	}
	if d.HasChange("description") {
		req.Description = aws.String(d.Get("description").(string))
	}
	if d.HasChange("name") {
		req.Name = aws.String(d.Get("name").(string))
	}
	if d.HasChange("schema") {
		req.Schema = aws.String(d.Get("schema").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 model: %+v", req)
	_, err := conn.UpdateModel(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 model: %s", err)
	}

	return append(diags, resourceModelRead(ctx, d, meta)...)
}

func resourceModelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 model (%s)", d.Id())
	_, err := conn.DeleteModel(ctx, &apigatewayv2.DeleteModelInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		ModelId: aws.String(d.Id()),
	})
	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 model: %s", err)
	}

	return diags
}

func resourceModelImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'api-id/model-id'", d.Id())
	}

	d.SetId(parts[1])
	d.Set("api_id", parts[0])

	return []*schema.ResourceData{d}, nil
}
