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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// @SDKResource("aws_apigatewayv2_model", name="Model")
func resourceModel() *schema.Resource {
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
			names.AttrContentType: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "must be alphanumeric"),
				),
			},
			names.AttrSchema: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 32768),
					validation.StringIsJSON,
				),
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
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

	name := d.Get(names.AttrName).(string)
	input := &apigatewayv2.CreateModelInput{
		ApiId:       aws.String(d.Get("api_id").(string)),
		ContentType: aws.String(d.Get(names.AttrContentType).(string)),
		Name:        aws.String(name),
		Schema:      aws.String(d.Get(names.AttrSchema).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateModel(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 Model (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ModelId))

	return append(diags, resourceModelRead(ctx, d, meta)...)
}

func resourceModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	output, err := findModelByTwoPartKey(ctx, conn, d.Get("api_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 Model (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 Model (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrContentType, output.ContentType)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrSchema, output.Schema)

	return diags
}

func resourceModelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	input := &apigatewayv2.UpdateModelInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		ModelId: aws.String(d.Id()),
	}

	if d.HasChange(names.AttrContentType) {
		input.ContentType = aws.String(d.Get(names.AttrContentType).(string))
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange(names.AttrName) {
		input.Name = aws.String(d.Get(names.AttrName).(string))
	}

	if d.HasChange(names.AttrSchema) {
		input.Schema = aws.String(d.Get(names.AttrSchema).(string))
	}

	_, err := conn.UpdateModel(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 Model (%s): %s", d.Id(), err)
	}

	return append(diags, resourceModelRead(ctx, d, meta)...)
}

func resourceModelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 Model: %s", d.Id())
	_, err := conn.DeleteModel(ctx, &apigatewayv2.DeleteModelInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		ModelId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 Model (%s): %s", d.Id(), err)
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

func findModelByTwoPartKey(ctx context.Context, conn *apigatewayv2.Client, apiID, modelID string) (*apigatewayv2.GetModelOutput, error) {
	input := &apigatewayv2.GetModelInput{
		ApiId:   aws.String(apiID),
		ModelId: aws.String(modelID),
	}

	return findModel(ctx, conn, input)
}

func findModel(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetModelInput) (*apigatewayv2.GetModelOutput, error) {
	output, err := conn.GetModel(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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
