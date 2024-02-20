// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_api_gateway_documentation_version")
func ResourceDocumentationVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDocumentationVersionCreate,
		ReadWithoutTimeout:   resourceDocumentationVersionRead,
		UpdateWithoutTimeout: resourceDocumentationVersionUpdate,
		DeleteWithoutTimeout: resourceDocumentationVersionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceDocumentationVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	restApiId := d.Get("rest_api_id").(string)

	params := &apigateway.CreateDocumentationVersionInput{
		DocumentationVersion: aws.String(d.Get("version").(string)),
		RestApiId:            aws.String(restApiId),
	}
	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway Documentation Version: %s", params)

	version, err := conn.CreateDocumentationVersionWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Documentation Version: %s", err)
	}

	d.SetId(restApiId + "/" + *version.Version)

	return append(diags, resourceDocumentationVersionRead(ctx, d, meta)...)
}

func resourceDocumentationVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)
	log.Printf("[DEBUG] Reading API Gateway Documentation Version %s", d.Id())

	apiId, docVersion, err := DecodeDocumentationVersionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Documentation Version (%s): %s", d.Id(), err)
	}

	version, err := conn.GetDocumentationVersionWithContext(ctx, &apigateway.GetDocumentationVersionInput{
		DocumentationVersion: aws.String(docVersion),
		RestApiId:            aws.String(apiId),
	})
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			log.Printf("[WARN] API Gateway Documentation Version (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Documentation Version (%s): %s", d.Id(), err)
	}

	d.Set("rest_api_id", apiId)
	d.Set("description", version.Description)
	d.Set("version", version.Version)

	return diags
}

func resourceDocumentationVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)
	log.Printf("[DEBUG] Updating API Gateway Documentation Version %s", d.Id())

	_, err := conn.UpdateDocumentationVersionWithContext(ctx, &apigateway.UpdateDocumentationVersionInput{
		DocumentationVersion: aws.String(d.Get("version").(string)),
		RestApiId:            aws.String(d.Get("rest_api_id").(string)),
		PatchOperations: []*apigateway.PatchOperation{
			{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/description"),
				Value: aws.String(d.Get("description").(string)),
			},
		},
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Documentation Version (%s): %s", d.Id(), err)
	}
	log.Printf("[DEBUG] Updated API Gateway Documentation Version %s", d.Id())

	return append(diags, resourceDocumentationVersionRead(ctx, d, meta)...)
}

func resourceDocumentationVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)
	log.Printf("[DEBUG] Deleting API Gateway Documentation Version: %s", d.Id())

	_, err := conn.DeleteDocumentationVersionWithContext(ctx, &apigateway.DeleteDocumentationVersionInput{
		DocumentationVersion: aws.String(d.Get("version").(string)),
		RestApiId:            aws.String(d.Get("rest_api_id").(string)),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Documentation Version (%s): %s", d.Id(), err)
	}
	return diags
}

func DecodeDocumentationVersionID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Expected ID in the form of REST-API-ID/VERSION, given: %q", id)
	}
	return parts[0], parts[1], nil
}
