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

// @SDKResource("aws_api_gateway_documentation_version", name="Documentation Version")
func resourceDocumentationVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDocumentationVersionCreate,
		ReadWithoutTimeout:   resourceDocumentationVersionRead,
		UpdateWithoutTimeout: resourceDocumentationVersionUpdate,
		DeleteWithoutTimeout: resourceDocumentationVersionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDocumentationVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID := d.Get("rest_api_id").(string)
	input := &apigateway.CreateDocumentationVersionInput{
		DocumentationVersion: aws.String(d.Get(names.AttrVersion).(string)),
		RestApiId:            aws.String(apiID),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateDocumentationVersion(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Documentation Version: %s", err)
	}

	d.SetId(documentationVersionCreateResourceID(apiID, aws.ToString(output.Version)))

	return append(diags, resourceDocumentationVersionRead(ctx, d, meta)...)
}

func resourceDocumentationVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID, documentationVersion, err := documentationVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	version, err := findDocumentationVersionByTwoPartKey(ctx, conn, apiID, documentationVersion)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Documentation Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Documentation Version (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDescription, version.Description)
	d.Set("rest_api_id", apiID)
	d.Set(names.AttrVersion, version.Version)

	return diags
}

func resourceDocumentationVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID, documentationVersion, err := documentationVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.UpdateDocumentationVersion(ctx, &apigateway.UpdateDocumentationVersionInput{
		DocumentationVersion: aws.String(documentationVersion),
		PatchOperations: []types.PatchOperation{
			{
				Op:    types.OpReplace,
				Path:  aws.String("/description"),
				Value: aws.String(d.Get(names.AttrDescription).(string)),
			},
		},
		RestApiId: aws.String(apiID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Documentation Version (%s): %s", d.Id(), err)
	}

	return append(diags, resourceDocumentationVersionRead(ctx, d, meta)...)
}

func resourceDocumentationVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID, documentationVersion, err := documentationVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting API Gateway Documentation Version: %s", d.Id())
	_, err = conn.DeleteDocumentationVersion(ctx, &apigateway.DeleteDocumentationVersionInput{
		DocumentationVersion: aws.String(documentationVersion),
		RestApiId:            aws.String(apiID),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Documentation Version (%s): %s", d.Id(), err)
	}

	return diags
}

func findDocumentationVersionByTwoPartKey(ctx context.Context, conn *apigateway.Client, apiID, documentationVersion string) (*apigateway.GetDocumentationVersionOutput, error) {
	input := &apigateway.GetDocumentationVersionInput{
		DocumentationVersion: aws.String(documentationVersion),
		RestApiId:            aws.String(apiID),
	}

	output, err := conn.GetDocumentationVersion(ctx, input)

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

const documentationVersionResourceIDSeparator = "/"

func documentationVersionCreateResourceID(apiID, documentationVersion string) string {
	parts := []string{apiID, documentationVersion}
	id := strings.Join(parts, documentationVersionResourceIDSeparator)

	return id
}

func documentationVersionParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, documentationVersionResourceIDSeparator)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%[1]s), expected REST-API-ID%[2]sVERSION", id, documentationVersionResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}
