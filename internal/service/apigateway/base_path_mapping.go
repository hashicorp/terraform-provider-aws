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
			"domain_name_id": {
				Type:     schema.TypeString,
				Optional: true,
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
	var id string
	input := &apigateway.CreateBasePathMappingInput{
		RestApiId:  aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(domainName),
		BasePath:   aws.String(basePath),
		Stage:      aws.String(d.Get("stage_name").(string)),
	}

	if v, ok := d.GetOk("domain_name_id"); ok {
		domainNameId := v.(string)
		input.DomainNameId = aws.String(domainNameId)
		id = basePathMappingCreateResourceID(domainName, basePath, domainNameId)
	} else {
		id = basePathMappingCreateResourceID(domainName, basePath, "")
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

	domainName, basePath, domainNameId, err := basePathMappingParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	mapping, err := findBasePathMappingByThreePartKey(ctx, conn, domainName, basePath, domainNameId)

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
	d.Set("domain_name_id", domainNameId)

	return diags
}

func resourceBasePathMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName, basePath, domainNameId, err := basePathMappingParseResourceID(d.Id())
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

	if domainNameId != "" {
		input.DomainNameId = aws.String(domainNameId)
	}

	_, err = conn.UpdateBasePathMapping(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Base Path Mapping (%s): %s", d.Id(), err)
	}

	if d.HasChange("base_path") {
		id := basePathMappingCreateResourceID(d.Get(names.AttrDomainName).(string), d.Get("base_path").(string), domainNameId)
		d.SetId(id)
	}

	return append(diags, resourceBasePathMappingRead(ctx, d, meta)...)
}

func resourceBasePathMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName, basePath, domainNameId, err := basePathMappingParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting API Gateway Base Path Mapping: %s", d.Id())
	input := &apigateway.DeleteBasePathMappingInput{
		DomainName: aws.String(domainName),
		BasePath:   aws.String(basePath),
	}

	if domainNameId != "" {
		input.DomainNameId = aws.String(domainNameId)
	}

	_, err = conn.DeleteBasePathMapping(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Base Path Mapping (%s): %s", d.Id(), err)
	}

	return diags
}

func findBasePathMappingByThreePartKey(ctx context.Context, conn *apigateway.Client, domainName, basePath, domainNameId string) (*apigateway.GetBasePathMappingOutput, error) {
	input := &apigateway.GetBasePathMappingInput{
		BasePath:   aws.String(basePath),
		DomainName: aws.String(domainName),
	}

	if domainNameId != "" {
		input.DomainNameId = aws.String(domainNameId)
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

func basePathMappingCreateResourceID(domainName, basePath, domainNameId string) string {
	var id string
	parts := []string{domainName, basePath}

	if domainNameId != "" {
		parts = append(parts, domainNameId)
	}

	id = strings.Join(parts, basePathMappingResourceIDSeparator)

	return id
}

func basePathMappingParseResourceID(id string) (string, string, string, error) {
	err := fmt.Errorf("Unexpected format of ID (%[1]s), expected DOMAIN-NAME%[2]sBASEPATH or DOMAIN-NAME%[2]sBASEPATH%[2]sDOMAIN-NAME-ID", id, basePathMappingResourceIDSeparator)

	switch parts := strings.SplitN(id, basePathMappingResourceIDSeparator, 3); len(parts) {
	case 2:
		if domainName, basePath := parts[0], parts[1]; domainName != "" {
			if basePath == "" {
				basePath = emptyBasePathMappingValue
			}
			return domainName, basePath, "", nil
		}
	case 3:
		if domainName, basePath, domainNameID := parts[0], parts[1], parts[2]; domainName != "" && domainNameID != "" {
			if basePath == "" {
				basePath = emptyBasePathMappingValue
			}
			return domainName, basePath, domainNameID, nil
		}
	}

	return "", "", "", err
}
