// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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

const emptyBasePathMappingValue = "(none)"

// @SDKResource("aws_api_gateway_base_path_mapping", name="Base Path Mapping")
func resourceBasePathMapping() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBasePathMappingCreate,
		ReadWithoutTimeout:   resourceBasePathMappingRead,
		UpdateWithoutTimeout: resourceBasePathMappingUpdate,
		DeleteWithoutTimeout: resourceBasePathMappingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"base_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stage_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceBasePathMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName, basePath := d.Get(names.AttrDomainName).(string), d.Get("base_path").(string)
	id := basePathMappingCreateResourceID(domainName, basePath)
	input := &apigateway.CreateBasePathMappingInput{
		RestApiId:  aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(domainName),
		BasePath:   aws.String(basePath),
		Stage:      aws.String(d.Get("stage_name").(string)),
	}

	const (
		timeout = 30 * time.Second
	)
	_, err := tfresource.RetryWhenIsA[*types.BadRequestException](ctx, timeout, func() (interface{}, error) {
		return conn.CreateBasePathMapping(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Base Path Mapping (%s): %s", err, id)
	}

	d.SetId(id)

	return append(diags, resourceBasePathMappingRead(ctx, d, meta)...)
}

func resourceBasePathMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName, basePath, err := basePathMappingParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	mapping, err := findBasePathMappingByTwoPartKey(ctx, conn, domainName, basePath)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Base Path Mapping (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Base Path Mapping (%s): %s", d.Id(), err)
	}

	mappingBasePath := aws.ToString(mapping.BasePath)
	if mappingBasePath == emptyBasePathMappingValue {
		mappingBasePath = ""
	}

	d.Set("api_id", mapping.RestApiId)
	d.Set("base_path", mappingBasePath)
	d.Set(names.AttrDomainName, domainName)
	d.Set("stage_name", mapping.Stage)

	return diags
}

func resourceBasePathMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName, basePath, err := basePathMappingParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	operations := make([]types.PatchOperation, 0)

	if d.HasChange("stage_name") {
		operations = append(operations, types.PatchOperation{
			Op:    types.Op("replace"),
			Path:  aws.String("/stage"),
			Value: aws.String(d.Get("stage_name").(string)),
		})
	}

	if d.HasChange("api_id") {
		operations = append(operations, types.PatchOperation{
			Op:    types.Op("replace"),
			Path:  aws.String("/restapiId"),
			Value: aws.String(d.Get("api_id").(string)),
		})
	}

	if d.HasChange("base_path") {
		operations = append(operations, types.PatchOperation{
			Op:    types.Op("replace"),
			Path:  aws.String("/basePath"),
			Value: aws.String(d.Get("base_path").(string)),
		})
	}

	input := apigateway.UpdateBasePathMappingInput{
		BasePath:        aws.String(basePath),
		DomainName:      aws.String(domainName),
		PatchOperations: operations,
	}

	_, err = conn.UpdateBasePathMapping(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Base Path Mapping (%s): %s", d.Id(), err)
	}

	if d.HasChange("base_path") {
		id := basePathMappingCreateResourceID(d.Get(names.AttrDomainName).(string), d.Get("base_path").(string))
		d.SetId(id)
	}

	return append(diags, resourceBasePathMappingRead(ctx, d, meta)...)
}

func resourceBasePathMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName, basePath, err := basePathMappingParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting API Gateway Base Path Mapping: %s", d.Id())
	_, err = conn.DeleteBasePathMapping(ctx, &apigateway.DeleteBasePathMappingInput{
		DomainName: aws.String(domainName),
		BasePath:   aws.String(basePath),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Base Path Mapping (%s): %s", d.Id(), err)
	}

	return diags
}

func findBasePathMappingByTwoPartKey(ctx context.Context, conn *apigateway.Client, domainName, basePath string) (*apigateway.GetBasePathMappingOutput, error) {
	input := &apigateway.GetBasePathMappingInput{
		BasePath:   aws.String(basePath),
		DomainName: aws.String(domainName),
	}

	output, err := conn.GetBasePathMapping(ctx, input)

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

const basePathMappingResourceIDSeparator = "/"

func basePathMappingCreateResourceID(domainName, basePath string) string {
	parts := []string{domainName, basePath}
	id := strings.Join(parts, basePathMappingResourceIDSeparator)

	return id
}

func basePathMappingParseResourceID(id string) (string, string, error) {
	err := fmt.Errorf("Unexpected format of ID (%[1]s), expected DOMAIN%[2]sBASEPATH", id, basePathMappingResourceIDSeparator)

	parts := strings.SplitN(id, basePathMappingResourceIDSeparator, 2)
	if len(parts) != 2 {
		return "", "", err
	}

	domainName := parts[0]
	basePath := parts[1]

	if domainName == "" {
		return "", "", err
	}

	if basePath == "" {
		basePath = emptyBasePathMappingValue
	}

	return domainName, basePath, nil
}
