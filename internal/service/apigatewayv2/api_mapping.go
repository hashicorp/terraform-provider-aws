// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_apigatewayv2_api_mapping")
func ResourceAPIMapping() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPIMappingCreate,
		ReadWithoutTimeout:   resourceAPIMappingRead,
		UpdateWithoutTimeout: resourceAPIMappingUpdate,
		DeleteWithoutTimeout: resourceAPIMappingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAPIMappingImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"api_mapping_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stage": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAPIMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	input := &apigatewayv2.CreateApiMappingInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(d.Get("domain_name").(string)),
		Stage:      aws.String(d.Get("stage").(string)),
	}

	if v, ok := d.GetOk("api_mapping_key"); ok {
		input.ApiMappingKey = aws.String(v.(string))
	}

	output, err := conn.CreateApiMappingWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 API Mapping: %s", err)
	}

	d.SetId(aws.StringValue(output.ApiMappingId))

	return append(diags, resourceAPIMappingRead(ctx, d, meta)...)
}

func resourceAPIMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	output, err := FindAPIMappingByTwoPartKey(ctx, conn, d.Id(), d.Get("domain_name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 API mapping (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 API Mapping (%s): %s", d.Id(), err)
	}

	d.Set("api_id", output.ApiId)
	d.Set("api_mapping_key", output.ApiMappingKey)
	d.Set("stage", output.Stage)

	return diags
}

func resourceAPIMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	input := &apigatewayv2.UpdateApiMappingInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		ApiMappingId: aws.String(d.Id()),
		DomainName:   aws.String(d.Get("domain_name").(string)),
	}

	if d.HasChange("api_mapping_key") {
		input.ApiMappingKey = aws.String(d.Get("api_mapping_key").(string))
	}

	if d.HasChange("stage") {
		input.Stage = aws.String(d.Get("stage").(string))
	}

	_, err := conn.UpdateApiMappingWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 API Mapping (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAPIMappingRead(ctx, d, meta)...)
}

func resourceAPIMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 API Mapping (%s)", d.Id())
	_, err := conn.DeleteApiMappingWithContext(ctx, &apigatewayv2.DeleteApiMappingInput{
		ApiMappingId: aws.String(d.Id()),
		DomainName:   aws.String(d.Get("domain_name").(string)),
	})

	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 API Mapping (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAPIMappingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'api-mapping-id/domain-name'", d.Id())
	}

	d.SetId(parts[0])
	d.Set("domain_name", parts[1])

	return []*schema.ResourceData{d}, nil
}

func FindAPIMappingByTwoPartKey(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, id, domainName string) (*apigatewayv2.GetApiMappingOutput, error) {
	input := &apigatewayv2.GetApiMappingInput{
		ApiMappingId: aws.String(id),
		DomainName:   aws.String(domainName),
	}

	return findAPIMapping(ctx, conn, input)
}

func findAPIMapping(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, input *apigatewayv2.GetApiMappingInput) (*apigatewayv2.GetApiMappingOutput, error) {
	output, err := conn.GetApiMappingWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
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
