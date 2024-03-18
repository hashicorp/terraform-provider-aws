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
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const EmptyBasePathMappingValue = "(none)"

// @SDKResource("aws_api_gateway_base_path_mapping")
func ResourceBasePathMapping() *schema.Resource {
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
			"stage_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBasePathMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)
	input := &apigateway.CreateBasePathMappingInput{
		RestApiId:  aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(d.Get("domain_name").(string)),
		BasePath:   aws.String(d.Get("base_path").(string)),
		Stage:      aws.String(d.Get("stage_name").(string)),
	}

	err := retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
		_, err := conn.CreateBasePathMapping(ctx, input)

		if err != nil {
			if errs.IsA[*awstypes.BadRequestException](err) {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateBasePathMapping(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Base Path Mapping: %s", err)
	}

	id := fmt.Sprintf("%s/%s", d.Get("domain_name").(string), d.Get("base_path").(string))
	d.SetId(id)

	return append(diags, resourceBasePathMappingRead(ctx, d, meta)...)
}

func resourceBasePathMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	operations := make([]awstypes.PatchOperation, 0)

	if d.HasChange("stage_name") {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.Op("replace"),
			Path:  aws.String("/stage"),
			Value: aws.String(d.Get("stage_name").(string)),
		})
	}

	if d.HasChange("api_id") {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.Op("replace"),
			Path:  aws.String("/restapiId"),
			Value: aws.String(d.Get("api_id").(string)),
		})
	}

	if d.HasChange("base_path") {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.Op("replace"),
			Path:  aws.String("/basePath"),
			Value: aws.String(d.Get("base_path").(string)),
		})
	}

	domainName, basePath, err := DecodeBasePathMappingID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Base Path Mapping (%s): %s", d.Id(), err)
	}

	input := apigateway.UpdateBasePathMappingInput{
		BasePath:        aws.String(basePath),
		DomainName:      aws.String(domainName),
		PatchOperations: operations,
	}

	log.Printf("[INFO] Updating API Gateway Base Path Mapping: %+v", input)

	_, err = conn.UpdateBasePathMapping(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Base Path Mapping (%s): %s", d.Id(), err)
	}

	if d.HasChange("base_path") {
		id := fmt.Sprintf("%s/%s", d.Get("domain_name").(string), d.Get("base_path").(string))
		d.SetId(id)
	}

	log.Printf("[DEBUG] API Gateway Base Path Mapping updated: %s", d.Id())

	return append(diags, resourceBasePathMappingRead(ctx, d, meta)...)
}

func resourceBasePathMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName, basePath, err := DecodeBasePathMappingID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Base Path Mapping (%s): %s", d.Id(), err)
	}

	mapping, err := conn.GetBasePathMapping(ctx, &apigateway.GetBasePathMappingInput{
		DomainName: aws.String(domainName),
		BasePath:   aws.String(basePath),
	})
	if err != nil {
		if !d.IsNewResource() && errs.IsA[*awstypes.NotFoundException](err) {
			log.Printf("[WARN] API Gateway Base Path Mapping (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading API Gateway Base Path Mapping (%s): %s", d.Id(), err)
	}

	mappingBasePath := aws.ToString(mapping.BasePath)

	if mappingBasePath == EmptyBasePathMappingValue {
		mappingBasePath = ""
	}

	d.Set("base_path", mappingBasePath)
	d.Set("domain_name", domainName)
	d.Set("api_id", mapping.RestApiId)
	d.Set("stage_name", mapping.Stage)

	return diags
}

func resourceBasePathMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName, basePath, err := DecodeBasePathMappingID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Base Path Mapping (%s): %s", d.Id(), err)
	}

	_, err = conn.DeleteBasePathMapping(ctx, &apigateway.DeleteBasePathMappingInput{
		DomainName: aws.String(domainName),
		BasePath:   aws.String(basePath),
	})

	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Base Path Mapping (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeBasePathMappingID(id string) (string, string, error) {
	idFormatErr := fmt.Errorf("Unexpected format of ID (%q), expected DOMAIN/BASEPATH", id)

	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return "", "", idFormatErr
	}

	domainName := parts[0]
	basePath := parts[1]

	if domainName == "" {
		return "", "", idFormatErr
	}

	if basePath == "" {
		basePath = EmptyBasePathMappingValue
	}

	return domainName, basePath, nil
}
